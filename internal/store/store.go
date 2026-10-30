package store

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/pwenxiang51-wq/polaris/internal/shared"
)

var nodeIDPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,47}$`)

const historyRetention = 31 * 24 * time.Hour
const failedGeoRetryAfter = 24 * time.Hour
const rawHistoryWindow = time.Hour
const minuteHistoryWindow = 24 * time.Hour
const historyCompactInterval = 15 * time.Minute
const maxHistoryFileBytes = 32 * 1024 * 1024

type ProbeToken struct {
	NodeID    string    `json:"node_id"`
	NodeName  string    `json:"node_name"`
	Hash      string    `json:"hash"`
	CreatedAt time.Time `json:"created_at"`
}

type MetadataUpdate struct {
	ExpiryDate          string
	PriceUSD            *float64
	MonthlyQuotaBytes   uint64
	TotalQuotaBytes     uint64
	QuotaMode           string
	BandwidthMbps       int
	PublicIPv4Override  string
	PublicIPv6Override  string
	CountryOverride     string
	CountryCodeOverride string
	RegionOverride      string
	CityOverride        string
	ISPOverride         string
	ASNOverride         string
}

type Store struct {
	mu          sync.RWMutex
	dir         string
	historyDir  string
	nodes       map[string]shared.Report
	tokens      map[string]ProbeToken
	history     map[string][]shared.HistoryPoint
	lastCompact map[string]time.Time
	ipMetadata  map[string]shared.IPMetadata
}

func Open(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	if err := os.Chmod(dir, 0700); err != nil {
		return nil, err
	}
	historyDir := filepath.Join(dir, "history")
	if err := os.MkdirAll(historyDir, 0700); err != nil {
		return nil, err
	}
	if err := os.Chmod(historyDir, 0700); err != nil {
		return nil, err
	}
	s := &Store{
		dir: dir, historyDir: historyDir,
		nodes: make(map[string]shared.Report), tokens: make(map[string]ProbeToken),
		history: make(map[string][]shared.HistoryPoint), lastCompact: make(map[string]time.Time),
		ipMetadata: make(map[string]shared.IPMetadata),
	}
	if err := s.load(filepath.Join(dir, "nodes.json"), &s.nodes); err != nil {
		return nil, fmt.Errorf("load node state: %w", err)
	}
	if err := s.load(filepath.Join(dir, "tokens.json"), &s.tokens); err != nil {
		return nil, fmt.Errorf("load token state: %w", err)
	}
	if err := s.load(filepath.Join(dir, "ip-metadata.json"), &s.ipMetadata); err != nil {
		return nil, fmt.Errorf("load IP metadata cache: %w", err)
	}
	if err := s.cleanupOrphanHistoryFiles(); err != nil {
		return nil, fmt.Errorf("clean orphan history files: %w", err)
	}
	migrated := false
	for nodeID, node := range s.nodes {
		points, truncated, err := s.readHistoryFile(nodeID)
		if err != nil {
			return nil, fmt.Errorf("load history for %s: %w", nodeID, err)
		}
		if len(node.History) > 0 {
			points = mergeHistory(points, node.History)
			migrated = true
		}
		points = compactHistory(points, time.Now().UTC())
		if truncated {
			log.Printf("history file for node %s exceeded %d MiB; oldest records were discarded and history was compacted to %d points", nodeID, maxHistoryFileBytes/(1024*1024), len(points))
		}
		s.history[nodeID] = points
		s.lastCompact[nodeID] = time.Now().UTC()
		if err := s.writeHistoryFile(nodeID, points); err != nil {
			return nil, fmt.Errorf("persist history for %s: %w", nodeID, err)
		}
		node.History = nil
		if node.UsageMonth == "" {
			node.UsageMonth = time.Now().UTC().Format("2006-01")
		}
		s.nodes[nodeID] = node
	}
	if migrated {
		if err := s.saveNodesLocked(); err != nil {
			return nil, fmt.Errorf("migrate legacy node history: %w", err)
		}
	}
	return s, nil
}

func (s *Store) load(path string, target any) error {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(data) > 16*1024*1024 {
		return errors.New("state file exceeds safety limit")
	}
	return json.Unmarshal(data, target)
}

func (s *Store) saveLocked(name string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return writeAtomic(filepath.Join(s.dir, name), data, 0600)
}

func (s *Store) saveNodesLocked() error {
	state := make(map[string]shared.Report, len(s.nodes))
	for id, node := range s.nodes {
		node.History = nil
		state[id] = node
	}
	return s.saveLocked("nodes.json", state)
}

func writeAtomic(path string, data []byte, mode os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".polaris-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(mode); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	if dirFile, err := os.Open(dir); err == nil {
		_ = dirFile.Sync()
		_ = dirFile.Close()
	}
	return nil
}

func (s *Store) historyPath(nodeID string) string {
	return filepath.Join(s.historyDir, nodeID+".jsonl")
}

func (s *Store) cleanupOrphanHistoryFiles() error {
	entries, err := os.ReadDir(s.historyDir)
	if err != nil {
		return err
	}
	removed := 0
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".jsonl") {
			continue
		}
		nodeID := strings.TrimSuffix(name, ".jsonl")
		if !nodeIDPattern.MatchString(nodeID) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			continue
		}
		if _, exists := s.nodes[nodeID]; exists {
			continue
		}
		if err := os.Remove(filepath.Join(s.historyDir, name)); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		removed++
	}
	if removed > 0 {
		log.Printf("removed %d orphan Polaris history file(s)", removed)
	}
	return nil
}

func (s *Store) readHistoryFile(nodeID string) ([]shared.HistoryPoint, bool, error) {
	if !nodeIDPattern.MatchString(nodeID) {
		return nil, false, errors.New("invalid node identifier in state")
	}
	file, err := os.Open(s.historyPath(nodeID))
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, false, err
	}
	truncated := info.Size() > maxHistoryFileBytes
	var reader io.Reader = file
	if truncated {
		start := info.Size() - maxHistoryFileBytes
		if start > 0 {
			previous := make([]byte, 1)
			if _, err := file.ReadAt(previous, start-1); err != nil {
				return nil, true, err
			}
			if _, err := file.Seek(start, io.SeekStart); err != nil {
				return nil, true, err
			}
			buffered := bufio.NewReaderSize(file, 64*1024)
			if previous[0] != '\n' {
				for {
					_, readErr := buffered.ReadSlice('\n')
					if readErr == nil {
						break
					}
					if errors.Is(readErr, bufio.ErrBufferFull) {
						continue
					}
					if errors.Is(readErr, io.EOF) {
						return nil, true, nil
					}
					return nil, true, readErr
				}
			}
			reader = buffered
		}
	}
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	var points []shared.HistoryPoint
	for scanner.Scan() {
		var point shared.HistoryPoint
		if err := json.Unmarshal(scanner.Bytes(), &point); err != nil {
			return nil, truncated, fmt.Errorf("invalid history record: %w", err)
		}
		if !point.At.IsZero() {
			if point.Samples < 1 {
				point.Samples = 1
			}
			points = append(points, point)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, truncated, err
	}
	sort.Slice(points, func(i, j int) bool { return points[i].At.Before(points[j].At) })
	return dedupeHistory(points), truncated, nil
}

func (s *Store) writeHistoryFile(nodeID string, points []shared.HistoryPoint) error {
	if !nodeIDPattern.MatchString(nodeID) {
		return errors.New("invalid node identifier")
	}
	var data []byte
	for _, point := range points {
		line, err := json.Marshal(point)
		if err != nil {
			return err
		}
		data = append(data, line...)
		data = append(data, '\n')
	}
	return writeAtomic(s.historyPath(nodeID), data, 0600)
}

func (s *Store) appendHistoryLocked(nodeID string, point shared.HistoryPoint, now time.Time) error {
	if point.Samples < 1 {
		point.Samples = 1
	}
	line, err := json.Marshal(point)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(s.historyPath(nodeID), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	if err := file.Chmod(0600); err != nil {
		_ = file.Close()
		return err
	}
	if _, err := file.Write(append(line, '\n')); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	s.history[nodeID] = append(s.history[nodeID], point)
	if last := s.lastCompact[nodeID]; last.IsZero() || now.Sub(last) >= historyCompactInterval || len(s.history[nodeID]) > 3000 {
		points := compactHistory(s.history[nodeID], now)
		if err := s.writeHistoryFile(nodeID, points); err != nil {
			return err
		}
		s.history[nodeID] = points
		s.lastCompact[nodeID] = now
	}
	return nil
}

func mergeHistory(existing, legacy []shared.HistoryPoint) []shared.HistoryPoint {
	combined := append(append([]shared.HistoryPoint(nil), existing...), legacy...)
	sort.Slice(combined, func(i, j int) bool { return combined[i].At.Before(combined[j].At) })
	return dedupeHistory(combined)
}

func dedupeHistory(points []shared.HistoryPoint) []shared.HistoryPoint {
	out := make([]shared.HistoryPoint, 0, len(points))
	for _, point := range points {
		if len(out) > 0 && out[len(out)-1].At.Equal(point.At) {
			out[len(out)-1] = point
			continue
		}
		out = append(out, point)
	}
	return out
}

func compactHistory(points []shared.HistoryPoint, now time.Time) []shared.HistoryPoint {
	cutoff := now.Add(-historyRetention).Truncate(time.Hour)
	filtered := make([]shared.HistoryPoint, 0, len(points))
	for _, point := range points {
		if point.At.IsZero() || point.At.Before(cutoff) {
			continue
		}
		if point.Samples < 1 {
			point.Samples = 1
		}
		filtered = append(filtered, point)
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].At.Before(filtered[j].At) })
	buckets := make(map[int64]*historyBucket)
	keys := make([]int64, 0, len(filtered))
	for _, point := range filtered {
		age := now.Sub(point.At)
		if age <= rawHistoryWindow {
			key := -point.At.UnixNano()
			bucket := &historyBucket{point: point, weight: point.Samples, at: point.At, raw: true}
			buckets[key] = bucket
			keys = append(keys, key)
			continue
		}
		bucketSize := time.Minute
		if age > minuteHistoryWindow {
			bucketSize = time.Hour
		}
		bucketAt := point.At.Truncate(bucketSize)
		key := bucketAt.UnixNano()
		bucket, exists := buckets[key]
		if !exists {
			bucket = &historyBucket{at: bucketAt}
			buckets[key] = bucket
			keys = append(keys, key)
		}
		bucket.add(point)
	}
	out := make([]shared.HistoryPoint, 0, len(buckets))
	for _, key := range keys {
		bucket := buckets[key]
		if bucket.raw {
			out = append(out, bucket.point)
		} else {
			out = append(out, bucket.finish())
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].At.Before(out[j].At) })
	return dedupeHistory(out)
}

type historyBucket struct {
	at               time.Time
	weight           int
	cpu, mem, rx, tx float64
	latencies        map[string]*latencyBucket
	point            shared.HistoryPoint
	raw              bool
}

type latencyBucket struct {
	name, target    string
	weight          int
	latencyWeighted float64
	latencyWeight   int
	lossWeighted    float64
}

func (b *historyBucket) add(point shared.HistoryPoint) {
	weight := point.Samples
	if weight < 1 {
		weight = 1
	}
	b.cpu += point.CPUPercent * float64(weight)
	b.mem += point.MemPercent * float64(weight)
	b.rx += point.RXRate * float64(weight)
	b.tx += point.TXRate * float64(weight)
	b.weight += weight
	if b.latencies == nil {
		b.latencies = make(map[string]*latencyBucket)
	}
	for _, result := range point.Latencies {
		key := result.Name + "\x00" + result.Target
		item := b.latencies[key]
		if item == nil {
			item = &latencyBucket{name: result.Name, target: result.Target}
			b.latencies[key] = item
		}
		item.lossWeighted += result.LossPercent * float64(weight)
		item.weight += weight
		if result.Reachable || result.LossPercent < 100 {
			item.latencyWeighted += result.LatencyMS * float64(weight)
			item.latencyWeight += weight
		}
	}
}

func (b *historyBucket) finish() shared.HistoryPoint {
	if b.weight < 1 {
		b.weight = 1
	}
	point := shared.HistoryPoint{At: b.at, Samples: b.weight, CPUPercent: b.cpu / float64(b.weight), MemPercent: b.mem / float64(b.weight), RXRate: b.rx / float64(b.weight), TXRate: b.tx / float64(b.weight)}
	for _, item := range b.latencies {
		latency := 0.0
		reachable := item.latencyWeight > 0
		if reachable {
			latency = item.latencyWeighted / float64(item.latencyWeight)
		}
		loss := item.lossWeighted / float64(item.weight)
		point.Latencies = append(point.Latencies, shared.ProbeResult{Name: item.name, Target: item.target, LatencyMS: latency, LossPercent: loss, Reachable: reachable})
	}
	sort.Slice(point.Latencies, func(i, j int) bool { return point.Latencies[i].Name < point.Latencies[j].Name })
	return point
}

func (s *Store) LookupToken(raw string) (ProbeToken, bool) {
	digest := sha256.Sum256([]byte(raw))
	hash := hex.EncodeToString(digest[:])
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tokens[hash]
	return t, ok
}

func (s *Store) CreateToken(nodeID, nodeName string) (string, error) {
	nodeID = strings.TrimSpace(nodeID)
	nodeName = strings.TrimSpace(nodeName)
	if !nodeIDPattern.MatchString(nodeID) {
		return "", errors.New("node ID must be 1-48 characters using letters, numbers, dots, underscores, or hyphens")
	}
	if utf8.RuneCountInString(nodeName) < 1 || utf8.RuneCountInString(nodeName) > 80 || len(nodeName) > 320 {
		return "", errors.New("node name must contain 1-80 characters")
	}
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return "", err
	}
	raw := base64.RawURLEncoding.EncodeToString(secret)
	digest := sha256.Sum256([]byte(raw))
	hash := hex.EncodeToString(digest[:])
	created := ProbeToken{NodeID: nodeID, NodeName: nodeName, Hash: hash, CreatedAt: time.Now().UTC()}

	s.mu.Lock()
	defer s.mu.Unlock()
	for key, token := range s.tokens {
		if token.NodeID == nodeID {
			delete(s.tokens, key)
		}
	}
	s.tokens[hash] = created
	if existing, ok := s.nodes[nodeID]; ok {
		existing.Name = nodeName
		s.nodes[nodeID] = existing
	} else {
		s.nodes[nodeID] = shared.Report{NodeID: nodeID, Name: nodeName, FirstSeen: time.Now().UTC(), UsageMonth: time.Now().UTC().Format("2006-01")}
	}
	if err := s.saveLocked("tokens.json", s.tokens); err != nil {
		delete(s.tokens, hash)
		return "", err
	}
	if err := s.saveNodesLocked(); err != nil {
		return "", err
	}
	return raw, nil
}

func (s *Store) DeleteNode(nodeID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.nodes, nodeID)
	delete(s.history, nodeID)
	delete(s.lastCompact, nodeID)
	for key, token := range s.tokens {
		if token.NodeID == nodeID {
			delete(s.tokens, key)
		}
	}
	if err := s.saveLocked("tokens.json", s.tokens); err != nil {
		return err
	}
	if err := s.saveNodesLocked(); err != nil {
		return err
	}
	if err := os.Remove(s.historyPath(nodeID)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (s *Store) Update(report shared.Report, nodeName string) error {
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()
	previous, found := s.nodes[report.NodeID]
	if found {
		report.FirstSeen = previous.FirstSeen
		elapsed := now.Sub(previous.LastSeen).Seconds()
		if elapsed > 0 && elapsed < 3600 {
			report.RXRate = positiveRate(report.RXBytes, previous.RXBytes, elapsed)
			report.TXRate = positiveRate(report.TXBytes, previous.TXBytes, elapsed)
		}
		preserveManaged(&report, previous)
		if report.Country == "" && previous.MetadataLookupIP == identityIP(report) {
			report.Country, report.CountryCode = previous.Country, previous.CountryCode
			report.Region, report.City, report.ISP, report.ASN = previous.Region, previous.City, previous.ISP, previous.ASN
		}
		month := now.Format("2006-01")
		report.UsageMonth = month
		if previous.UsageMonth == month {
			report.MonthlyUsageBytes = previous.MonthlyUsageBytes
			report.MonthlyRXUsageBytes = previous.MonthlyRXUsageBytes
			report.MonthlyTXUsageBytes = previous.MonthlyTXUsageBytes
		} else {
			report.MonthlyUsageBytes = 0
			report.MonthlyRXUsageBytes = 0
			report.MonthlyTXUsageBytes = 0
		}
		if !previous.LastSeen.IsZero() {
			rxDelta := counterDelta(report.RXBytes, previous.RXBytes)
			txDelta := counterDelta(report.TXBytes, previous.TXBytes)
			report.TotalRXUsageBytes = saturatingAdd(previous.TotalRXUsageBytes, rxDelta)
			report.TotalTXUsageBytes = saturatingAdd(previous.TotalTXUsageBytes, txDelta)
			report.MonthlyRXUsageBytes = saturatingAdd(report.MonthlyRXUsageBytes, rxDelta)
			report.MonthlyTXUsageBytes = saturatingAdd(report.MonthlyTXUsageBytes, txDelta)
			legacyTotal := previous.TotalUsageBytes - minUint64(previous.TotalUsageBytes, saturatingAdd(previous.TotalRXUsageBytes, previous.TotalTXUsageBytes))
			legacyMonth := previous.MonthlyUsageBytes - minUint64(previous.MonthlyUsageBytes, saturatingAdd(previous.MonthlyRXUsageBytes, previous.MonthlyTXUsageBytes))
			report.TotalUsageBytes = saturatingAdd(legacyTotal, saturatingAdd(report.TotalRXUsageBytes, report.TotalTXUsageBytes))
			report.MonthlyUsageBytes = saturatingAdd(legacyMonth, saturatingAdd(report.MonthlyRXUsageBytes, report.MonthlyTXUsageBytes))
		} else {
			report.TotalUsageBytes = previous.TotalUsageBytes
			report.TotalRXUsageBytes = previous.TotalRXUsageBytes
			report.TotalTXUsageBytes = previous.TotalTXUsageBytes
		}
	} else {
		report.FirstSeen = now
		report.UsageMonth = now.Format("2006-01")
		report.QuotaMode = "total"
	}
	if report.FirstSeen.IsZero() {
		report.FirstSeen = now
	}
	report.Name = nodeName
	report.LastSeen = now
	report.CPUPercent = clampPercent(report.CPUPercent)
	report.MemPercent = percent(report.MemUsedBytes, report.MemTotalBytes)
	report.DiskPercent = percent(report.DiskUsedBytes, report.DiskTotalBytes)
	point := shared.HistoryPoint{At: now, CPUPercent: report.CPUPercent, MemPercent: report.MemPercent, RXRate: report.RXRate, TXRate: report.TXRate, Latencies: append([]shared.ProbeResult(nil), report.Latencies...), Samples: 1}
	report.History = nil
	if err := s.appendHistoryLocked(report.NodeID, point, now); err != nil {
		return err
	}
	s.nodes[report.NodeID] = report
	return s.saveNodesLocked()
}

func preserveManaged(report *shared.Report, previous shared.Report) {
	report.MetadataLookupIP = previous.MetadataLookupIP
	report.ExpiryDate = previous.ExpiryDate
	report.PriceUSD = previous.PriceUSD
	report.MonthlyQuotaBytes = previous.MonthlyQuotaBytes
	report.TotalQuotaBytes = previous.TotalQuotaBytes
	report.QuotaMode = previous.QuotaMode
	report.PublicIPv4Override = previous.PublicIPv4Override
	report.PublicIPv6Override = previous.PublicIPv6Override
	report.CountryOverride = previous.CountryOverride
	report.CountryCodeOverride = previous.CountryCodeOverride
	report.RegionOverride = previous.RegionOverride
	report.CityOverride = previous.CityOverride
	report.ISPOverride = previous.ISPOverride
	report.ASNOverride = previous.ASNOverride
}

func counterDelta(current, previous uint64) uint64 {
	if current >= previous {
		return current - previous
	}
	return current
}

func minUint64(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

func saturatingAdd(a, b uint64) uint64 {
	if ^uint64(0)-a < b {
		return ^uint64(0)
	}
	return a + b
}

func identityIP(node shared.Report) string {
	if node.PublicIPv4 != "" {
		return node.PublicIPv4
	}
	return node.PublicIPv6
}

func (s *Store) Node(nodeID string) (shared.Report, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	node, ok := s.nodes[nodeID]
	if !ok {
		return shared.Report{}, false
	}
	return applyOverrides(node), true
}

func (s *Store) Nodes() []shared.Report {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]shared.Report, 0, len(s.nodes))
	for _, raw := range s.nodes {
		node := applyOverrides(raw)
		if ip := identityIP(node); ip != "" {
			if meta, ok := s.ipMetadata[ip]; ok {
				node.GeoAttemptedAt = meta.AttemptedAt
				node.GeoSuccess = meta.Success
				node.GeoNextRetryAt = meta.NextRetryAt
				// When cache has a successful lookup but node fields were never applied, fill them.
				if meta.Success && node.Country == "" && node.CountryOverride == "" {
					node.Country, node.CountryCode = meta.Country, meta.CountryCode
					node.Region, node.City = meta.Region, meta.City
					node.ISP, node.ASN = meta.ISP, meta.ASN
				}
			}
		}
		result = append(result, node)
	}
	sort.Slice(result, func(i, j int) bool {
		if strings.EqualFold(result[i].Name, result[j].Name) {
			return result[i].NodeID < result[j].NodeID
		}
		return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
	})
	return result
}

func applyOverrides(node shared.Report) shared.Report {
	if node.PublicIPv4Override != "" {
		node.PublicIPv4 = node.PublicIPv4Override
	}
	if node.PublicIPv6Override != "" {
		node.PublicIPv6 = node.PublicIPv6Override
	}
	if node.CountryOverride != "" {
		node.Country = node.CountryOverride
	}
	if node.CountryCodeOverride != "" {
		node.CountryCode = node.CountryCodeOverride
	}
	if node.RegionOverride != "" {
		node.Region = node.RegionOverride
	}
	if node.CityOverride != "" {
		node.City = node.CityOverride
	}
	if node.ISPOverride != "" {
		node.ISP = node.ISPOverride
	}
	if node.ASNOverride != "" {
		node.ASN = node.ASNOverride
	}
	return node
}

func (s *Store) History(nodeID string, since time.Time) ([]shared.HistoryPoint, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.nodes[nodeID]; !ok {
		return nil, false
	}
	points := s.history[nodeID]
	result := make([]shared.HistoryPoint, 0, len(points))
	for _, point := range points {
		if !point.At.Before(since) {
			result = append(result, point)
		}
	}
	return result, true
}

func (s *Store) MetadataIP(nodeID string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	node, ok := s.nodes[nodeID]
	if !ok {
		return "", false
	}
	return identityIP(node), true
}

func (s *Store) ClaimMetadataLookup(nodeID, ip string) bool {
	if ip == "" {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	node, ok := s.nodes[nodeID]
	if !ok || node.MetadataLookupIP == ip {
		return false
	}
	node.MetadataLookupIP = ip
	s.nodes[nodeID] = node
	if err := s.saveNodesLocked(); err != nil {
		return false
	}
	return true
}

func (s *Store) SetGeoMetadata(nodeID, ip, country, countryCode, region, city, isp, asn string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	node, ok := s.nodes[nodeID]
	if !ok {
		return errors.New("node not found")
	}
	if node.MetadataLookupIP != ip {
		return errors.New("metadata IP changed before lookup completed")
	}
	countryCode = strings.ToUpper(countryCode)
	if node.Country == country && node.CountryCode == countryCode && node.Region == region && node.City == city && node.ISP == isp && node.ASN == asn {
		return nil
	}
	node.Country, node.CountryCode = country, countryCode
	node.Region, node.City, node.ISP, node.ASN = region, city, isp, asn
	s.nodes[nodeID] = node
	return s.saveNodesLocked()
}

func (s *Store) IPMetadata(ip string) (shared.IPMetadata, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	metadata, ok := s.ipMetadata[ip]
	return metadata, ok
}

func (s *Store) ClaimIPMetadataLookup(ip string) bool {
	if ip == "" {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	previous, exists := s.ipMetadata[ip]
	if exists {
		if previous.Success || (!previous.NextRetryAt.IsZero() && now.Before(previous.NextRetryAt)) {
			return false
		}
		if previous.NextRetryAt.IsZero() && !previous.AttemptedAt.IsZero() && now.Sub(previous.AttemptedAt) < failedGeoRetryAfter {
			return false
		}
	} else if len(s.ipMetadata) >= 2048 {
		oldestIP := ""
		var oldest time.Time
		for candidate, value := range s.ipMetadata {
			if oldestIP == "" || value.AttemptedAt.Before(oldest) {
				oldestIP, oldest = candidate, value.AttemptedAt
			}
		}
		if oldestIP != "" {
			delete(s.ipMetadata, oldestIP)
		}
	}
	s.ipMetadata[ip] = shared.IPMetadata{AttemptedAt: now}
	if err := s.saveLocked("ip-metadata.json", s.ipMetadata); err != nil {
		if exists {
			s.ipMetadata[ip] = previous
		} else {
			delete(s.ipMetadata, ip)
		}
		return false
	}
	return true
}

func (s *Store) SaveIPMetadata(ip string, metadata shared.IPMetadata) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	previous := s.ipMetadata[ip]
	if !metadata.Success {
		metadata.FailureCount = previous.FailureCount + 1
		delay := time.Hour
		if metadata.FailureCount == 2 {
			delay = 6 * time.Hour
		} else if metadata.FailureCount >= 3 {
			delay = 24 * time.Hour
		}
		metadata.NextRetryAt = metadata.AttemptedAt.Add(delay)
	} else {
		metadata.FailureCount = 0
		metadata.NextRetryAt = time.Time{}
	}
	if _, exists := s.ipMetadata[ip]; !exists {
		metadata.AttemptedAt = time.Now().UTC()
	}
	if metadata.AttemptedAt.IsZero() {
		metadata.AttemptedAt = time.Now().UTC()
	}
	s.ipMetadata[ip] = metadata
	for nodeID, node := range s.nodes {
		if identityIP(node) == ip {
			node.GeoAttemptedAt = metadata.AttemptedAt
			node.GeoSuccess = metadata.Success
			node.GeoNextRetryAt = metadata.NextRetryAt
			s.nodes[nodeID] = node
		}
	}
	if err := s.saveLocked("ip-metadata.json", s.ipMetadata); err != nil {
		return err
	}
	return s.saveNodesLocked()
}

func (s *Store) UpdateMetadata(nodeID string, update MetadataUpdate) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	node, ok := s.nodes[nodeID]
	if !ok {
		return errors.New("node not found")
	}
	node.ExpiryDate = update.ExpiryDate
	node.PriceUSD = update.PriceUSD
	node.MonthlyQuotaBytes, node.TotalQuotaBytes = update.MonthlyQuotaBytes, update.TotalQuotaBytes
	node.QuotaMode = update.QuotaMode
	node.BandwidthMbps = update.BandwidthMbps
	node.PublicIPv4Override, node.PublicIPv6Override = update.PublicIPv4Override, update.PublicIPv6Override
	node.CountryOverride, node.CountryCodeOverride = update.CountryOverride, update.CountryCodeOverride
	node.RegionOverride, node.CityOverride = update.RegionOverride, update.CityOverride
	node.ISPOverride, node.ASNOverride = update.ISPOverride, update.ASNOverride
	s.nodes[nodeID] = node
	return s.saveNodesLocked()
}

func positiveRate(current, previous uint64, elapsed float64) float64 {
	if current < previous || elapsed <= 0 {
		return 0
	}
	return float64(current-previous) / elapsed
}

func percent(used, total uint64) float64 {
	if total == 0 {
		return 0
	}
	return clampPercent(float64(used) * 100 / float64(total))
}

func clampPercent(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}
