package main

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/pwenxiang51-wq/polaris/internal/security"
	"github.com/pwenxiang51-wq/polaris/internal/shared"
	"github.com/pwenxiang51-wq/polaris/internal/store"
	"github.com/pwenxiang51-wq/polaris/internal/webui"
)

const sessionCookieName = "polaris_session"
const sessionLifetime = 12 * time.Hour

var nodeIDPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,47}$`)
var countryCodePattern = regexp.MustCompile(`^[A-Z]{2}$`)

type Config struct {
	Listen            string
	AdminUser         string
	AdminPasswordHash string
	SessionSecret     []byte
	DataDir           string
	TurnstileSiteKey  string
	TurnstileSecret   string
	AllowedHostname   string
	ExternalURL       string
	TrustProxyHeaders bool
	EnableGeoLookup   bool
}

type App struct {
	cfg        Config
	store      *store.Store
	limiter    *loginLimiter
	turnstile  *http.Client
	sessionsMu sync.Mutex
	sessions   map[string]int64
}

type loginLimiter struct {
	mu      sync.Mutex
	entries map[string]loginBucket
}

type loginBucket struct {
	Count   int
	ResetAt time.Time
}

type apiError struct {
	Error string `json:"error"`
}

type loginRequest struct {
	Username       string `json:"username"`
	Password       string `json:"password"`
	TurnstileToken string `json:"turnstile_token"`
}

type tokenRequest struct {
	NodeID   string `json:"node_id"`
	NodeName string `json:"node_name"`
}

type nodeMetadataRequest struct {
	ExpiryDate          string   `json:"expiry_date"`
	PriceUSD            *float64 `json:"price_usd"`
	MonthlyQuotaBytes   uint64   `json:"monthly_quota_bytes"`
	TotalQuotaBytes     uint64   `json:"total_quota_bytes"`
	PublicIPv4Override  string   `json:"public_ipv4_override"`
	PublicIPv6Override  string   `json:"public_ipv6_override"`
	CountryOverride     string   `json:"country_override"`
	CountryCodeOverride string   `json:"country_code_override"`
	RegionOverride      string   `json:"region_override"`
	CityOverride        string   `json:"city_override"`
	ISPOverride         string   `json:"isp_override"`
	ASNOverride         string   `json:"asn_override"`
}

type geoLookupResult struct {
	Success     bool   `json:"success"`
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
	Region      string `json:"region"`
	City        string `json:"city"`
	Connection  struct {
		ISP string `json:"isp"`
		ASN string `json:"asn"`
	} `json:"connection"`
}

type turnstileResponse struct {
	Success  bool     `json:"success"`
	Hostname string   `json:"hostname"`
	Action   string   `json:"action"`
	Errors   []string `json:"error-codes"`
}

func loadConfig() (Config, error) {
	adminPasswordHash, err := secretEnv("POLARIS_ADMIN_PASSWORD_HASH")
	if err != nil {
		return Config{}, err
	}
	sessionSecret, err := secretEnv("POLARIS_SESSION_SECRET")
	if err != nil {
		return Config{}, err
	}
	turnstileSecret, err := secretEnv("POLARIS_TURNSTILE_SECRET")
	if err != nil {
		return Config{}, err
	}
	cfg := Config{
		Listen:            envOr("POLARIS_LISTEN", "0.0.0.0:8080"),
		AdminUser:         strings.TrimSpace(os.Getenv("POLARIS_ADMIN_USER")),
		AdminPasswordHash: adminPasswordHash,
		SessionSecret:     []byte(sessionSecret),
		DataDir:           envOr("POLARIS_DATA_DIR", "/data"),
		TurnstileSiteKey:  strings.TrimSpace(os.Getenv("POLARIS_TURNSTILE_SITEKEY")),
		TurnstileSecret:   turnstileSecret,
		AllowedHostname:   strings.ToLower(strings.TrimSpace(os.Getenv("POLARIS_ALLOWED_HOSTNAME"))),
		ExternalURL:       strings.TrimRight(strings.TrimSpace(os.Getenv("POLARIS_EXTERNAL_URL")), "/"),
		TrustProxyHeaders: strings.EqualFold(strings.TrimSpace(os.Getenv("POLARIS_TRUST_PROXY_HEADERS")), "true"),
		EnableGeoLookup:   !strings.EqualFold(strings.TrimSpace(os.Getenv("POLARIS_ENABLE_GEO_LOOKUP")), "false"),
	}
	if cfg.AdminUser == "" || len(cfg.AdminUser) > 64 {
		return Config{}, errors.New("POLARIS_ADMIN_USER must contain 1-64 characters")
	}
	if !strings.HasPrefix(cfg.AdminPasswordHash, "pbkdf2_sha256:") {
		return Config{}, errors.New("POLARIS_ADMIN_PASSWORD_HASH must be a PBKDF2-SHA256 password hash")
	}
	if len(cfg.SessionSecret) < 64 {
		return Config{}, errors.New("POLARIS_SESSION_SECRET must contain at least 64 characters of random data")
	}
	if (cfg.TurnstileSiteKey == "") != (cfg.TurnstileSecret == "") {
		return Config{}, errors.New("both Turnstile keys must be configured together")
	}
	if cfg.TurnstileSiteKey == "" {
		return Config{}, errors.New("Cloudflare Turnstile site key and secret key are required")
	}
	if cfg.AllowedHostname == "" {
		return Config{}, errors.New("POLARIS_ALLOWED_HOSTNAME must match the public dashboard hostname")
	}
	u, err := url.Parse(cfg.ExternalURL)
	if err != nil || u.Scheme != "https" || u.Host == "" || !strings.EqualFold(u.Hostname(), cfg.AllowedHostname) {
		return Config{}, errors.New("POLARIS_EXTERNAL_URL must be an HTTPS URL matching POLARIS_ALLOWED_HOSTNAME")
	}
	return cfg, nil
}

func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func secretEnv(key string) (string, error) {
	if path := strings.TrimSpace(os.Getenv(key + "_FILE")); path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("%s_FILE could not be read", key)
		}
		value := strings.TrimSpace(string(data))
		if value == "" {
			return "", fmt.Errorf("%s_FILE is empty", key)
		}
		return value, nil
	}
	return strings.TrimSpace(os.Getenv(key)), nil
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		fatal(err.Error())
	}
	st, err := store.Open(filepath.Clean(cfg.DataDir))
	if err != nil {
		fatal("could not open data store")
	}
	app := &App{
		cfg: cfg, store: st,
		limiter:   &loginLimiter{entries: make(map[string]loginBucket)},
		turnstile: &http.Client{Timeout: 5 * time.Second},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", app.health)
	mux.HandleFunc("GET /api/public-config", app.publicConfig)
	mux.HandleFunc("POST /api/login", app.login)
	mux.HandleFunc("POST /api/logout", app.logout)
	mux.HandleFunc("GET /api/me", app.me)
	mux.HandleFunc("GET /api/nodes", app.listNodes)
	mux.HandleFunc("GET /api/nodes/{id}/history", app.nodeHistory)
	mux.HandleFunc("POST /api/nodes/token", app.createNodeToken)
	mux.HandleFunc("PATCH /api/nodes/{id}/metadata", app.updateNodeMetadata)
	mux.HandleFunc("DELETE /api/nodes/{id}", app.deleteNode)
	mux.HandleFunc("POST /api/v1/report", app.receiveReport)
	mux.HandleFunc("GET /assets/", app.serveAsset)
	mux.HandleFunc("GET /", app.index)

	server := &http.Server{
		Addr: cfg.Listen, Handler: app.securityHeaders(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      12 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 16,
	}
	log.Printf("Polaris is listening on %s", cfg.Listen)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		fatal("HTTP server stopped")
	}
}

func fatal(message string) {
	fmt.Fprintln(os.Stderr, "Polaris:", message)
	os.Exit(1)
}

func (a *App) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000")
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; base-uri 'self'; object-src 'none'; frame-ancestors 'none'; form-action 'self'; script-src 'self' https://challenges.cloudflare.com; frame-src https://challenges.cloudflare.com; connect-src 'self' https://challenges.cloudflare.com; style-src 'self' 'unsafe-inline'; img-src 'self' data:")
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func (a *App) index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	data, err := webui.Files().Open("index.html")
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	defer data.Close()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.Copy(w, data)
}

func (a *App) serveAsset(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(filepath.ToSlash(filepath.Clean(r.URL.Path)), "/")
	if !strings.HasPrefix(name, "assets/") || strings.Contains(name, "..") {
		http.NotFound(w, r)
		return
	}
	file, err := webui.Files().Open(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()
	if strings.HasSuffix(name, ".css") {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	} else if strings.HasSuffix(name, ".js") {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	} else if strings.HasSuffix(name, ".svg") {
		w.Header().Set("Content-Type", "image/svg+xml")
	}
	_, _ = io.Copy(w, file)
}

func (a *App) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "polaris"})
}

func (a *App) publicConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"turnstile_sitekey": a.cfg.TurnstileSiteKey,
		"turnstile_enabled": a.cfg.TurnstileSiteKey != "",
		"external_url":      a.cfg.ExternalURL,
	})
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	if !a.validMutationOrigin(r) {
		writeError(w, http.StatusForbidden, "Request origin rejected")
		return
	}
	ip := a.clientIP(r)
	if !a.limiter.Allow(ip, time.Now()) {
		writeError(w, http.StatusTooManyRequests, "Too many login attempts. Try again later.")
		return
	}
	var req loginRequest
	if err := decodeJSON(w, r, &req, 8*1024); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request")
		return
	}
	if a.cfg.TurnstileSecret == "" || req.TurnstileToken == "" || !a.verifyTurnstile(r.Context(), req.TurnstileToken, ip) {
		writeError(w, http.StatusUnauthorized, "Human verification failed. Please try again.")
		return
	}
	userOK := hmac.Equal([]byte(req.Username), []byte(a.cfg.AdminUser))
	passwordOK := security.VerifyPassword(req.Password, a.cfg.AdminPasswordHash)
	if !userOK || !passwordOK {
		log.Printf("login rejected for source %s", ip)
		writeError(w, http.StatusUnauthorized, "Invalid username or password")
		return
	}
	cookie, err := a.createSession(a.cfg.AdminUser)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Could not create session")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookieName, Value: cookie, Path: "/", MaxAge: int(sessionLifetime.Seconds()),
		Expires: time.Now().Add(sessionLifetime), HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode,
	})
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "username": a.cfg.AdminUser})
}

func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	if !a.validMutationOrigin(r) {
		writeError(w, http.StatusForbidden, "Request origin rejected")
		return
	}
	a.revokeSession(r)
	http.SetCookie(w, &http.Cookie{Name: sessionCookieName, Value: "", Path: "/", MaxAge: -1, Expires: time.Unix(1, 0), HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode})
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *App) me(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.authenticate(r); !ok {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"authenticated": true, "username": a.cfg.AdminUser})
}

func (a *App) listNodes(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.authenticate(r); !ok {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"nodes": a.store.Nodes()})
}

func (a *App) nodeHistory(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.authenticate(r); !ok {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	nodeID := r.PathValue("id")
	if !nodeIDPattern.MatchString(nodeID) {
		writeError(w, http.StatusBadRequest, "Invalid node ID")
		return
	}
	points, ok := a.store.History(nodeID, time.Now().UTC().Add(-24*time.Hour))
	if !ok {
		writeError(w, http.StatusNotFound, "Node not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"node_id": nodeID, "history": points, "range": "24h"})
}

func (a *App) updateNodeMetadata(w http.ResponseWriter, r *http.Request) {
	if !a.requireAdminMutation(w, r) {
		return
	}
	nodeID := r.PathValue("id")
	if !nodeIDPattern.MatchString(nodeID) {
		writeError(w, http.StatusBadRequest, "Invalid node ID")
		return
	}
	var req nodeMetadataRequest
	if err := decodeJSON(w, r, &req, 16*1024); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid metadata request")
		return
	}
	req.ExpiryDate = strings.TrimSpace(req.ExpiryDate)
	if req.ExpiryDate != "" {
		if _, err := time.Parse("2006-01-02", req.ExpiryDate); err != nil {
			writeError(w, http.StatusBadRequest, "Expiry date must use YYYY-MM-DD")
			return
		}
	}
	if req.PriceUSD != nil && (math.IsNaN(*req.PriceUSD) || math.IsInf(*req.PriceUSD, 0) || *req.PriceUSD < 0 || *req.PriceUSD > 1000000) {
		writeError(w, http.StatusBadRequest, "Price must be between 0 and 1000000 USD")
		return
	}
	if req.MonthlyQuotaBytes > 1<<60 || req.TotalQuotaBytes > 1<<60 {
		writeError(w, http.StatusBadRequest, "Quota value is too large")
		return
	}
	for _, item := range []struct {
		name, value string
		max         int
	}{
		{"public_ipv4_override", req.PublicIPv4Override, 64}, {"public_ipv6_override", req.PublicIPv6Override, 64},
		{"country_override", req.CountryOverride, 80}, {"country_code_override", req.CountryCodeOverride, 2},
		{"region_override", req.RegionOverride, 100}, {"city_override", req.CityOverride, 100},
		{"isp_override", req.ISPOverride, 160}, {"asn_override", req.ASNOverride, 40},
	} {
		if len(item.value) > item.max || strings.ContainsAny(item.value, "\r\n\x00") {
			writeError(w, http.StatusBadRequest, "Invalid field: "+item.name)
			return
		}
	}
	req.PublicIPv4Override = strings.TrimSpace(req.PublicIPv4Override)
	req.PublicIPv6Override = strings.TrimSpace(req.PublicIPv6Override)
	if req.PublicIPv4Override != "" {
		ip := net.ParseIP(req.PublicIPv4Override)
		if ip == nil || ip.To4() == nil {
			writeError(w, http.StatusBadRequest, "IPv4 override is invalid")
			return
		}
	}
	if req.PublicIPv6Override != "" {
		ip := net.ParseIP(req.PublicIPv6Override)
		if ip == nil || ip.To4() != nil {
			writeError(w, http.StatusBadRequest, "IPv6 override is invalid")
			return
		}
	}
	req.CountryOverride = strings.TrimSpace(req.CountryOverride)
	req.CountryCodeOverride = strings.ToUpper(strings.TrimSpace(req.CountryCodeOverride))
	if req.CountryCodeOverride != "" && !countryCodePattern.MatchString(req.CountryCodeOverride) {
		writeError(w, http.StatusBadRequest, "Country code must be a two-letter ISO code")
		return
	}
	req.RegionOverride, req.CityOverride = strings.TrimSpace(req.RegionOverride), strings.TrimSpace(req.CityOverride)
	req.ISPOverride, req.ASNOverride = strings.TrimSpace(req.ISPOverride), strings.TrimSpace(req.ASNOverride)
	if err := a.store.UpdateMetadata(nodeID, store.MetadataUpdate{
		ExpiryDate: req.ExpiryDate, PriceUSD: req.PriceUSD, MonthlyQuotaBytes: req.MonthlyQuotaBytes, TotalQuotaBytes: req.TotalQuotaBytes,
		PublicIPv4Override: req.PublicIPv4Override, PublicIPv6Override: req.PublicIPv6Override, CountryOverride: req.CountryOverride,
		CountryCodeOverride: req.CountryCodeOverride, RegionOverride: req.RegionOverride, CityOverride: req.CityOverride,
		ISPOverride: req.ISPOverride, ASNOverride: req.ASNOverride,
	}); err != nil {
		writeError(w, http.StatusNotFound, "Node not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *App) createNodeToken(w http.ResponseWriter, r *http.Request) {
	if !a.requireAdminMutation(w, r) {
		return
	}
	var req tokenRequest
	if err := decodeJSON(w, r, &req, 8*1024); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request")
		return
	}
	token, err := a.store.CreateToken(req.NodeID, req.NodeName)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"node_id": req.NodeID, "node_name": strings.TrimSpace(req.NodeName),
		"token": token, "server_url": a.cfg.ExternalURL,
	})
}

func (a *App) deleteNode(w http.ResponseWriter, r *http.Request) {
	if !a.requireAdminMutation(w, r) {
		return
	}
	nodeID := r.PathValue("id")
	if !nodeIDPattern.MatchString(nodeID) {
		writeError(w, http.StatusBadRequest, "Invalid node ID")
		return
	}
	if err := a.store.DeleteNode(nodeID); err != nil {
		writeError(w, http.StatusInternalServerError, "Could not update node state")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *App) receiveReport(w http.ResponseWriter, r *http.Request) {
	rawToken := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	if rawToken == "" || rawToken == r.Header.Get("Authorization") {
		writeError(w, http.StatusUnauthorized, "Invalid agent credentials")
		return
	}
	binding, ok := a.store.LookupToken(rawToken)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Invalid agent credentials")
		return
	}
	var report shared.Report
	if err := decodeJSON(w, r, &report, 64*1024); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid report")
		return
	}
	if report.NodeID != binding.NodeID || report.NodeID == "" {
		writeError(w, http.StatusForbidden, "Agent identity does not match token")
		return
	}
	report.DemoData = false
	if report.CPUCount < 1 || report.CPUCount > 4096 {
		writeError(w, http.StatusBadRequest, "Invalid CPU count")
		return
	}
	if len(report.Latencies) > 16 {
		writeError(w, http.StatusBadRequest, "Too many probe results")
		return
	}
	if a.cfg.TrustProxyHeaders && report.PublicIPv4 == "" && report.PublicIPv6 == "" {
		if clientIP := a.trustedProbeClientIP(r); clientIP != "" {
			if parsed := net.ParseIP(clientIP); parsed != nil && parsed.To4() != nil {
				report.PublicIPv4 = parsed.String()
			} else if parsed != nil {
				report.PublicIPv6 = parsed.String()
			}
		}
	}
	if err := a.store.Update(report, binding.NodeName); err != nil {
		log.Printf("failed to persist node report for %s", binding.NodeID)
		writeError(w, http.StatusInternalServerError, "Could not store report")
		return
	}
	if a.cfg.EnableGeoLookup {
		if ip, ok := a.store.MetadataIP(binding.NodeID); ok && ip != "" {
			_ = a.store.ClaimMetadataLookup(binding.NodeID, ip)
			if a.store.ClaimIPMetadataLookup(ip) {
				metadata, found := a.lookupIPMetadata(r.Context(), ip)
				cached := shared.IPMetadata{AttemptedAt: time.Now().UTC(), Success: found}
				if found {
					cached.Country, cached.CountryCode = metadata.Country, metadata.CountryCode
					cached.Region, cached.City = metadata.Region, metadata.City
					cached.ISP, cached.ASN = metadata.Connection.ISP, metadata.Connection.ASN
				}
				if err := a.store.SaveIPMetadata(ip, cached); err != nil {
					log.Printf("could not persist IP metadata cache")
				}
			}
			if cached, exists := a.store.IPMetadata(ip); exists && cached.Success {
				if err := a.store.SetGeoMetadata(binding.NodeID, ip, cached.Country, cached.CountryCode, cached.Region, cached.City, cached.ISP, cached.ASN); err != nil {
					log.Printf("could not apply IP metadata for node %s", binding.NodeID)
				}
			}
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) trustedProbeClientIP(r *http.Request) string {
	candidate := strings.TrimSpace(r.Header.Get("CF-Connecting-IP"))
	ip := net.ParseIP(candidate)
	if ip == nil || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsMulticast() || ip.IsUnspecified() {
		return ""
	}
	return ip.String()
}

func (a *App) lookupIPMetadata(parent context.Context, ip string) (geoLookupResult, bool) {
	if parsed := net.ParseIP(ip); parsed == nil {
		return geoLookupResult{}, false
	}
	ctx, cancel := context.WithTimeout(parent, 2500*time.Millisecond)
	defer cancel()
	client := &http.Client{Timeout: 2500 * time.Millisecond, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://ipwho.is/"+url.PathEscape(ip), nil)
	if err != nil {
		return geoLookupResult{}, false
	}
	resp, err := client.Do(req)
	if err != nil {
		return geoLookupResult{}, false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return geoLookupResult{}, false
	}
	var result geoLookupResult
	if err := json.NewDecoder(io.LimitReader(resp.Body, 16*1024)).Decode(&result); err != nil || !result.Success {
		return geoLookupResult{}, false
	}
	return result, true
}

func (a *App) requireAdminMutation(w http.ResponseWriter, r *http.Request) bool {
	if !a.validMutationOrigin(r) {
		writeError(w, http.StatusForbidden, "Request origin rejected")
		return false
	}
	if _, ok := a.authenticate(r); !ok {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return false
	}
	return true
}

func (a *App) validMutationOrigin(r *http.Request) bool {
	if r.Header.Get("X-Polaris-Request") != "1" {
		return false
	}
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return false
	}
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return false
	}
	external, err := url.Parse(a.cfg.ExternalURL)
	if err != nil {
		return false
	}
	return strings.EqualFold(parsed.Host, external.Host)
}

func (a *App) verifyTurnstile(ctx context.Context, responseToken, remoteIP string) bool {
	if a.cfg.TurnstileSecret == "" || len(responseToken) > 2048 {
		return false
	}
	form := url.Values{}
	form.Set("secret", a.cfg.TurnstileSecret)
	form.Set("response", responseToken)
	if net.ParseIP(remoteIP) != nil {
		form.Set("remoteip", remoteIP)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://challenges.cloudflare.com/turnstile/v0/siteverify", strings.NewReader(form.Encode()))
	if err != nil {
		return false
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := a.turnstile.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}
	var result turnstileResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 32*1024)).Decode(&result); err != nil || !result.Success {
		return false
	}
	if result.Action != "login" || !strings.EqualFold(result.Hostname, a.cfg.AllowedHostname) {
		return false
	}
	return true
}

func (a *App) createSession(username string) (string, error) {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	sessionID := base64.RawURLEncoding.EncodeToString(nonce)
	expires := time.Now().Add(sessionLifetime).Unix()
	payload := fmt.Sprintf("%s|%d|%s", username, expires, sessionID)
	encoded := base64.RawURLEncoding.EncodeToString([]byte(payload))
	mac := hmac.New(sha256.New, a.cfg.SessionSecret)
	_, _ = mac.Write([]byte(encoded))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	a.sessionsMu.Lock()
	defer a.sessionsMu.Unlock()
	if a.sessions == nil {
		a.sessions = make(map[string]int64)
	}
	for id, expiry := range a.sessions {
		if expiry < time.Now().Unix() {
			delete(a.sessions, id)
		}
	}
	if len(a.sessions) >= 4096 {
		return "", errors.New("session capacity reached")
	}
	a.sessions[sessionID] = expires
	return encoded + "." + signature, nil
}

func (a *App) authenticate(r *http.Request) (string, bool) {
	username, sessionID, expires, ok := a.readSession(r)
	if !ok {
		return "", false
	}
	a.sessionsMu.Lock()
	registeredExpiry, exists := a.sessions[sessionID]
	a.sessionsMu.Unlock()
	if !exists || registeredExpiry != expires {
		return "", false
	}
	return username, true
}

func (a *App) revokeSession(r *http.Request) {
	_, sessionID, _, ok := a.readSession(r)
	if !ok {
		return
	}
	a.sessionsMu.Lock()
	delete(a.sessions, sessionID)
	a.sessionsMu.Unlock()
}

func (a *App) readSession(r *http.Request) (string, string, int64, bool) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || len(cookie.Value) > 2048 {
		return "", "", 0, false
	}
	parts := strings.Split(cookie.Value, ".")
	if len(parts) != 2 {
		return "", "", 0, false
	}
	mac := hmac.New(sha256.New, a.cfg.SessionSecret)
	_, _ = mac.Write([]byte(parts[0]))
	expected, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(mac.Sum(nil), expected) {
		return "", "", 0, false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", "", 0, false
	}
	fields := strings.Split(string(payload), "|")
	if len(fields) != 3 || fields[0] != a.cfg.AdminUser {
		return "", "", 0, false
	}
	var expires int64
	if _, err := fmt.Sscan(fields[1], &expires); err != nil || expires < time.Now().Unix() || expires > time.Now().Add(sessionLifetime+time.Minute).Unix() {
		return "", "", 0, false
	}
	if fields[2] == "" {
		return "", "", 0, false
	}
	return fields[0], fields[2], expires, true
}

func (a *App) clientIP(r *http.Request) string {
	if a.cfg.TrustProxyHeaders {
		candidate := strings.TrimSpace(r.Header.Get("CF-Connecting-IP"))
		if ip := net.ParseIP(candidate); ip != nil {
			return ip.String()
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && net.ParseIP(host) != nil {
		return host
	}
	return "unknown"
}

func (l *loginLimiter) Allow(ip string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	entry, exists := l.entries[ip]
	if now.After(entry.ResetAt) {
		entry = loginBucket{Count: 0, ResetAt: now.Add(15 * time.Minute)}
		exists = false
	}
	if entry.Count >= 8 {
		l.entries[ip] = entry
		return false
	}
	if !exists && len(l.entries) >= 2048 {
		for key, value := range l.entries {
			if now.After(value.ResetAt) {
				delete(l.entries, key)
			}
		}
		if len(l.entries) >= 2048 {
			return false
		}
	}
	entry.Count++
	if entry.ResetAt.IsZero() {
		entry.ResetAt = now.Add(15 * time.Minute)
	}
	l.entries[ip] = entry
	return true
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any, maxBytes int64) error {
	if r.Method != http.MethodPost && r.Method != http.MethodPatch {
		return errors.New("invalid method")
	}
	if !strings.HasPrefix(strings.ToLower(r.Header.Get("Content-Type")), "application/json") {
		return errors.New("invalid content type")
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("multiple JSON values")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, apiError{Error: message})
}
