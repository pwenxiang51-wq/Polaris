package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pwenxiang51-wq/polaris/internal/shared"
)

const reportInterval = 15 * time.Second

type AgentConfig struct {
	ServerURL    string
	NodeID       string
	NodeName     string
	Token        string
	ProbeTargets []ProbeTarget
}

type ProbeTarget struct {
	Name    string
	Address string
}

var networkClient = &http.Client{
	Timeout: 5 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig:     &tls.Config{MinVersion: tls.VersionTLS12},
		MaxIdleConns:        8,
		MaxIdleConnsPerHost: 2,
		IdleConnTimeout:     30 * time.Second,
		Proxy:               http.ProxyFromEnvironment,
	},
	CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse },
}

func main() {
	log.SetFlags(log.LstdFlags | log.LUTC)
	cfg, err := loadAgentConfig()
	if err != nil {
		log.Fatal(err)
	}
	ctx, stop := signalContext()
	defer stop()
	log.Printf("Polaris probe started for node %s", cfg.NodeID)
	ticker := time.NewTicker(reportInterval)
	defer ticker.Stop()
	for {
		if err := sendReport(ctx, cfg); err != nil && ctx.Err() == nil {
			log.Printf("report failed: %v", err)
		}
		select {
		case <-ctx.Done():
			log.Printf("Polaris probe stopped")
			return
		case <-ticker.C:
		}
	}
}

func signalContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}

func loadAgentConfig() (AgentConfig, error) {
	targets, err := parseProbeTargets(os.Getenv("POLARIS_PROBE_TARGETS"))
	if err != nil {
		return AgentConfig{}, err
	}
	cfg := AgentConfig{
		ServerURL:    strings.TrimRight(strings.TrimSpace(os.Getenv("POLARIS_SERVER_URL")), "/"),
		NodeID:       strings.TrimSpace(os.Getenv("POLARIS_NODE_ID")),
		NodeName:     strings.TrimSpace(os.Getenv("POLARIS_NODE_NAME")),
		Token:        strings.TrimSpace(os.Getenv("POLARIS_AGENT_TOKEN")),
		ProbeTargets: targets,
	}
	if cfg.ServerURL == "" || cfg.NodeID == "" || cfg.Token == "" {
		return AgentConfig{}, errors.New("POLARIS_SERVER_URL, POLARIS_NODE_ID, and POLARIS_AGENT_TOKEN are required")
	}
	parsedURL, urlErr := url.Parse(cfg.ServerURL)
	if urlErr != nil || parsedURL.Scheme != "https" || parsedURL.Host == "" || parsedURL.User != nil || parsedURL.RawQuery != "" || parsedURL.Fragment != "" || (parsedURL.Path != "" && parsedURL.Path != "/") {
		return AgentConfig{}, errors.New("POLARIS_SERVER_URL must be a root HTTPS URL without user info, query, or fragment")
	}
	if len(cfg.NodeID) > 48 || strings.Trim(cfg.NodeID, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789._-") != "" || cfg.NodeID[0] == '.' || cfg.NodeID[0] == '-' || cfg.NodeID[0] == '_' {
		return AgentConfig{}, errors.New("POLARIS_NODE_ID contains invalid characters")
	}
	if len(cfg.Token) < 40 || len(cfg.Token) > 256 {
		return AgentConfig{}, errors.New("POLARIS_AGENT_TOKEN has an invalid length")
	}
	if cfg.NodeName == "" {
		cfg.NodeName = cfg.NodeID
	}
	if len(cfg.NodeName) > 80 {
		return AgentConfig{}, errors.New("POLARIS_NODE_NAME must be at most 80 characters")
	}
	return cfg, nil
}

func sendReport(ctx context.Context, cfg AgentConfig) error {
	collectCtx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()
	report, err := collect(collectCtx, cfg)
	if err != nil {
		return fmt.Errorf("collect metrics: %w", err)
	}
	payload, err := json.Marshal(report)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(collectCtx, http.MethodPost, cfg.ServerURL+"/api/v1/report", strings.NewReader(string(payload)))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Polaris-Probe/1.0")
	resp, err := networkClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("server returned HTTP %d", resp.StatusCode)
	}
	return nil
}

func collect(ctx context.Context, cfg AgentConfig) (shared.Report, error) {
	hostname, _ := os.Hostname()
	publicIPv4, publicIPv6 := discoverLocalPublicAddresses()
	cpuPercent, err := cpuUsage(ctx)
	if err != nil {
		return shared.Report{}, err
	}
	memTotal, memAvailable, swapTotal, swapFree, err := memoryStats()
	if err != nil {
		return shared.Report{}, err
	}
	diskTotal, diskUsed, err := diskStats("/")
	if err != nil {
		return shared.Report{}, err
	}
	load1, load5, load15 := loadAverage()
	uptime := uptimeSeconds()
	rx, tx := networkCounters()
	osName := osPrettyName()
	kernel := kernelRelease()
	return shared.Report{
		NodeID: cfg.NodeID, Name: cfg.NodeName, Hostname: hostname,
		OS: osName, OSID: osReleaseID(), Kernel: kernel, Arch: runtime.GOARCH,
		PublicIPv4: publicIPv4, PublicIPv6: publicIPv6,
		CPUPercent: cpuPercent, CPUCount: runtime.NumCPU(),
		MemTotalBytes: memTotal, MemUsedBytes: maxUint64(0, memTotal-memAvailable),
		SwapTotalBytes: swapTotal, SwapUsedBytes: maxUint64(0, swapTotal-swapFree),
		DiskTotalBytes: diskTotal, DiskUsedBytes: diskUsed,
		Load1: load1, Load5: load5, Load15: load15, UptimeSeconds: uptime,
		RXBytes: rx, TXBytes: tx, Latencies: probeTargets(ctx, cfg.ProbeTargets),
	}, nil
}

func cpuUsage(ctx context.Context) (float64, error) {
	firstTotal, firstIdle, err := cpuCounters()
	if err != nil {
		return 0, err
	}
	timer := time.NewTimer(300 * time.Millisecond)
	select {
	case <-ctx.Done():
		timer.Stop()
		return 0, ctx.Err()
	case <-timer.C:
	}
	secondTotal, secondIdle, err := cpuCounters()
	if err != nil {
		return 0, err
	}
	totalDelta := secondTotal - firstTotal
	idleDelta := secondIdle - firstIdle
	if totalDelta == 0 || idleDelta > totalDelta {
		return 0, nil
	}
	return float64(totalDelta-idleDelta) * 100 / float64(totalDelta), nil
}

func cpuCounters() (uint64, uint64, error) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return 0, 0, errors.New("CPU counters are unavailable")
	}
	fields := strings.Fields(scanner.Text())
	if len(fields) < 5 || fields[0] != "cpu" {
		return 0, 0, errors.New("CPU counters are malformed")
	}
	var total uint64
	var idle uint64
	for index, item := range fields[1:] {
		value, parseErr := strconv.ParseUint(item, 10, 64)
		if parseErr != nil {
			return 0, 0, parseErr
		}
		total += value
		if index == 3 || index == 4 {
			idle += value
		}
	}
	return total, idle, nil
}

func memoryStats() (uint64, uint64, uint64, uint64, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0, 0, 0, err
	}
	defer file.Close()
	values := make(map[string]uint64)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		value, parseErr := strconv.ParseUint(fields[1], 10, 64)
		if parseErr == nil {
			values[strings.TrimSuffix(fields[0], ":")] = value * 1024
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, 0, 0, 0, err
	}
	total := values["MemTotal"]
	available, ok := values["MemAvailable"]
	if !ok {
		available = values["MemFree"] + values["Buffers"] + values["Cached"]
	}
	if total == 0 {
		return 0, 0, 0, 0, errors.New("memory totals are unavailable")
	}
	if available > total {
		available = total
	}
	return total, available, values["SwapTotal"], values["SwapFree"], nil
}

func diskStats(path string) (uint64, uint64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, 0, err
	}
	blockSize := uint64(stat.Bsize)
	total := stat.Blocks * blockSize
	available := stat.Bavail * blockSize
	used := total - minUint64(total, available)
	return total, used, nil
}

func loadAverage() (float64, float64, float64) {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return 0, 0, 0
	}
	fields := strings.Fields(string(data))
	if len(fields) < 3 {
		return 0, 0, 0
	}
	load1, _ := strconv.ParseFloat(fields[0], 64)
	load5, _ := strconv.ParseFloat(fields[1], 64)
	load15, _ := strconv.ParseFloat(fields[2], 64)
	return load1, load5, load15
}

func uptimeSeconds() uint64 {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return 0
	}
	seconds, _ := strconv.ParseFloat(fields[0], 64)
	if seconds < 0 {
		return 0
	}
	return uint64(seconds)
}

func networkCounters() (uint64, uint64) {
	file, err := os.Open("/proc/net/dev")
	if err != nil {
		return 0, 0
	}
	defer file.Close()
	var rxTotal, txTotal uint64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.Contains(line, ":") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		name := strings.TrimSpace(parts[0])
		if name == "lo" || strings.HasPrefix(name, "veth") || strings.HasPrefix(name, "docker") || strings.HasPrefix(name, "br-") || strings.HasPrefix(name, "virbr") || strings.HasPrefix(name, "cni") || strings.HasPrefix(name, "flannel") {
			continue
		}
		fields := strings.Fields(parts[1])
		if len(fields) < 9 {
			continue
		}
		rx, rxErr := strconv.ParseUint(fields[0], 10, 64)
		tx, txErr := strconv.ParseUint(fields[8], 10, 64)
		if rxErr == nil && txErr == nil {
			rxTotal += rx
			txTotal += tx
		}
	}
	return rxTotal, txTotal
}

func osReleaseID() string {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return runtime.GOOS
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ID=") {
			return strings.Trim(strings.TrimPrefix(line, "ID="), "\"'")
		}
	}
	return runtime.GOOS
}

func osPrettyName() string {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return runtime.GOOS
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			return strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"'")
		}
	}
	return runtime.GOOS
}

func kernelRelease() string {
	data, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err != nil {
		return "unknown"
	}
	value := strings.TrimSpace(string(data))
	if value == "" {
		return "unknown"
	}
	return value
}

func parseProbeTargets(raw string) ([]ProbeTarget, error) {
	if strings.TrimSpace(raw) == "" {
		raw = "CT=114.114.114.114:53;CU=123.125.81.6:53;CM=www.10086.cn:443"
	}
	parts := strings.Split(raw, ";")
	if len(parts) > 16 {
		return nil, errors.New("POLARIS_PROBE_TARGETS may contain at most 16 targets")
	}
	seen := make(map[string]bool, len(parts))
	targets := make([]ProbeTarget, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		name, address, ok := strings.Cut(part, "=")
		name, address = strings.TrimSpace(name), strings.TrimSpace(address)
		if !ok || name == "" || len(name) > 32 || address == "" || seen[strings.ToLower(name)] {
			return nil, fmt.Errorf("invalid or duplicate network target %q", part)
		}
		host, portText, err := net.SplitHostPort(address)
		if err != nil || strings.TrimSpace(host) == "" {
			return nil, fmt.Errorf("target %q must use host:port format", name)
		}
		port, err := strconv.Atoi(portText)
		if err != nil || port < 1 || port > 65535 {
			return nil, fmt.Errorf("target %q has an invalid port", name)
		}
		for _, r := range name {
			if !(r == '-' || r == '_' || r == ' ' || r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9') {
				return nil, fmt.Errorf("target name %q contains unsupported characters", name)
			}
		}
		seen[strings.ToLower(name)] = true
		targets = append(targets, ProbeTarget{Name: name, Address: net.JoinHostPort(host, portText)})
	}
	if len(targets) == 0 {
		return nil, errors.New("POLARIS_PROBE_TARGETS contains no usable targets")
	}
	return targets, nil
}

func probeTargets(ctx context.Context, targets []ProbeTarget) []shared.ProbeResult {
	results := make([]shared.ProbeResult, 0, len(targets))
	for _, target := range targets {
		var total time.Duration
		successes := 0
		for attempt := 0; attempt < 3; attempt++ {
			if ctx.Err() != nil {
				break
			}
			started := time.Now()
			dialer := net.Dialer{Timeout: 400 * time.Millisecond}
			conn, err := dialer.DialContext(ctx, "tcp", target.Address)
			if err == nil {
				total += time.Since(started)
				successes++
				_ = conn.Close()
			}
			if attempt < 2 {
				timer := time.NewTimer(50 * time.Millisecond)
				select {
				case <-ctx.Done():
					timer.Stop()
				case <-timer.C:
				}
			}
		}
		latency := 0.0
		if successes > 0 {
			latency = float64(total.Microseconds()) / 1000 / float64(successes)
		}
		loss := float64(3-successes) * 100 / 3
		results = append(results, shared.ProbeResult{Name: target.Name, Target: target.Address, LatencyMS: latency, LossPercent: loss, Reachable: successes > 0})
	}
	return results
}

func discoverLocalPublicAddresses() (string, string) {
	var ipv4, ipv6 string
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range interfaces {
			if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
				continue
			}
			addresses, err := iface.Addrs()
			if err != nil {
				continue
			}
			for _, address := range addresses {
				var ip net.IP
				switch value := address.(type) {
				case *net.IPNet:
					ip = value.IP
				case *net.IPAddr:
					ip = value.IP
				}
				if !isPublicRoutedIP(ip) {
					continue
				}
				if ip4 := ip.To4(); ip4 != nil && ipv4 == "" {
					ipv4 = ip4.String()
				} else if ip4 == nil && ipv6 == "" {
					ipv6 = ip.String()
				}
			}
		}
	}
	if ipv4 == "" {
		if ip := routeSelectedLocalIP("udp4", "1.1.1.1:53"); isPublicRoutedIP(ip) && ip.To4() != nil {
			ipv4 = ip.String()
		}
	}
	if ipv6 == "" {
		if ip := routeSelectedLocalIP("udp6", "[2606:4700:4700::1111]:53"); isPublicRoutedIP(ip) && ip.To4() == nil {
			ipv6 = ip.String()
		}
	}
	return ipv4, ipv6
}

func routeSelectedLocalIP(network, destination string) net.IP {
	conn, err := net.Dial(network, destination)
	if err != nil {
		return nil
	}
	defer conn.Close()
	address, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		return nil
	}
	return address.IP
}

var nonPublicAddressPrefixes = func() []netip.Prefix {
	cidrs := []string{
		"0.0.0.0/8", "10.0.0.0/8", "100.64.0.0/10", "127.0.0.0/8",
		"169.254.0.0/16", "172.16.0.0/12", "192.0.0.0/24", "192.0.2.0/24",
		"192.88.99.0/24", "192.168.0.0/16", "198.18.0.0/15", "198.51.100.0/24",
		"203.0.113.0/24", "224.0.0.0/4", "240.0.0.0/4",
		"2001:db8::/32", "2001::/23", "2002::/16", "64:ff9b:1::/48",
	}
	prefixes := make([]netip.Prefix, 0, len(cidrs))
	for _, cidr := range cidrs {
		prefix, err := netip.ParsePrefix(cidr)
		if err != nil {
			panic("invalid built-in non-public address range: " + cidr)
		}
		prefixes = append(prefixes, prefix)
	}
	return prefixes
}()

func isPublicRoutedIP(ip net.IP) bool {
	if ip == nil || !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsMulticast() || ip.IsUnspecified() {
		return false
	}
	var address netip.Addr
	if ipv4 := ip.To4(); ipv4 != nil {
		var bytes [4]byte
		copy(bytes[:], ipv4)
		address = netip.AddrFrom4(bytes)
	} else if ipv6 := ip.To16(); ipv6 != nil {
		var bytes [16]byte
		copy(bytes[:], ipv6)
		address = netip.AddrFrom16(bytes)
	} else {
		return false
	}
	for _, prefix := range nonPublicAddressPrefixes {
		if prefix.Contains(address) {
			return false
		}
	}
	return true
}

func maxUint64(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

func minUint64(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}
