package ct_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/t0mer/go-certi/internal/ct"
)

func TestSslmateFetch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"tbs_sha256": "abc123",
				"dns_names":  []string{"example.com", "www.example.com"},
				"issuer":     map[string]any{"name": "Let's Encrypt"},
				"not_before": "2026-01-01T00:00:00Z",
				"not_after":  "2026-04-01T00:00:00Z",
			},
		})
	}))
	defer srv.Close()

	client := ct.NewWithBaseURL("test-api-key", srv.URL, "")
	certs, err := client.FetchCerts(context.Background(), "example.com", false)
	if err != nil {
		t.Fatalf("FetchCerts: %v", err)
	}
	if len(certs) == 0 {
		t.Error("expected at least one cert")
	}
	if certs[0].Source != "sslmate" {
		t.Errorf("source = %q, want sslmate", certs[0].Source)
	}
	if len(certs[0].SANs) != 2 {
		t.Errorf("SANs = %v, want 2", certs[0].SANs)
	}
}

func TestCrtshFetch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"serial_number": "01020304",
				"issuer_name":   "C=US, O=Let's Encrypt, CN=R3",
				"common_name":   "example.com",
				"name_value":    "example.com\nsub.example.com",
				"not_before":    "2026-01-01T00:00:00",
				"not_after":     "2026-04-01T00:00:00",
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
	if len(certs[0].SANs) != 2 {
		t.Errorf("SANs = %v", certs[0].SANs)
	}
}

func TestFallbackToCrtsh(t *testing.T) {
	// sslmate fails (500), should fall back to crt.sh
	sslmateSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer sslmateSrv.Close()

	crtshSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]map[string]any{
			{"serial_number": "aa", "issuer_name": "CA", "common_name": "x.com",
				"name_value": "x.com", "not_before": "2026-01-01T00:00:00", "not_after": "2026-04-01T00:00:00"},
		})
	}))
	defer crtshSrv.Close()

	client := ct.NewWithBaseURL("test-key", sslmateSrv.URL, crtshSrv.URL)
	certs, err := client.FetchCerts(context.Background(), "x.com", false)
	if err != nil {
		t.Fatalf("FetchCerts: %v", err)
	}
	if len(certs) == 0 || certs[0].Source != "crt.sh" {
		t.Errorf("expected crt.sh fallback, got source=%q len=%d", certs[0].Source, len(certs))
	}
}
