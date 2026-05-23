package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/t0mer/go-certi/internal/api"
	"github.com/t0mer/go-certi/internal/auth"
	"github.com/t0mer/go-certi/internal/db"
	"github.com/t0mer/go-certi/internal/models"
)

func setupTestServer(t *testing.T) *api.Server {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	q := models.New(conn)
	authSvc := auth.New("test-secret")
	return api.New(conn, q, authSvc, nil, nil, nil)
}

func TestAPIFQDNCRUD(t *testing.T) {
	srv := setupTestServer(t)

	body, _ := json.Marshal(map[string]any{
		"fqdn":    "test.example.com",
		"enabled": true,
	})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/fqdns", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	srv.ServeHTTP(w, r)
	if w.Code != http.StatusCreated {
		t.Fatalf("create: got %d, want 201. body: %s", w.Code, w.Body)
	}
	var created api.FQDNResponse
	json.NewDecoder(w.Body).Decode(&created)
	if created.FQDN != "test.example.com" {
		t.Errorf("FQDN = %q", created.FQDN)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/api/v1/fqdns/"+created.ID, nil)
	srv.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("get: got %d", w.Code)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/api/v1/fqdns", nil)
	srv.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("list: got %d", w.Code)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodDelete, "/api/v1/fqdns/"+created.ID, nil)
	srv.ServeHTTP(w, r)
	if w.Code != http.StatusNoContent {
		t.Fatalf("delete: got %d", w.Code)
	}
}

func TestAPISchedulesCRUD(t *testing.T) {
	srv := setupTestServer(t)

	body, _ := json.Marshal(map[string]any{
		"name":      "Hourly",
		"cron_expr": "@every 1h",
	})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/schedules", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	srv.ServeHTTP(w, r)
	if w.Code != http.StatusCreated {
		t.Fatalf("create schedule: got %d. body: %s", w.Code, w.Body)
	}
}

func TestAPISettingsGetUpdate(t *testing.T) {
	srv := setupTestServer(t)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil)
	srv.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("get settings: got %d", w.Code)
	}

	body, _ := json.Marshal(map[string]any{"theme": "dark"})
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPut, "/api/v1/settings", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	srv.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("update settings: got %d. body: %s", w.Code, w.Body)
	}
}

func TestAPICertificatesList(t *testing.T) {
	srv := setupTestServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/certificates", nil)
	srv.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("list certs: got %d", w.Code)
	}
}

func TestAPIChannelsCRUD(t *testing.T) {
	srv := setupTestServer(t)

	body, _ := json.Marshal(map[string]any{
		"name":   "Telegram",
		"type":   "shoutrrr",
		"config": map[string]string{"url": "telegram://token@telegram?chats=123"},
	})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/channels", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	srv.ServeHTTP(w, r)
	if w.Code != http.StatusCreated {
		t.Fatalf("create channel: got %d. body: %s", w.Code, w.Body)
	}
}
