package db_test

import (
	"path/filepath"
	"testing"

	"github.com/t0mer/go-certi/internal/db"
)

func TestOpen_CreatesAndMigrates(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	conn, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer conn.Close()

	// Verify schema_migrations table exists and has our migration recorded
	var count int
	err = conn.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
	if err != nil {
		t.Fatalf("schema_migrations query: %v", err)
	}
	if count < 1 {
		t.Fatal("expected at least one migration to be recorded")
	}
}

func TestOpen_IdempotentMigrations(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	// Open twice — migrations must not error on second run
	for i := range 2 {
		conn, err := db.Open(dbPath)
		if err != nil {
			t.Fatalf("Open attempt %d: %v", i+1, err)
		}
		conn.Close()
	}
}
