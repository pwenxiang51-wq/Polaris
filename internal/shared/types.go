package shared

import "time"

type ProbeResult struct {
	Name        string  `json:"name"`
	Target      string  `json:"target"`
	LatencyMS   float64 `json:"latency_ms"`
	LossPercent float64 `json:"loss_percent"`
	Reachable   bool    `json:"reachable"`
}

type HistoryPoint struct {
	At         time.Time     `json:"at"`
	CPUPercent float64       `json:"cpu_percent"`
	MemPercent float64       `json:"mem_percent"`
	RXRate     float64       `json:"rx_rate"`
	TXRate     float64       `json:"tx_rate"`
	Latencies  []ProbeResult `json:"latencies,omitempty"`
	Samples    int           `json:"samples,omitempty"`
}

type Report struct {
	NodeID         string         `json:"node_id"`
	Name           string         `json:"name"`
	Hostname       string         `json:"hostname"`
	OS             string         `json:"os"`
	OSID           string         `json:"os_id,omitempty"`
	Kernel         string         `json:"kernel"`
	Arch           string         `json:"arch"`
	Country        string         `json:"country"`
	CountryCode    string         `json:"country_code"`
	Region         string         `json:"region"`
	City           string         `json:"city"`
	ISP            string         `json:"isp"`
	ASN            string         `json:"asn"`
	PublicIPv4     string         `json:"public_ipv4"`
	PublicIPv6     string         `json:"public_ipv6"`
	CPUPercent     float64        `json:"cpu_percent"`
	MemPercent     float64        `json:"mem_percent"`
	DiskPercent    float64        `json:"disk_percent"`
	CPUCount       int            `json:"cpu_count"`
	MemTotalBytes  uint64         `json:"mem_total_bytes"`
	MemUsedBytes   uint64         `json:"mem_used_bytes"`
	SwapTotalBytes uint64         `json:"swap_total_bytes"`
	SwapUsedBytes  uint64         `json:"swap_used_bytes"`
	DiskTotalBytes uint64         `json:"disk_total_bytes"`
	DiskUsedBytes  uint64         `json:"disk_used_bytes"`
	Load1          float64        `json:"load1"`
	Load5          float64        `json:"load5"`
	Load15         float64        `json:"load15"`
	UptimeSeconds  uint64         `json:"uptime_seconds"`
	RXBytes        uint64         `json:"rx_bytes"`
	TXBytes        uint64         `json:"tx_bytes"`
	RXRate         float64        `json:"rx_rate"`
	TXRate         float64        `json:"tx_rate"`
	Latencies      []ProbeResult  `json:"latencies"`
	FirstSeen      time.Time      `json:"first_seen"`
	LastSeen       time.Time      `json:"last_seen"`
	History        []HistoryPoint `json:"history,omitempty"`
	DemoData       bool           `json:"demo_data,omitempty"`

	MetadataLookupIP    string    `json:"metadata_lookup_ip,omitempty"`
	GeoAttemptedAt      time.Time `json:"geo_attempted_at,omitempty"`
	GeoSuccess          bool      `json:"geo_success"`
	GeoNextRetryAt      time.Time `json:"geo_next_retry_at,omitempty"`
	ExpiryDate          string    `json:"expiry_date,omitempty"`
	PriceUSD            *float64  `json:"price_usd,omitempty"`
	MonthlyQuotaBytes   uint64    `json:"monthly_quota_bytes,omitempty"`
	TotalQuotaBytes     uint64    `json:"total_quota_bytes,omitempty"`
	MonthlyUsageBytes   uint64    `json:"monthly_usage_bytes,omitempty"`
	TotalUsageBytes     uint64    `json:"total_usage_bytes,omitempty"`
	MonthlyRXUsageBytes uint64    `json:"monthly_rx_usage_bytes,omitempty"`
	MonthlyTXUsageBytes uint64    `json:"monthly_tx_usage_bytes,omitempty"`
	TotalRXUsageBytes   uint64    `json:"total_rx_usage_bytes,omitempty"`
	TotalTXUsageBytes   uint64    `json:"total_tx_usage_bytes,omitempty"`
	QuotaMode           string    `json:"quota_mode,omitempty"`
	BandwidthMbps       int       `json:"bandwidth_mbps,omitempty"`
	UsageMonth          string    `json:"usage_month,omitempty"`

	PublicIPv4Override  string `json:"public_ipv4_override,omitempty"`
	PublicIPv6Override  string `json:"public_ipv6_override,omitempty"`
	CountryOverride     string `json:"country_override,omitempty"`
	CountryCodeOverride string `json:"country_code_override,omitempty"`
	RegionOverride      string `json:"region_override,omitempty"`
	CityOverride        string `json:"city_override,omitempty"`
	ISPOverride         string `json:"isp_override,omitempty"`
	ASNOverride         string `json:"asn_override,omitempty"`
}

type IPMetadata struct {
	AttemptedAt     time.Time `json:"attempted_at"`
	Success         bool      `json:"success"`
	FailureCount    int       `json:"failure_count,omitempty"`
	NextRetryAt     time.Time `json:"next_retry_at,omitempty"`
	FailureCategory string    `json:"failure_category,omitempty"`
	Country         string    `json:"country,omitempty"`
	CountryCode     string    `json:"country_code,omitempty"`
	Region          string    `json:"region,omitempty"`
	City            string    `json:"city,omitempty"`
	ISP             string    `json:"isp,omitempty"`
	ASN             string    `json:"asn,omitempty"`
}
