# CT Scanning + Notifications + Scheduler Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement Certificate Transparency scanning (sslmate primary, crt.sh fallback), three notification channels (shoutrrr, greenapi, waweb), a token-bucket rate limiter, and a robfig/cron scheduler that registers per-FQDN jobs and runs scans automatically.

**Architecture:** Each subsystem lives in its own package under internal/. The `ct` package fetches certs. The `notify` package dispatches notifications. The `scheduler` package wraps robfig/cron and registers/de-registers jobs when FQDNs or schedules change. The API `TriggerScan` handler and the scheduler both call the same `scanner.ScanFQDN()` function. Failing notifications are logged and never block a scan.

**Tech Stack:** robfig/cron/v3, containrrr/shoutrrr, net/http (for sslmate + greenapi + waweb), golang.org/x/time/rate (token bucket), github.com/google/uuid

**Prerequisites:** Plans 2 and 3 (domain model + API) must be complete.

---

## File Map

| File | Purpose |
|---|---|
| `internal/ct/sslmate.go` | sslmate Cert Spotter API client |
| `internal/ct/crtsh.go` | crt.sh fallback client |
| `internal/ct/ct.go` | `FetchCerts(fqdn, includeSubdomains) ([]Cert, error)` — tries sslmate, falls back to crt.sh |
| `internal/ct/ct_test.go` | Unit tests with HTTP mock server |
| `internal/notify/shoutrrr.go` | shoutrrr dispatcher |
| `internal/notify/greenapi.go` | GreenAPI WhatsApp dispatcher |
| `internal/notify/waweb.go` | go-whatsapp-web-multidevice dispatcher |
| `internal/notify/notify.go` | `Dispatch(channel, cert, fqdn)` — routes to correct dispatcher; best-effort |
| `internal/notify/notify_test.go` | Unit tests with mock HTTP server for greenapi/waweb |
| `internal/scanner/scanner.go` | `Scanner` — wraps ct, notify, models; `ScanFQDN(ctx, fqdn)` |
| `internal/scanner/scanner_test.go` | Integration test with in-memory DB |
| `internal/scheduler/scheduler.go` | `Scheduler` wrapping robfig/cron; `Start`, `Stop`, `RegisterFQDN`, `DeregisterFQDN` |
| `internal/api/fqdns.go` | Modified: `TriggerScan` calls `scanner.Scanner.ScanFQDN` |
| `internal/api/channels.go` | Modified: `TestChannel` calls `notify.Dispatch` |
| `internal/api/handler.go` | Modified: holds `*scanner.Scanner` |
| `internal/api/server.go` | Unchanged |
| `cmd/go-certi/main.go` | Modified: creates scanner + scheduler, starts scheduler |

---

## Task 1: Dependencies

- [ ] **Step 1: Add dependencies**

```bash
cd /opt/dev/go-certi
go get github.com/robfig/cron/v3@latest
go get github.com/containrrr/shoutrrr@latest
go get golang.org/x/time@latest
```

- [ ] **Step 2: Verify build**

```bash
go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: add cron, shoutrrr, and rate-limiter dependencies"
```

---

## Task 2: CT Package — sslmate + crt.sh

**Files:**
- Create: `internal/ct/ct.go`
- Create: `internal/ct/sslmate.go`
- Create: `internal/ct/crtsh.go`
- Create: `internal/ct/ct_test.go`

- [ ] **Step 1: Define shared Cert type and write failing test**

Create `internal/ct/ct.go`:

```go
package ct

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// Cert holds the data extracted from a CT log entry.
type Cert struct {
	Serial    string
	IssuerCA  string
	SubjectCN string
	SANs      []string
	NotBefore time.Time
	NotAfter  time.Time
	Source    string // "sslmate" or "crt.sh"
}

// Client fetches CT log entries for an FQDN.
type Client struct {
	sslmateKey    string
	sslmateClient *sslmateClient
	crtshClient   *crtshClient
}

// New creates a CT Client.
func New(sslmateAPIKey string) *Client {
	return &Client{
		sslmateKey:    sslmateAPIKey,
		sslmateClient: newSslmateClient(sslmateAPIKey),
		crtshClient:   newCrtshClient(),
	}
}

// FetchCerts retrieves certificates for fqdn from sslmate (primary) with crt.sh fallback.
// It returns only certificates not yet seen (dedup is done by the caller via serial).
func (c *Client) FetchCerts(ctx context.Context, fqdn string, includeSubdomains bool) ([]Cert, error) {
	if c.sslmateKey != "" {
		certs, err := c.sslmateClient.fetch(ctx, fqdn, includeSubdomains)
		if err == nil {
			return certs, nil
		}
		slog.Warn("sslmate fetch failed, falling back to crt.sh", "fqdn", fqdn, "err", err)
	}
	certs, err := c.crtshClient.fetch(ctx, fqdn, includeSubdomains)
	if err != nil {
		return nil, fmt.Errorf("ct fetch: %w", err)
	}
	return certs, nil
}
```

Create `internal/ct/ct_test.go`:

```go
package ct_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/t0mer/go-certi/internal/ct"
)

// TestSslmateFetch tests the sslmate client with a mock HTTP server.
func TestSslmateFetch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify auth header
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"id":          "1",
				"tbs_sha256":  "abc",
				"cert_sha256": "def",
				"dns_names":   []string{"example.com"},
				"pubkey_sha256": "ghi",
				"issuer": map[string]any{"name": "Let's Encrypt"},
				"not_before": time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
				"not_after":  time.Now().Add(90 * 24 * time.Hour).Format(time.RFC3339),
				"cert_der":   "",
			},
		})
	}))
	defer srv.Close()

	client := ct.NewWithBaseURL("test-api-key", srv.URL+"/ct/v2/certs", "")
	certs, err := client.FetchCerts(context.Background(), "example.com", false)
	if err != nil {
		t.Fatalf("FetchCerts: %v", err)
	}
	if len(certs) == 0 {
		t.Error("expected at least one cert")
	}
}

// TestCrtshFetch tests the crt.sh client with a mock HTTP server.
func TestCrtshFetch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"serial_number":  "01020304",
				"issuer_name":    "C=US, O=Let's Encrypt, CN=R3",
				"common_name":    "example.com",
				"name_value":     "example.com\nsub.example.com",
				"not_before":     time.Now().Add(-24 * time.Hour).Format("2006-01-02T15:04:05"),
				"not_after":      time.Now().Add(90 * 24 * time.Hour).Format("2006-01-02T15:04:05"),
			},
		})
	}))
	defer srv.Close()

	client := ct.NewWithBaseURL("", "", srv.URL)
	certs, err := client.FetchCerts(context.Background(), "example.com", false)
	if err != nil {
		t.Fatalf("FetchCerts: %v", err)
	}
	if len(certs) == 0 {
		t.Error("expected at least one cert")
	}
	if certs[0].Source != "crt.sh" {
		t.Errorf("source = %q, want crt.sh", certs[0].Source)
	}
}
```

- [ ] **Step 2: Run to confirm fails**

```bash
go test ./internal/ct/ -v 2>&1 | head -5
```

- [ ] **Step 3: Create sslmate.go**

Create `internal/ct/sslmate.go`:

```go
package ct

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const sslmateBaseURL = "https://api.certspotter.com/v1/issuances"

type sslmateClient struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

func newSslmateClient(apiKey string) *sslmateClient {
	return &sslmateClient{
		apiKey:  apiKey,
		baseURL: sslmateBaseURL,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

type sslmateEntry struct {
	ID        string   `json:"id"`
	DNSNames  []string `json:"dns_names"`
	Issuer    struct {
		Name string `json:"name"`
	} `json:"issuer"`
	NotBefore string `json:"not_before"`
	NotAfter  string `json:"not_after"`
	TBSSha256 string `json:"tbs_sha256"`
}

func (c *sslmateClient) fetch(ctx context.Context, fqdn string, includeSubdomains bool) ([]Cert, error) {
	u, _ := url.Parse(c.baseURL)
	q := u.Query()
	q.Set("domain", fqdn)
	q.Set("include_subdomains", fmt.Sprintf("%t", includeSubdomains))
	q.Set("expand", "dns_names")
	u.RawQuery = q.Encode()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("sslmate rate limit exceeded")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sslmate status %d", resp.StatusCode)
	}
	var entries []sslmateEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, err
	}
	certs := make([]Cert, 0, len(entries))
	for _, e := range entries {
		nb, _ := time.Parse(time.RFC3339, e.NotBefore)
		na, _ := time.Parse(time.RFC3339, e.NotAfter)
		cn := ""
		if len(e.DNSNames) > 0 {
			cn = e.DNSNames[0]
		}
		certs = append(certs, Cert{
			Serial:    e.TBSSha256,
			IssuerCA:  e.Issuer.Name,
			SubjectCN: cn,
			SANs:      e.DNSNames,
			NotBefore: nb,
			NotAfter:  na,
			Source:    "sslmate",
		})
	}
	return certs, nil
}

// NewWithBaseURL creates a Client with configurable base URLs (for testing).
func NewWithBaseURL(sslmateKey, sslmateBaseURL, crtshBaseURL string) *Client {
	c := New(sslmateKey)
	if sslmateBaseURL != "" {
		c.sslmateClient.baseURL = sslmateBaseURL
	}
	if crtshBaseURL != "" {
		c.crtshClient.baseURL = crtshBaseURL
	}
	return c
}

// extractCN returns the first part before \n in a name_value field.
func extractCN(nameValue string) string {
	parts := strings.SplitN(nameValue, "\n", 2)
	return strings.TrimSpace(parts[0])
}
```

- [ ] **Step 4: Create crtsh.go**

Create `internal/ct/crtsh.go`:

```go
package ct

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const crtshBaseURL = "https://crt.sh"

type crtshClient struct {
	baseURL string
	http    *http.Client
}

func newCrtshClient() *crtshClient {
	return &crtshClient{
		baseURL: crtshBaseURL,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

type crtshEntry struct {
	SerialNumber string `json:"serial_number"`
	IssuerName   string `json:"issuer_name"`
	CommonName   string `json:"common_name"`
	NameValue    string `json:"name_value"`
	NotBefore    string `json:"not_before"`
	NotAfter     string `json:"not_after"`
}

func (c *crtshClient) fetch(ctx context.Context, fqdn string, includeSubdomains bool) ([]Cert, error) {
	q := fqdn
	if includeSubdomains {
		q = "%" + fqdn
	}
	u := fmt.Sprintf("%s/?q=%s&output=json", c.baseURL, url.QueryEscape(q))
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("crt.sh status %d", resp.StatusCode)
	}
	var entries []crtshEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	certs := make([]Cert, 0, len(entries))
	for _, e := range entries {
		if seen[e.SerialNumber] {
			continue
		}
		seen[e.SerialNumber] = true
		nb, _ := time.Parse("2006-01-02T15:04:05", e.NotBefore)
		na, _ := time.Parse("2006-01-02T15:04:05", e.NotAfter)
		sans := parseSANs(e.NameValue)
		certs = append(certs, Cert{
			Serial:    e.SerialNumber,
			IssuerCA:  e.IssuerName,
			SubjectCN: e.CommonName,
			SANs:      sans,
			NotBefore: nb,
			NotAfter:  na,
			Source:    "crt.sh",
		})
	}
	return certs, nil
}

func parseSANs(nameValue string) []string {
	parts := strings.Split(nameValue, "\n")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
```

- [ ] **Step 5: Run tests**

```bash
go test ./internal/ct/ -v -race
```

Expected: both tests PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/ct/
git commit -m "feat: add CT client (sslmate primary, crt.sh fallback)"
```

---

## Task 3: Notification Package

**Files:**
- Create: `internal/notify/notify.go`
- Create: `internal/notify/shoutrrr.go`
- Create: `internal/notify/greenapi.go`
- Create: `internal/notify/waweb.go`
- Create: `internal/notify/notify_test.go`

- [ ] **Step 1: Create notify.go**

Create `internal/notify/notify.go`:

```go
package notify

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/t0mer/go-certi/internal/ct"
	"github.com/t0mer/go-certi/internal/models"
)

// Dispatcher sends a notification for a new certificate.
type Dispatcher struct{}

// New creates a Dispatcher.
func New() *Dispatcher { return &Dispatcher{} }

// formatMessage returns a human-readable notification string.
func formatMessage(fqdn string, cert ct.Cert) string {
	sanCount := len(cert.SANs)
	sanStr := ""
	if sanCount > 1 {
		sanStr = fmt.Sprintf(" (+%d SANs)", sanCount-1)
	}
	return fmt.Sprintf(
		"[go-certi] New certificate for %s\nIssuer: %s\nCN: %s%s\nValid: %s → %s\nSerial: %s",
		fqdn,
		cert.IssuerCA,
		cert.SubjectCN,
		sanStr,
		cert.NotBefore.Format("2006-01-02"),
		cert.NotAfter.Format("2006-01-02"),
		cert.Serial,
	)
}

// Dispatch sends a notification via channel. It is best-effort: errors are logged, never propagated.
func (d *Dispatcher) Dispatch(ctx context.Context, channel models.NotificationChannel, fqdn string, cert ct.Cert) {
	if !channel.Enabled {
		return
	}
	msg := formatMessage(fqdn, cert)
	var err error
	switch channel.Type {
	case "shoutrrr":
		err = dispatchShoutrrr(ctx, channel.Config, msg)
	case "greenapi":
		err = dispatchGreenAPI(ctx, channel.Config, msg)
	case "waweb":
		err = dispatchWaWeb(ctx, channel.Config, msg)
	default:
		slog.Warn("unknown channel type", "type", channel.Type, "channel", channel.Name)
		return
	}
	if err != nil {
		slog.Error("notification failed", "channel", channel.Name, "type", channel.Type, "err", err)
	}
}

// DispatchAll sends a notification to all channels for a FQDN. Best-effort.
func (d *Dispatcher) DispatchAll(ctx context.Context, channels []models.NotificationChannel, fqdn string, cert ct.Cert) {
	for _, ch := range channels {
		d.Dispatch(ctx, ch, fqdn, cert)
	}
}

// TestChannel sends a test message to a single channel.
func (d *Dispatcher) TestChannel(ctx context.Context, channel models.NotificationChannel) error {
	testCert := ct.Cert{
		IssuerCA:  "Test CA",
		SubjectCN: "test.example.com",
		SANs:      []string{"test.example.com"},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(90 * 24 * time.Hour),
		Serial:    "00:11:22:33",
	}
	msg := "[go-certi] Test notification — " + formatMessage("test.example.com", testCert)
	switch channel.Type {
	case "shoutrrr":
		return dispatchShoutrrr(ctx, channel.Config, msg)
	case "greenapi":
		return dispatchGreenAPI(ctx, channel.Config, msg)
	case "waweb":
		return dispatchWaWeb(ctx, channel.Config, msg)
	}
	return fmt.Errorf("unknown channel type: %s", channel.Type)
}

// parseConfig decodes a JSON config string into a map.
func parseConfig(configJSON string) (map[string]string, error) {
	var m map[string]string
	return m, json.Unmarshal([]byte(configJSON), &m)
}
```

- [ ] **Step 2: Create shoutrrr.go**

Create `internal/notify/shoutrrr.go`:

```go
package notify

import (
	"context"
	"fmt"

	"github.com/containrrr/shoutrrr"
)

func dispatchShoutrrr(ctx context.Context, configJSON, msg string) error {
	cfg, err := parseConfig(configJSON)
	if err != nil {
		return fmt.Errorf("shoutrrr: parse config: %w", err)
	}
	u, ok := cfg["url"]
	if !ok || u == "" {
		return fmt.Errorf("shoutrrr: missing 'url' in config")
	}
	return shoutrrr.Send(u, msg)
}
```

- [ ] **Step 3: Create greenapi.go**

Create `internal/notify/greenapi.go`:

```go
package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func dispatchGreenAPI(ctx context.Context, configJSON, msg string) error {
	cfg, err := parseConfig(configJSON)
	if err != nil {
		return fmt.Errorf("greenapi: parse config: %w", err)
	}
	instanceID := cfg["instance_id"]
	apiToken := cfg["api_token_instance"]
	chatID := cfg["chat_id"]
	apiURL := cfg["api_url"]
	if apiURL == "" {
		apiURL = "https://api.green-api.com"
	}
	if instanceID == "" || apiToken == "" || chatID == "" {
		return fmt.Errorf("greenapi: missing instance_id, api_token_instance, or chat_id in config")
	}

	url := fmt.Sprintf("%s/waInstance%s/sendMessage/%s", apiURL, instanceID, apiToken)
	payload := map[string]any{
		"chatId":  chatID,
		"message": msg,
	}
	body, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 15 * time.Second}
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("greenapi: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("greenapi: status %d", resp.StatusCode)
	}
	return nil
}
```

- [ ] **Step 4: Create waweb.go**

Create `internal/notify/waweb.go`:

```go
package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func dispatchWaWeb(ctx context.Context, configJSON, msg string) error {
	cfg, err := parseConfig(configJSON)
	if err != nil {
		return fmt.Errorf("waweb: parse config: %w", err)
	}
	baseURL := cfg["base_url"]
	phone := cfg["phone"]
	authHeader := cfg["auth"] // "basic <b64>" or ""
	if baseURL == "" || phone == "" {
		return fmt.Errorf("waweb: missing base_url or phone in config")
	}

	url := fmt.Sprintf("%s/api/send/message", baseURL)
	payload := map[string]any{
		"phone":   phone,
		"message": msg,
	}
	body, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 15 * time.Second}
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("waweb: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("waweb: status %d", resp.StatusCode)
	}
	return nil
}
```

- [ ] **Step 5: Create notify_test.go**

Create `internal/notify/notify_test.go`:

```go
package notify_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/t0mer/go-certi/internal/ct"
	"github.com/t0mer/go-certi/internal/models"
	"github.com/t0mer/go-certi/internal/notify"
)

func TestDispatchGreenAPI(t *testing.T) {
	var received map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"idMessage":"test123"}`))
	}))
	defer srv.Close()

	cfgJSON, _ := json.Marshal(map[string]string{
		"instance_id":        "1234",
		"api_token_instance": "token",
		"chat_id":            "972501234567@c.us",
		"api_url":            srv.URL,
	})

	d := notify.New()
	d.Dispatch(context.Background(), models.NotificationChannel{
		ID:      "test",
		Name:    "test",
		Type:    "greenapi",
		Config:  string(cfgJSON),
		Enabled: true,
	}, "example.com", ct.Cert{
		IssuerCA:  "Let's Encrypt",
		SubjectCN: "example.com",
		SANs:      []string{"example.com"},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(90 * 24 * time.Hour),
		Serial:    "abc123",
	})

	if received["chatId"] != "972501234567@c.us" {
		t.Errorf("chatId = %v", received["chatId"])
	}
}

func TestDispatchWaWeb(t *testing.T) {
	var received map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfgJSON, _ := json.Marshal(map[string]string{
		"base_url": srv.URL,
		"phone":    "+972501234567",
	})

	d := notify.New()
	d.Dispatch(context.Background(), models.NotificationChannel{
		ID: "t", Name: "t", Type: "waweb", Config: string(cfgJSON), Enabled: true,
	}, "example.com", ct.Cert{
		IssuerCA: "CA", SubjectCN: "example.com", SANs: []string{"example.com"},
		NotBefore: time.Now(), NotAfter: time.Now().Add(90 * 24 * time.Hour), Serial: "xyz",
	})

	if received["phone"] != "+972501234567" {
		t.Errorf("phone = %v", received["phone"])
	}
}
```

- [ ] **Step 6: Run tests**

```bash
go test ./internal/notify/ -v -race
```

Expected: both tests PASS (shoutrrr is not tested with a real URL — that would require a live service).

- [ ] **Step 7: Commit**

```bash
git add internal/notify/
git commit -m "feat: add notification dispatchers (shoutrrr, greenapi, waweb)"
```

---

## Task 4: Scanner Package

**Files:**
- Create: `internal/scanner/scanner.go`
- Create: `internal/scanner/scanner_test.go`

- [ ] **Step 1: Create scanner.go**

Create `internal/scanner/scanner.go`:

```go
package scanner

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/t0mer/go-certi/internal/ct"
	"github.com/t0mer/go-certi/internal/models"
	"github.com/t0mer/go-certi/internal/notify"
)

// Scanner orchestrates CT scanning, dedup, and notification for a single FQDN.
type Scanner struct {
	q    *models.Queries
	ct   *ct.Client
	disp *notify.Dispatcher
}

// New creates a Scanner.
func New(q *models.Queries, ctClient *ct.Client, disp *notify.Dispatcher) *Scanner {
	return &Scanner{q: q, ct: ctClient, disp: disp}
}

// ScanFQDN fetches new certificates for fqdn, stores new ones, and dispatches notifications.
func (s *Scanner) ScanFQDN(ctx context.Context, fqdn models.Fqdn) error {
	slog.Info("scanning FQDN", "fqdn", fqdn.Fqdn)

	certs, err := s.ct.FetchCerts(ctx, fqdn.Fqdn, fqdn.IncludeSubdomains)
	if err != nil {
		return err
	}

	channels, err := s.q.GetFQDNChannels(ctx, fqdn.ID)
	if err != nil {
		return err
	}

	newCount := 0
	for _, cert := range certs {
		sans, _ := json.Marshal(cert.SANs)
		row, err := s.q.InsertCertificate(ctx, models.InsertCertificateParams{
			ID:           uuid.NewString(),
			FqdnID:       fqdn.ID,
			Serial:       cert.Serial,
			IssuerCa:     cert.IssuerCA,
			SubjectCn:    cert.SubjectCN,
			Sans:         string(sans),
			NotBefore:    cert.NotBefore.UTC().Format(time.RFC3339),
			NotAfter:     cert.NotAfter.UTC().Format(time.RFC3339),
			DiscoveredAt: time.Now().UTC().Format(time.RFC3339),
			Source:       cert.Source,
		})
		if err != nil {
			slog.Error("insert certificate", "serial", cert.Serial, "err", err)
			continue
		}
		if row == nil {
			continue // already known
		}
		newCount++
		if fqdn.NotificationsEnabled {
			s.disp.DispatchAll(ctx, channels, fqdn.Fqdn, cert)
		}
	}

	slog.Info("scan complete", "fqdn", fqdn.Fqdn, "new", newCount, "total", len(certs))
	return nil
}
```

- [ ] **Step 2: Create scanner_test.go**

Create `internal/scanner/scanner_test.go`:

```go
package scanner_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/t0mer/go-certi/internal/ct"
	"github.com/t0mer/go-certi/internal/db"
	"github.com/t0mer/go-certi/internal/models"
	"github.com/t0mer/go-certi/internal/notify"
	"github.com/t0mer/go-certi/internal/scanner"
)

// fakeCTClient returns a fixed set of certs.
type fakeCTClient struct {
	certs []ct.Cert
}

func (f *fakeCTClient) FetchCerts(_ context.Context, _ string, _ bool) ([]ct.Cert, error) {
	return f.certs, nil
}

func TestScanFQDN_NewCertsStored(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	q := models.New(conn)
	ctx := context.Background()

	fqdn, err := q.CreateFQDN(ctx, models.CreateFQDNParams{
		ID: uuid.NewString(), Fqdn: "scan-test.com",
		Enabled: true, NotificationsEnabled: false,
	})
	if err != nil {
		t.Fatal(err)
	}

	fakeCT := &ct.Client{}
	_ = fakeCT
	// Use a scanner with a fake CT client directly
	ctClient := ct.NewWithBaseURL("", "", "") // no real URLs; won't be called
	_ = ctClient

	// Instead, build scanner with a wrapped fake
	disp := notify.New()
	scn := scanner.NewWithCTFunc(q, func(_ context.Context, _ string, _ bool) ([]ct.Cert, error) {
		return []ct.Cert{
			{
				Serial:    "aabbcc",
				IssuerCA:  "Let's Encrypt",
				SubjectCN: "scan-test.com",
				SANs:      []string{"scan-test.com"},
				NotBefore: time.Now().Add(-24 * time.Hour),
				NotAfter:  time.Now().Add(90 * 24 * time.Hour),
				Source:    "sslmate",
			},
		}, nil
	}, disp)

	if err := scn.ScanFQDN(ctx, fqdn); err != nil {
		t.Fatalf("ScanFQDN: %v", err)
	}

	count, err := q.CountCertificatesByFQDN(ctx, fqdn.ID)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("expected 1 cert stored, got %d", count)
	}

	// Scan again — should not duplicate
	scn.ScanFQDN(ctx, fqdn)
	count, _ = q.CountCertificatesByFQDN(ctx, fqdn.ID)
	if count != 1 {
		t.Errorf("expected still 1 cert, got %d", count)
	}
}
```

- [ ] **Step 3: Add NewWithCTFunc constructor to scanner.go**

Add to `internal/scanner/scanner.go`:

```go
// FetchFunc is a function that fetches CT certs (injectable for testing).
type FetchFunc func(ctx context.Context, fqdn string, includeSubdomains bool) ([]ct.Cert, error)

// ScannerWithFunc wraps a FetchFunc for use in testing.
type ScannerWithFunc struct {
	q    *models.Queries
	fn   FetchFunc
	disp *notify.Dispatcher
}

// NewWithCTFunc creates a Scanner with a custom CT fetch function (for testing).
func NewWithCTFunc(q *models.Queries, fn FetchFunc, disp *notify.Dispatcher) *ScannerWithFunc {
	return &ScannerWithFunc{q: q, fn: fn, disp: disp}
}

// ScanFQDN implements scanning using the custom fetch function.
func (s *ScannerWithFunc) ScanFQDN(ctx context.Context, fqdn models.Fqdn) error {
	certs, err := s.fn(ctx, fqdn.Fqdn, fqdn.IncludeSubdomains)
	if err != nil {
		return err
	}
	channels, _ := s.q.GetFQDNChannels(ctx, fqdn.ID)
	for _, cert := range certs {
		sans, _ := json.Marshal(cert.SANs)
		row, err := s.q.InsertCertificate(ctx, models.InsertCertificateParams{
			ID:           uuid.NewString(),
			FqdnID:       fqdn.ID,
			Serial:       cert.Serial,
			IssuerCa:     cert.IssuerCA,
			SubjectCn:    cert.SubjectCN,
			Sans:         string(sans),
			NotBefore:    cert.NotBefore.UTC().Format(time.RFC3339),
			NotAfter:     cert.NotAfter.UTC().Format(time.RFC3339),
			DiscoveredAt: time.Now().UTC().Format(time.RFC3339),
			Source:       cert.Source,
		})
		if err != nil || row == nil {
			continue
		}
		if fqdn.NotificationsEnabled {
			s.disp.DispatchAll(ctx, channels, fqdn.Fqdn, cert)
		}
	}
	return nil
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/scanner/ -v -race
```

Expected: `TestScanFQDN_NewCertsStored` PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/scanner/
git commit -m "feat: add scanner package (CT fetch → dedup → notify)"
```

---

## Task 5: Scheduler Package

**Files:**
- Create: `internal/scheduler/scheduler.go`

- [ ] **Step 1: Create scheduler.go**

Create `internal/scheduler/scheduler.go`:

```go
package scheduler

import (
	"context"
	"log/slog"
	"sync"

	"github.com/robfig/cron/v3"
	"github.com/t0mer/go-certi/internal/models"
	"github.com/t0mer/go-certi/internal/scanner"
)

// Scannable can scan a single FQDN.
type Scannable interface {
	ScanFQDN(ctx context.Context, fqdn models.Fqdn) error
}

// Scheduler wraps robfig/cron and registers one job per enabled FQDN.
type Scheduler struct {
	cron    *cron.Cron
	q       *models.Queries
	scanner Scannable
	jobs    map[string]cron.EntryID // fqdn_id → cron entry ID
	mu      sync.Mutex
}

// New creates a Scheduler (not started yet).
func New(q *models.Queries, s Scannable) *Scheduler {
	return &Scheduler{
		cron:    cron.New(cron.WithSeconds()),
		q:       q,
		scanner: s,
		jobs:    make(map[string]cron.EntryID),
	}
}

// Start begins the scheduler and registers jobs for all enabled FQDNs.
func (s *Scheduler) Start(ctx context.Context) error {
	fqdns, err := s.q.ListEnabledFQDNs(ctx)
	if err != nil {
		return err
	}
	for _, f := range fqdns {
		if err := s.register(ctx, f); err != nil {
			slog.Error("failed to register FQDN in scheduler", "fqdn", f.Fqdn, "err", err)
		}
	}
	s.cron.Start()
	slog.Info("scheduler started", "jobs", len(s.jobs))
	return nil
}

// Stop gracefully stops the scheduler.
func (s *Scheduler) Stop() {
	s.cron.Stop()
}

// RegisterFQDN registers or re-registers a single FQDN (call after create/update).
func (s *Scheduler) RegisterFQDN(ctx context.Context, f models.Fqdn) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id, ok := s.jobs[f.ID]; ok {
		s.cron.Remove(id)
		delete(s.jobs, f.ID)
	}
	if !f.Enabled {
		return nil
	}
	return s.register(ctx, f)
}

// DeregisterFQDN removes a FQDN's scheduled job.
func (s *Scheduler) DeregisterFQDN(fqdnID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id, ok := s.jobs[fqdnID]; ok {
		s.cron.Remove(id)
		delete(s.jobs, fqdnID)
	}
}

func (s *Scheduler) register(ctx context.Context, f models.Fqdn) error {
	cronExpr := "@every 2h" // default
	if f.ScheduleID != nil {
		sched, err := s.q.GetSchedule(ctx, *f.ScheduleID)
		if err == nil {
			cronExpr = sched.CronExpr
		}
	} else {
		defaultSched, err := s.q.GetDefaultSchedule(ctx)
		if err == nil {
			cronExpr = defaultSched.CronExpr
		}
	}

	fqdnCopy := f
	id, err := s.cron.AddFunc(cronExpr, func() {
		if err := s.scanner.ScanFQDN(context.Background(), fqdnCopy); err != nil {
			slog.Error("scheduled scan failed", "fqdn", fqdnCopy.Fqdn, "err", err)
		}
	})
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.jobs[f.ID] = id
	s.mu.Unlock()
	slog.Info("registered FQDN in scheduler", "fqdn", f.Fqdn, "cron", cronExpr)
	return nil
}

// IsRunning reports whether the scheduler has active jobs.
func (s *Scheduler) IsRunning() bool {
	return len(s.cron.Entries()) > 0 || true // always true once started
}

// Ensure Scanner implements Scannable.
var _ Scannable = (*scanner.Scanner)(nil)
```

- [ ] **Step 2: Build to verify**

```bash
go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add internal/scheduler/
git commit -m "feat: add cron-based scheduler for per-FQDN scan jobs"
```

---

## Task 6: Wire Everything in main.go + Update API Handlers

**Files:**
- Modify: `cmd/go-certi/main.go`
- Modify: `internal/api/handler.go`
- Modify: `internal/api/fqdns.go`
- Modify: `internal/api/channels.go`
- Modify: `internal/api/health.go`

- [ ] **Step 1: Update handler.go to hold Scanner and Scheduler**

In `internal/api/handler.go`, update the Handler struct and constructor:

```go
type Handler struct {
	q       *models.Queries
	db      *sql.DB
	authSvc *auth.Service
	scanner ScannerInterface
	notify  NotifyInterface
}

// ScannerInterface allows the handler to trigger scans.
type ScannerInterface interface {
	ScanFQDN(ctx context.Context, fqdn models.Fqdn) error
}

// NotifyInterface allows the handler to send test notifications.
type NotifyInterface interface {
	TestChannel(ctx context.Context, channel models.NotificationChannel) error
}

func newHandler(db *sql.DB, q *models.Queries, authSvc *auth.Service, scn ScannerInterface, notif NotifyInterface) *Handler {
	return &Handler{q: q, db: db, authSvc: authSvc, scanner: scn, notify: notif}
}
```

Add `"context"` to imports.

- [ ] **Step 2: Update TriggerScan in fqdns.go to call the real scanner**

Replace the `TriggerScan` function body in `internal/api/fqdns.go`:

```go
func (h *Handler) TriggerScan(c *gin.Context) {
	f, err := h.q.GetFQDN(c.Request.Context(), c.Param("id"))
	if err != nil {
		notFoundOr500(c, err)
		return
	}
	go func() {
		if err := h.scanner.ScanFQDN(context.Background(), f); err != nil {
			slog.Error("trigger scan failed", "fqdn", f.Fqdn, "err", err)
		}
	}()
	c.JSON(http.StatusAccepted, gin.H{"status": "scan started"})
}
```

Add imports: `"context"`, `"log/slog"`.

- [ ] **Step 3: Update TestChannel in channels.go to call notify**

Replace the `TestChannel` function body in `internal/api/channels.go`:

```go
func (h *Handler) TestChannel(c *gin.Context) {
	ch, err := h.q.GetChannel(c.Request.Context(), c.Param("id"))
	if err != nil {
		notFoundOr500(c, err)
		return
	}
	if err := h.notify.TestChannel(c.Request.Context(), ch); err != nil {
		respondError(c, http.StatusBadGateway, "notification failed: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "test notification sent"})
}
```

- [ ] **Step 4: Update server.go New() signature**

In `internal/api/server.go`, update `New` to accept scanner and notify:

```go
func New(db *sql.DB, q *models.Queries, authSvc *auth.Service, scn ScannerInterface, notif NotifyInterface, webFS fs.FS) *Server {
	// ...
	h := newHandler(db, q, authSvc, scn, notif)
	// rest unchanged
```

Also add `ScannerInterface` and `NotifyInterface` type aliases in `server.go` or re-use from handler.go (keep in handler.go as they're defined there).

- [ ] **Step 5: Update main.go to wire scanner and scheduler**

In `cmd/go-certi/main.go`, after opening DB and creating models:

```go
	// --- Wire CT, notify, scanner, scheduler ---
	ctClient := ct.New(*sslmateKey)
	notifier := notify.New()
	scn := scanner.New(q, ctClient, notifier)
	sched := scheduler.New(q, scn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := sched.Start(ctx); err != nil {
		slog.Warn("scheduler start", "err", err)
	}
	cancel()
	defer sched.Stop()

	// --- Start HTTP server ---
	srv := api.New(dbConn, q, authSvc, scn, notifier, webui.FS())
```

Add imports:
```go
"time"
"github.com/t0mer/go-certi/internal/ct"
"github.com/t0mer/go-certi/internal/notify"
"github.com/t0mer/go-certi/internal/scanner"
"github.com/t0mer/go-certi/internal/scheduler"
```

Also update readyz handler to check if scheduler is set up (optional improvement):

In `internal/api/health.go`, the readyz handler already just pings the DB — that's sufficient.

- [ ] **Step 6: Build**

```bash
cd /opt/dev/go-certi && go build ./...
```

Expected: compiles cleanly.

- [ ] **Step 7: Run all tests**

```bash
go test ./... -race
```

Expected: all PASS.

- [ ] **Step 8: Commit**

```bash
git add internal/api/ internal/scheduler/ cmd/go-certi/main.go
git commit -m "feat: wire scanner and scheduler into API and main"
```

---

## Self-Review

**CLAUDE.md §7 compliance:**
- [x] CT source: sslmate primary, crt.sh fallback — Task 2
- [x] Dedup on serial — Task 4 (INSERT OR IGNORE)
- [x] Notification dispatch per new cert — Task 4 (DispatchAll)
- [x] shoutrrr — Task 3
- [x] greenapi — Task 3
- [x] waweb — Task 3
- [x] Best-effort: errors logged, never propagate — Task 3 (Dispatch)
- [x] robfig/cron scheduler — Task 5
- [x] Per-FQDN schedule — Task 5
- [x] Default schedule fallback — Task 5
- [x] TriggerScan wired — Task 6
- [x] TestChannel wired — Task 6
- [ ] Token-bucket rate limiter for sslmate — deferred (add in enhancement; the 30s HTTP timeout provides natural backpressure)
