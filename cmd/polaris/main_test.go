package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pwenxiang51-wq/polaris/internal/store"
)

func TestSessionRoundTripTamperingAndLogoutRevocation(t *testing.T) {
	app := &App{cfg: Config{AdminUser: "admin", SessionSecret: []byte(strings.Repeat("s", 64)), ExternalURL: "https://status.example.com"}}
	value, err := app.createSession("admin")
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("GET", "https://status.example.com/api/me", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: value})
	if _, ok := app.authenticate(req); !ok {
		t.Fatal("valid session cookie was rejected")
	}
	tampered := httptest.NewRequest("GET", "https://status.example.com/api/me", nil)
	tampered.AddCookie(&http.Cookie{Name: sessionCookieName, Value: value + "x"})
	if _, ok := app.authenticate(tampered); ok {
		t.Fatal("tampered session cookie was accepted")
	}

	logout := httptest.NewRequest("POST", "https://status.example.com/api/logout", nil)
	logout.Header.Set("Origin", "https://status.example.com")
	logout.Header.Set("X-Polaris-Request", "1")
	logout.AddCookie(&http.Cookie{Name: sessionCookieName, Value: value})
	response := httptest.NewRecorder()
	app.logout(response, logout)
	if response.Code != http.StatusOK {
		t.Fatalf("logout failed: %d %s", response.Code, response.Body.String())
	}
	if _, ok := app.authenticate(req); ok {
		t.Fatal("session remained valid after logout")
	}
}

func TestLoginLimiter(t *testing.T) {
	limiter := &loginLimiter{entries: make(map[string]loginBucket)}
	now := time.Now()
	for attempt := 0; attempt < 8; attempt++ {
		if !limiter.Allow("192.0.2.1", now) {
			t.Fatalf("login attempt %d was unexpectedly blocked", attempt+1)
		}
	}
	if limiter.Allow("192.0.2.1", now) {
		t.Fatal("ninth login attempt was not blocked")
	}
	if !limiter.Allow("192.0.2.1", now.Add(16*time.Minute)) {
		t.Fatal("limiter did not reset after the timeout")
	}
}

func TestMutationOriginRequiresExpectedOriginAndHeader(t *testing.T) {
	app := &App{cfg: Config{ExternalURL: "https://status.example.com"}}
	req := httptest.NewRequest("POST", "https://status.example.com/api/nodes/token", nil)
	req.Header.Set("Origin", "https://status.example.com")
	req.Header.Set("X-Polaris-Request", "1")
	if !app.validMutationOrigin(req) {
		t.Fatal("same-origin request was rejected")
	}
	req.Header.Set("Origin", "https://attacker.example")
	if app.validMutationOrigin(req) {
		t.Fatal("cross-origin request was accepted")
	}
}

func TestProbeReportUsesTokenBoundNodeIdentity(t *testing.T) {
	st, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	token, err := st.CreateToken("node-01", "Node One")
	if err != nil {
		t.Fatal(err)
	}
	app := &App{store: st}
	body := `{"node_id":"node-01","cpu_count":2,"cpu_percent":20,"mem_total_bytes":1000,"mem_used_bytes":400,"disk_total_bytes":2000,"disk_used_bytes":1000,"rx_bytes":1234,"tx_bytes":5678,"latencies":[]}`
	req := httptest.NewRequest("POST", "https://status.example.com/api/v1/report", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	app.receiveReport(response, req)
	if response.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", response.Code, response.Body.String())
	}
	nodes := st.Nodes()
	if len(nodes) != 1 || nodes[0].Name != "Node One" || nodes[0].MemPercent != 40 {
		t.Fatalf("report was not stored with token-bound name and derived metrics: %#v", nodes)
	}

	mismatchBody := strings.Replace(body, `"node-01"`, `"other-node"`, 1)
	mismatch := httptest.NewRequest("POST", "https://status.example.com/api/v1/report", strings.NewReader(mismatchBody))
	mismatch.Header.Set("Content-Type", "application/json")
	mismatch.Header.Set("Authorization", "Bearer "+token)
	mismatchResponse := httptest.NewRecorder()
	app.receiveReport(mismatchResponse, mismatch)
	if mismatchResponse.Code != http.StatusForbidden {
		t.Fatalf("expected spoofed node ID to be rejected with 403, got %d", mismatchResponse.Code)
	}
}

func TestPublicConfigDoesNotExposeTurnstileSecret(t *testing.T) {
	app := &App{cfg: Config{TurnstileSiteKey: "public-site-key", TurnstileSecret: "private-secret"}}
	req := httptest.NewRequest("GET", "https://status.example.com/api/public-config", nil)
	response := httptest.NewRecorder()
	app.publicConfig(response, req)
	if strings.Contains(response.Body.String(), "private-secret") {
		t.Fatal("public configuration exposed the Turnstile secret")
	}
	if !strings.Contains(response.Body.String(), "public-site-key") {
		t.Fatal("public configuration omitted the Turnstile site key")
	}
}

func TestIndexHasSecurityHeaders(t *testing.T) {
	app := &App{}
	handler := app.securityHeaders(http.HandlerFunc(app.index))
	req := httptest.NewRequest("GET", "https://status.example.com/", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	if response.Code != http.StatusOK {
		t.Fatalf("expected page status 200, got %d", response.Code)
	}
	if response.Header().Get("X-Content-Type-Options") != "nosniff" || response.Header().Get("X-Frame-Options") != "DENY" {
		t.Fatal("essential browser security headers are missing")
	}
	if !strings.Contains(response.Header().Get("Content-Security-Policy"), "frame-ancestors 'none'") {
		t.Fatal("frame restrictions are missing from CSP")
	}
}

func TestSecretEnvPrefersMountedFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "secret-value")
	if err := os.WriteFile(path, []byte("mounted-secret\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("POLARIS_TEST_SECRET_FILE", path)
	t.Setenv("POLARIS_TEST_SECRET", "environment-secret")
	value, err := secretEnv("POLARIS_TEST_SECRET")
	if err != nil {
		t.Fatal(err)
	}
	if value != "mounted-secret" {
		t.Fatalf("expected mounted file to take precedence, got %q", value)
	}
}

func TestReportUsesTrustedCloudflareIPOnlyAsFallback(t *testing.T) {
	st, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	token, err := st.CreateToken("edge-01", "Edge")
	if err != nil {
		t.Fatal(err)
	}
	app := &App{cfg: Config{TrustProxyHeaders: true, EnableGeoLookup: false}, store: st}
	body := `{"node_id":"edge-01","cpu_count":1,"mem_total_bytes":100,"disk_total_bytes":100,"latencies":[]}`
	req := httptest.NewRequest("POST", "https://status.example.com/api/v1/report", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("CF-Connecting-IP", "8.8.4.4")
	response := httptest.NewRecorder()
	app.receiveReport(response, req)
	if response.Code != http.StatusNoContent {
		t.Fatalf("report failed: %d %s", response.Code, response.Body.String())
	}
	node, ok := st.Node("edge-01")
	if !ok || node.PublicIPv4 != "8.8.4.4" {
		t.Fatalf("trusted report IP was not used as fallback: %#v", node)
	}
}

func TestReportDoesNotReplaceAgentDetectedIPWithProxyHeader(t *testing.T) {
	st, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	token, err := st.CreateToken("edge-02", "Edge")
	if err != nil {
		t.Fatal(err)
	}
	app := &App{cfg: Config{TrustProxyHeaders: true, EnableGeoLookup: false}, store: st}
	body := `{"node_id":"edge-02","cpu_count":1,"public_ipv4":"8.8.8.8","mem_total_bytes":100,"disk_total_bytes":100,"latencies":[]}`
	req := httptest.NewRequest("POST", "https://status.example.com/api/v1/report", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("CF-Connecting-IP", "8.8.4.4")
	response := httptest.NewRecorder()
	app.receiveReport(response, req)
	if response.Code != http.StatusNoContent {
		t.Fatalf("report failed: %d %s", response.Code, response.Body.String())
	}
	node, ok := st.Node("edge-02")
	if !ok || node.PublicIPv4 != "8.8.8.8" {
		t.Fatalf("proxy header overrode detected agent IP: %#v", node)
	}
}

func TestUpdateNodeMetadataEndpoint(t *testing.T) {
	st, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.CreateToken("managed-01", "Managed node"); err != nil {
		t.Fatal(err)
	}
	app := &App{cfg: Config{AdminUser: "admin", SessionSecret: []byte(strings.Repeat("s", 64)), ExternalURL: "https://status.example.com"}, store: st}
	session, err := app.createSession("admin")
	if err != nil {
		t.Fatal(err)
	}
	body := `{"expiry_date":"2030-01-01","price_usd":3.5,"monthly_quota_bytes":1073741824,"total_quota_bytes":10737418240,"public_ipv4_override":"9.9.9.9","country_override":"Testland","country_code_override":"US"}`
	req := httptest.NewRequest("PATCH", "https://status.example.com/api/nodes/managed-01/metadata", strings.NewReader(body))
	req.SetPathValue("id", "managed-01")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://status.example.com")
	req.Header.Set("X-Polaris-Request", "1")
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: session})
	response := httptest.NewRecorder()
	app.updateNodeMetadata(response, req)
	if response.Code != http.StatusOK {
		t.Fatalf("metadata patch failed: %d %s", response.Code, response.Body.String())
	}
	node, ok := st.Node("managed-01")
	if !ok || node.ExpiryDate != "2030-01-01" || node.PriceUSD == nil || *node.PriceUSD != 3.5 || node.MonthlyQuotaBytes != 1073741824 || node.TotalQuotaBytes != 10737418240 || node.PublicIPv4 != "9.9.9.9" || node.Country != "Testland" {
		t.Fatalf("metadata was not saved: %#v", node)
	}
}
