package store

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pwenxiang51-wq/polaris/internal/shared"
)

func TestTokenIsStoredAsHashAndPersists(t *testing.T) {
	dir := t.TempDir()
	st, err := Open(dir)
	if err != nil {
		t.Fatalf("Open returned an error: %v", err)
	}
	raw, err := st.CreateToken("test-node-01", "Test node")
	if err != nil {
		t.Fatalf("CreateToken returned an error: %v", err)
	}
	binding, ok := st.LookupToken(raw)
	if !ok || binding.NodeID != "test-node-01" {
		t.Fatal("new token was not bound to its node")
	}
	stored, err := os.ReadFile(filepath.Join(dir, "tokens.json"))
	if err != nil {
		t.Fatalf("could not read token file: %v", err)
	}
	if strings.Contains(string(stored), raw) {
		t.Fatal("raw token was written to persistent storage")
	}
	if !strings.Contains(string(stored), binding.Hash) {
		t.Fatal("token hash was not written to persistent storage")
	}

	reopened, err := Open(dir)
	if err != nil {
		t.Fatalf("reopening store failed: %v", err)
	}
	if _, ok := reopened.LookupToken(raw); !ok {
		t.Fatal("token binding did not persist")
	}
}

func TestCreateTokenRotatesExistingNodeToken(t *testing.T) {
	st, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	first, err := st.CreateToken("test-node", "Node")
	if err != nil {
		t.Fatal(err)
	}
	second, err := st.CreateToken("test-node", "Node renamed")
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("rotated token unexpectedly matched the old token")
	}
	if _, ok := st.LookupToken(first); ok {
		t.Fatal("old token remained valid after rotation")
	}
	if binding, ok := st.LookupToken(second); !ok || binding.NodeName != "Node renamed" {
		t.Fatal("new token was not bound to updated node metadata")
	}
}

func TestReportCalculatesRatesAndKeepsBoundedHistory(t *testing.T) {
	st, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	_, err = st.CreateToken("node-01", "Node 01")
	if err != nil {
		t.Fatal(err)
	}
	first := shared.Report{NodeID: "node-01", CPUCount: 2, CPUPercent: 25, MemTotalBytes: 1000, MemUsedBytes: 500, DiskTotalBytes: 2000, DiskUsedBytes: 1000, RXBytes: 1000, TXBytes: 2000}
	if err := st.Update(first, "Node 01"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(15 * time.Millisecond)
	second := first
	second.CPUPercent = 150
	second.MemUsedBytes = 700
	second.RXBytes = 2000
	second.TXBytes = 2500
	if err := st.Update(second, "Node 01"); err != nil {
		t.Fatal(err)
	}
	nodes := st.Nodes()
	if len(nodes) != 1 {
		t.Fatalf("expected one node, got %d", len(nodes))
	}
	if nodes[0].CPUPercent != 100 {
		t.Fatalf("CPU percentage was not clamped: %v", nodes[0].CPUPercent)
	}
	if nodes[0].MemPercent != 70 || nodes[0].DiskPercent != 50 {
		t.Fatalf("resource percentages were not calculated: memory=%v disk=%v", nodes[0].MemPercent, nodes[0].DiskPercent)
	}
	history, ok := st.History("node-01", time.Now().UTC().Add(-24*time.Hour))
	if !ok || len(history) != 2 {
		t.Fatalf("expected two history points, got %d", len(history))
	}
}

func TestTokenFileIsValidJSON(t *testing.T) {
	dir := t.TempDir()
	st, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.CreateToken("node-02", "Node 02"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "tokens.json"))
	if err != nil {
		t.Fatal(err)
	}
	var value map[string]ProbeToken
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatalf("token file is not valid JSON: %v", err)
	}
}

func TestHistoryIsStoredSeparatelyFromNodeStateAndPersists(t *testing.T) {
	dir := t.TempDir()
	st, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.CreateToken("history-node", "History node"); err != nil {
		t.Fatal(err)
	}
	report := shared.Report{NodeID: "history-node", CPUCount: 2, CPUPercent: 12, MemTotalBytes: 1000, MemUsedBytes: 500, RXBytes: 100, TXBytes: 200, Latencies: []shared.ProbeResult{{Name: "CT", Target: "example.com:53", LatencyMS: 23, LossPercent: 0, Reachable: true}}}
	if err := st.Update(report, "History node"); err != nil {
		t.Fatal(err)
	}
	nodesJSON, err := os.ReadFile(filepath.Join(dir, "nodes.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(nodesJSON), `"history"`) {
		t.Fatal("nodes.json should not embed history points")
	}
	historyPath := filepath.Join(dir, "history", "history-node.jsonl")
	if _, err := os.Stat(historyPath); err != nil {
		t.Fatalf("separate history file was not created: %v", err)
	}
	reopened, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	points, ok := reopened.History("history-node", time.Now().UTC().Add(-24*time.Hour))
	if !ok || len(points) != 1 || points[0].Latencies[0].Name != "CT" {
		t.Fatalf("history did not persist: %#v", points)
	}
}

func TestLegacyNodeHistoryMigratesToHistoryFile(t *testing.T) {
	dir := t.TempDir()
	legacyPoint := shared.HistoryPoint{At: time.Now().UTC().Add(-time.Minute), CPUPercent: 42, Samples: 1}
	legacy := map[string]shared.Report{"legacy-node": {NodeID: "legacy-node", Name: "Legacy", CPUCount: 1, History: []shared.HistoryPoint{legacyPoint}}}
	data, err := json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "nodes.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	st, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	points, ok := st.History("legacy-node", time.Now().UTC().Add(-24*time.Hour))
	if !ok || len(points) != 1 || points[0].CPUPercent != 42 {
		t.Fatalf("legacy history was not migrated: %#v", points)
	}
	currentState, err := os.ReadFile(filepath.Join(dir, "nodes.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(currentState), `"history"`) {
		t.Fatal("legacy history remained embedded in nodes.json")
	}
}

func TestTrafficAccountingHandlesCounterReset(t *testing.T) {
	st, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.CreateToken("traffic-node", "Traffic"); err != nil {
		t.Fatal(err)
	}
	base := shared.Report{NodeID: "traffic-node", CPUCount: 1, MemTotalBytes: 100, DiskTotalBytes: 100, RXBytes: 1000, TXBytes: 2000}
	if err := st.Update(base, "Traffic"); err != nil {
		t.Fatal(err)
	}
	node, _ := st.Node("traffic-node")
	if node.TotalUsageBytes != 0 {
		t.Fatalf("first report should establish counters only, got %d", node.TotalUsageBytes)
	}
	base.RXBytes, base.TXBytes = 1500, 2500
	if err := st.Update(base, "Traffic"); err != nil {
		t.Fatal(err)
	}
	node, _ = st.Node("traffic-node")
	if node.TotalUsageBytes != 1000 || node.MonthlyUsageBytes != 1000 {
		t.Fatalf("unexpected usage after increments: total=%d monthly=%d", node.TotalUsageBytes, node.MonthlyUsageBytes)
	}
	base.RXBytes, base.TXBytes = 100, 50
	if err := st.Update(base, "Traffic"); err != nil {
		t.Fatal(err)
	}
	node, _ = st.Node("traffic-node")
	if node.TotalUsageBytes != 1150 || node.MonthlyUsageBytes != 1150 {
		t.Fatalf("counter reset was not handled: total=%d monthly=%d", node.TotalUsageBytes, node.MonthlyUsageBytes)
	}
}

func TestHistoryCompactionKeepsRawRecentAndAggregatesOldPoints(t *testing.T) {
	now := time.Now().UTC()
	minuteBucket := now.Add(-2 * time.Hour).Truncate(time.Minute)
	points := []shared.HistoryPoint{
		{At: now.Add(-30 * time.Second), CPUPercent: 10, Samples: 1},
		{At: minuteBucket.Add(2 * time.Second), CPUPercent: 20, Samples: 1},
		{At: minuteBucket.Add(20 * time.Second), CPUPercent: 40, Samples: 1},
		{At: now.Add(-25 * time.Hour), CPUPercent: 60, Samples: 1},
		{At: now.Add(-32 * 24 * time.Hour), CPUPercent: 80, Samples: 1},
	}
	got := compactHistory(points, now)
	if len(got) != 3 {
		t.Fatalf("expected recent raw, minute aggregate, and hourly aggregate; got %#v", got)
	}
	var rawFound, minuteFound, hourFound bool
	for _, point := range got {
		age := now.Sub(point.At)
		if age < rawHistoryWindow && point.CPUPercent == 10 && point.Samples == 1 {
			rawFound = true
		}
		if age > rawHistoryWindow && age < minuteHistoryWindow && point.Samples == 2 && point.CPUPercent == 30 {
			minuteFound = true
		}
		if age > minuteHistoryWindow && point.CPUPercent == 60 {
			hourFound = true
		}
	}
	if !rawFound {
		t.Fatalf("recent raw point was not retained: %#v", got)
	}
	if !minuteFound {
		t.Fatalf("minute aggregate was not produced: %#v", got)
	}
	if !hourFound {
		t.Fatalf("hourly aggregate was not produced: %#v", got)
	}
}

func TestMetadataOverridesAndLookupClaim(t *testing.T) {
	st, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.CreateToken("meta-node", "Metadata"); err != nil {
		t.Fatal(err)
	}
	report := shared.Report{NodeID: "meta-node", CPUCount: 1, PublicIPv4: "8.8.4.4", MemTotalBytes: 100, DiskTotalBytes: 100}
	if err := st.Update(report, "Metadata"); err != nil {
		t.Fatal(err)
	}
	if !st.ClaimMetadataLookup("meta-node", "8.8.4.4") || st.ClaimMetadataLookup("meta-node", "8.8.4.4") {
		t.Fatal("metadata lookup should be claimed exactly once per IP")
	}
	price := 3.5
	if err := st.UpdateMetadata("meta-node", MetadataUpdate{PriceUSD: &price, CountryOverride: "Manual country", CountryCodeOverride: "US"}); err != nil {
		t.Fatal(err)
	}
	node, ok := st.Node("meta-node")
	if !ok || node.Country != "Manual country" || node.CountryCode != "US" || node.PriceUSD == nil || *node.PriceUSD != price {
		t.Fatalf("manual metadata did not apply: %#v", node)
	}
}

func TestIPMetadataCacheClaimsOnceAndPersists(t *testing.T) {
	dir := t.TempDir()
	st, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	ip := "8.8.8.8"
	if !st.ClaimIPMetadataLookup(ip) {
		t.Fatal("first lookup claim should succeed")
	}
	if st.ClaimIPMetadataLookup(ip) {
		t.Fatal("same IP lookup should only be claimed once")
	}
	metadata := shared.IPMetadata{
		AttemptedAt: time.Now().UTC(), Success: true, Country: "United States",
		CountryCode: "US", Region: "California", City: "Mountain View", ISP: "Example ISP", ASN: "AS15169",
	}
	if err := st.SaveIPMetadata(ip, metadata); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	cached, ok := reopened.IPMetadata(ip)
	if !ok || !cached.Success || cached.CountryCode != "US" || cached.ISP != "Example ISP" {
		t.Fatalf("IP metadata cache did not persist: %#v, exists=%v", cached, ok)
	}
	if reopened.ClaimIPMetadataLookup(ip) {
		t.Fatal("successful cached IP should not trigger another lookup claim")
	}
}

func TestFailedIPMetadataLookupCanRetryAfter24Hours(t *testing.T) {
	dir := t.TempDir()
	st, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	ip := "192.0.2.44"
	if !st.ClaimIPMetadataLookup(ip) {
		t.Fatal("initial lookup claim should succeed")
	}
	if err := st.SaveIPMetadata(ip, shared.IPMetadata{AttemptedAt: time.Now().UTC(), Success: false}); err != nil {
		t.Fatal(err)
	}
	if st.ClaimIPMetadataLookup(ip) {
		t.Fatal("failed lookup should not retry before 24 hours")
	}
	if err := st.SaveIPMetadata(ip, shared.IPMetadata{AttemptedAt: time.Now().UTC().Add(-25 * time.Hour), Success: false}); err != nil {
		t.Fatal(err)
	}
	if !st.ClaimIPMetadataLookup(ip) {
		t.Fatal("failed lookup should be retryable after 24 hours")
	}
	if st.ClaimIPMetadataLookup(ip) {
		t.Fatal("retry should be rate-limited for the next 24 hours")
	}
}

func TestHistoryRetentionIs31Days(t *testing.T) {
	now := time.Now().UTC()
	got := compactHistory([]shared.HistoryPoint{
		{At: now.Add(-31*24*time.Hour + time.Second), CPUPercent: 21, Samples: 1},
		{At: now.Add(-31*24*time.Hour - 2*time.Hour), CPUPercent: 84, Samples: 1},
	}, now)
	if len(got) != 1 {
		t.Fatalf("expected only the sample inside the 31-day retention window, got %d points", len(got))
	}
	if got[0].CPUPercent != 21 {
		t.Fatalf("the point inside the 31-day retention window was not retained: %#v", got[0])
	}
}

func TestStartupRemovesOrphanHistoryFiles(t *testing.T) {
	dir := t.TempDir()
	historyDir := filepath.Join(dir, "history")
	if err := os.MkdirAll(historyDir, 0700); err != nil {
		t.Fatal(err)
	}
	nodes := map[string]shared.Report{"active": {NodeID: "active", Name: "Active"}}
	nodesJSON, err := json.Marshal(nodes)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "nodes.json"), nodesJSON, 0600); err != nil {
		t.Fatal(err)
	}
	pointJSON, err := json.Marshal(shared.HistoryPoint{At: time.Now().UTC(), CPUPercent: 10, Samples: 1})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(historyDir, "active.jsonl"), append(pointJSON, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(historyDir, "orphan.jsonl"), append(pointJSON, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(historyDir, "orphan.jsonl")); !os.IsNotExist(err) {
		t.Fatalf("orphan history file was not removed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(historyDir, "active.jsonl")); err != nil {
		t.Fatalf("active node history was unexpectedly removed: %v", err)
	}
}

func TestOversizedHistoryIsCompactedAndTruncatedAtStartup(t *testing.T) {
	dir := t.TempDir()
	historyDir := filepath.Join(dir, "history")
	if err := os.MkdirAll(historyDir, 0700); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	nodes := map[string]shared.Report{"oversized": {NodeID: "oversized", Name: "Oversized", CPUCount: 1}}
	nodesJSON, err := json.Marshal(nodes)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "nodes.json"), nodesJSON, 0600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(historyDir, "oversized.jsonl")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	writer := bufio.NewWriterSize(file, 128*1024)
	padding := strings.Repeat("x", 4096)
	const count = 9000
	for i := 0; i < count; i++ {
		point := shared.HistoryPoint{At: now.Add(time.Duration(i-count) * time.Second), CPUPercent: float64(i % 100), Samples: 1}
		encoded, err := json.Marshal(point)
		if err != nil {
			_ = file.Close()
			t.Fatal(err)
		}
		line := fmt.Sprintf("%s,\"padding\":%q}\n", strings.TrimSuffix(string(encoded), "}"), padding)
		if _, err := writer.WriteString(line); err != nil {
			_ = file.Close()
			t.Fatal(err)
		}
	}
	if err := writer.Flush(); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if before.Size() <= maxHistoryFileBytes {
		t.Fatalf("test history file was not oversized: %d bytes", before.Size())
	}
	st, err := Open(dir)
	if err != nil {
		t.Fatalf("startup failed on oversized history: %v", err)
	}
	after, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if after.Size() >= before.Size() || after.Size() > maxHistoryFileBytes {
		t.Fatalf("oversized history was not compacted: before=%d after=%d", before.Size(), after.Size())
	}
	points, ok := st.History("oversized", now.Add(-24*time.Hour))
	if !ok || len(points) == 0 {
		t.Fatalf("recent history was not preserved after truncation: %d points, ok=%v", len(points), ok)
	}
	if points[len(points)-1].At.Before(now.Add(-time.Minute)) {
		t.Fatalf("newest history was not kept: %s", points[len(points)-1].At)
	}
}
