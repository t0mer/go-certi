package api_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/t0mer/go-certi/internal/api"
	"github.com/t0mer/go-certi/internal/auth"
	"github.com/t0mer/go-certi/internal/db"
	"github.com/t0mer/go-certi/internal/models"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

func TestHealthz(t *testing.T) {
	conn := openTestDB(t)
	srv := api.New(conn, models.New(conn), auth.New("test-secret"), nil, nil, nil, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestReadyz_WithDB(t *testing.T) {
	conn := openTestDB(t)
	srv := api.New(conn, models.New(conn), auth.New("test-secret"), nil, nil, nil, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestReadyz_ClosedDB(t *testing.T) {
	conn := openTestDB(t)
	conn.Close() // force PingContext to fail

	srv := api.New(conn, models.New(conn), auth.New("test-secret"), nil, nil, nil, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}
