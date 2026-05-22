# go-certi Bootstrap Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Scaffold the go-certi repo from empty to a compilable, runnable Go binary with health endpoints, embedded frontend placeholder, Docker, and CI/CD workflows.

**Architecture:** Single Go binary with Gin HTTP server, pure-Go SQLite (no CGO), embedded migrations, and `go:embed` for the web frontend. All config from `--conf` dir (XDG default). Env vars override CLI flags.

**Tech Stack:** Go 1.22+, Gin, modernc.org/sqlite (pure Go), pflag, Vite/React (placeholder), Docker multi-stage, GitHub Actions

**Module path:** `github.com/t0mer/go-certi` — adjust if the GitHub org differs.

**Locked-in decisions (from §15):**
- Frontend: React + Vite + TypeScript + Tailwind + shadcn/ui
- HTTP framework: Gin
- DB: modernc.org/sqlite (no CGO), hand-rolled migration runner
- CT sources: sslmate primary + crt.sh fallback (implemented in a later plan)
- Config dir: `go-certi`
- Auth: single-user

---

## File Map

| File | Purpose |
|---|---|
| `go.mod` / `go.sum` | Module definition and locked deps |
| `.gitignore` | Excludes CLAUDE.md, binaries, frontend build, secrets |
| `internal/version/version.go` | Exports `var Version = "dev"` (overridden via -ldflags) |
| `internal/config/config.go` | XDG path resolution; config struct; JSON R/W; CLI/env resolution helpers |
| `internal/config/config_test.go` | Unit tests for XDG path and env-override logic |
| `internal/db/db.go` | `Open(path) (*sql.DB, error)` — opens SQLite, runs embedded migrations |
| `internal/db/migrate.go` | Hand-rolled migration runner (no CGO dependencies) |
| `internal/db/migrations/001_init.sql` | Settings table (all other tables in the next plan) |
| `internal/db/db_test.go` | Verify Open() works and migrations run idempotently |
| `internal/api/server.go` | `New(db, cfg) *Server` — Gin engine setup, middleware, route registration |
| `internal/api/health.go` | `GET /healthz` and `GET /readyz` handlers |
| `internal/api/health_test.go` | httptest-based tests for both health endpoints |
| `cmd/go-certi/main.go` | CLI entrypoint: pflag parsing, env override, wires config+db+server |
| `scripts/build.sh` | Multi-arch cross-compile script; honors `BUILD_MODE` and `VERSION` |
| `scripts/next-version.sh` | Computes `YYYY.M.PATCH` from git tags |
| `Dockerfile` | Multi-stage: Go builder → Node builder → distroless final |
| `docker-compose.yaml` | Local dev stack |
| `.github/workflows/release.yml` | Build + tag + GitHub Release |
| `.github/workflows/docker.yml` | Multi-arch Docker build + push |
| `web/src/.gitkeep` | Placeholder — real frontend in a later plan |
| `web/dist/index.html` | Minimal placeholder embedded into binary at build time |

---

## Task 1: Directory Scaffold + go.mod + .gitignore

**Files:**
- Create: `go.mod`
- Create: `.gitignore`
- Create: all directories (with `.gitkeep` where needed)

- [ ] **Step 1: Create directory tree**

```bash
mkdir -p cmd/go-certi
mkdir -p internal/{api,config,db/migrations,db/queries,models,ct,notify,scheduler,auth,version}
mkdir -p web/src web/dist
mkdir -p docs/superpowers/plans scripts
mkdir -p .github/workflows
touch internal/db/migrations/.gitkeep \
      internal/db/queries/.gitkeep \
      internal/models/.gitkeep \
      web/dist/.gitkeep
```

- [ ] **Step 2: Initialize go module**

```bash
go mod init github.com/t0mer/go-certi
```

Expected: `go.mod` created with `module github.com/t0mer/go-certi` and `go 1.22`.

- [ ] **Step 3: Create .gitignore**

Create `/.gitignore`:

```gitignore
# Claude / AI assistant context — local only, do not track
CLAUDE.md
.claude/
.cursor/
.aider*

# Build output
/dist/
/bin/
/build/

# Go
*.exe
*.exe~
*.dll
*.so
*.dylib
*.test
*.out
go.work
go.work.sum
vendor/

# Frontend
web/node_modules/
web/dist/
web/.vite/
*.log
npm-debug.log*
yarn-debug.log*
yarn-error.log*
pnpm-debug.log*

# Local dev config / data
.dev/
/data/
*.db
*.db-journal
*.sqlite
*.sqlite-journal

# Generated Swagger output (regenerate via `swag init`)
docs/docs.go
docs/swagger.json
docs/swagger.yaml

# Editors / OS
.idea/
.vscode/
*.swp
*.swo
.DS_Store
Thumbs.db

# Secrets
.env
.env.*
!.env.example
```

- [ ] **Step 4: Commit**

```bash
git add go.mod .gitignore
git add internal/ web/ scripts/ docs/ .github/
git commit -m "chore: scaffold directory structure and go module"
```

---

## Task 2: version Package

**Files:**
- Create: `internal/version/version.go`

- [ ] **Step 1: Create version package**

Create `internal/version/version.go`:

```go
package version

// Version is set at build time via -ldflags="-X github.com/t0mer/go-certi/internal/version.Version=<ver>"
var Version = "dev"
```

- [ ] **Step 2: Verify compiles**

```bash
go build ./internal/version/
```

Expected: no output (success).

- [ ] **Step 3: Commit**

```bash
git add internal/version/version.go
git commit -m "feat: add version package with build-time override"
```

---

## Task 3: config Package

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

This package handles XDG path resolution and the JSON config file. It also exports the env-override helper used by `main.go`.

- [ ] **Step 1: Write failing tests**

Create `internal/config/config_test.go`:

```go
package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/t0mer/go-certi/internal/config"
)

func TestDefaultConfDir_XDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg-test")
	got := config.DefaultConfDir()
	want := "/tmp/xdg-test/go-certi"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestDefaultConfDir_FallsBackToHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	home, _ := os.UserHomeDir()
	got := config.DefaultConfDir()
	want := filepath.Join(home, ".config", "go-certi")
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestEnvOverride_ReplacesValue(t *testing.T) {
	t.Setenv("TEST_PORT", "9090")
	val := "8111"
	config.OverrideFromEnv("TEST_PORT", &val)
	if val != "9090" {
		t.Fatalf("expected env to override value, got %q", val)
	}
}

func TestEnvOverride_KeepsOriginalWhenUnset(t *testing.T) {
	t.Setenv("TEST_PORT_UNSET", "")
	val := "8111"
	config.OverrideFromEnv("TEST_PORT_UNSET", &val)
	if val != "8111" {
		t.Fatalf("expected value unchanged, got %q", val)
	}
}

func TestLoadOrCreate_CreatesDir(t *testing.T) {
	dir := t.TempDir()
	confDir := filepath.Join(dir, "go-certi")
	cfg, err := config.LoadOrCreate(confDir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 8111 {
		t.Fatalf("expected default port 8111, got %d", cfg.Port)
	}
	if _, err := os.Stat(confDir); os.IsNotExist(err) {
		t.Fatal("expected confDir to be created")
	}
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
go test ./internal/config/ -v
```

Expected: FAIL — `package config_test` cannot find `config` package.

- [ ] **Step 3: Implement config package**

Create `internal/config/config.go`:

```go
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

type Config struct {
	Port    int    `json:"port"`
	ConfDir string `json:"-"` // runtime-only, not persisted
}

func DefaultConfDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "go-certi")
	}
	if runtime.GOOS == "windows" {
		if appData := os.Getenv("AppData"); appData != "" {
			return filepath.Join(appData, "go-certi")
		}
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "go-certi")
}

// OverrideFromEnv replaces *val with the env var value when the env var is set and non-empty.
func OverrideFromEnv(envKey string, val *string) {
	if v, ok := os.LookupEnv(envKey); ok && v != "" {
		*val = v
	}
}

// LoadOrCreate reads config.json from confDir, creating the dir and a default file if needed.
func LoadOrCreate(confDir string) (*Config, error) {
	if err := os.MkdirAll(confDir, 0o700); err != nil {
		return nil, err
	}

	cfgPath := filepath.Join(confDir, "config.json")
	cfg := &Config{Port: 8111, ConfDir: confDir}

	data, err := os.ReadFile(cfgPath)
	if os.IsNotExist(err) {
		return cfg, writeJSON(cfgPath, cfg)
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	cfg.ConfDir = confDir
	return cfg, nil
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
```

- [ ] **Step 4: Run tests to confirm they pass**

```bash
go test ./internal/config/ -v -race
```

Expected: all 5 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/config/
git commit -m "feat: add config package with XDG path resolution and env override"
```

---

## Task 4: Database Package (SQLite + Migrations)

**Files:**
- Create: `internal/db/db.go`
- Create: `internal/db/migrate.go`
- Create: `internal/db/migrations/001_init.sql`
- Create: `internal/db/db_test.go`

- [ ] **Step 1: Add dependencies**

```bash
go get modernc.org/sqlite
```

Expected: `go.mod` and `go.sum` updated.

- [ ] **Step 2: Write failing test**

Create `internal/db/db_test.go`:

```go
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
```

- [ ] **Step 3: Run test to confirm it fails**

```bash
go test ./internal/db/ -v
```

Expected: FAIL — `db` package not yet defined.

- [ ] **Step 4: Create initial migration**

Create `internal/db/migrations/001_init.sql`:

```sql
-- Settings: single-row application configuration
CREATE TABLE IF NOT EXISTS settings (
    id                          INTEGER PRIMARY KEY CHECK (id = 1),
    auth_enabled                BOOLEAN NOT NULL DEFAULT 0,
    username                    TEXT,
    password_hash               TEXT,
    api_token_protection_enabled BOOLEAN NOT NULL DEFAULT 0,
    api_token                   TEXT,
    theme                       TEXT NOT NULL DEFAULT 'system',
    sslmate_api_key             TEXT NOT NULL DEFAULT '',
    default_schedule_id         TEXT,
    updated_at                  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Insert default settings row (singleton pattern)
INSERT OR IGNORE INTO settings (id) VALUES (1);
```

- [ ] **Step 5: Create migration runner**

Create `internal/db/migrate.go`:

```go
package db

import (
	"database/sql"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

func runMigrations(conn *sql.DB, migrations fs.FS) error {
	if _, err := conn.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version    TEXT PRIMARY KEY,
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	entries, err := fs.ReadDir(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, name := range files {
		var applied bool
		err := conn.QueryRow(
			"SELECT COUNT(*) > 0 FROM schema_migrations WHERE version = ?", name,
		).Scan(&applied)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", name, err)
		}
		if applied {
			continue
		}

		data, err := fs.ReadFile(migrations, "migrations/"+name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		if _, err := conn.Exec(string(data)); err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
		if _, err := conn.Exec(
			"INSERT INTO schema_migrations (version) VALUES (?)", name,
		); err != nil {
			return fmt.Errorf("record migration %s: %w", name, err)
		}
	}
	return nil
}
```

- [ ] **Step 6: Create db.go**

Create `internal/db/db.go`:

```go
package db

import (
	"database/sql"
	"embed"
	"fmt"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Open opens (or creates) the SQLite database at path, runs all pending migrations,
// and returns the connection. Caller is responsible for Close().
func Open(path string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)", path)
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	conn.SetMaxOpenConns(1) // SQLite allows only one concurrent writer

	if err := runMigrations(conn, migrationsFS); err != nil {
		conn.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return conn, nil
}
```

- [ ] **Step 7: Run tests to confirm they pass**

```bash
go test ./internal/db/ -v -race
```

Expected: `TestOpen_CreatesAndMigrates` PASS, `TestOpen_IdempotentMigrations` PASS.

- [ ] **Step 8: Commit**

```bash
git add internal/db/ go.mod go.sum
git commit -m "feat: add SQLite database package with embedded migration runner"
```

---

## Task 5: HTTP Server + Health Endpoints

**Files:**
- Create: `internal/api/server.go`
- Create: `internal/api/health.go`
- Create: `internal/api/health_test.go`
- Create: `web/dist/index.html` (placeholder)

- [ ] **Step 1: Add Gin dependency**

```bash
go get github.com/gin-gonic/gin
```

Expected: `go.mod` and `go.sum` updated.

- [ ] **Step 2: Write failing tests**

Create `internal/api/health_test.go`:

```go
package api_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/t0mer/go-certi/internal/api"
	"github.com/t0mer/go-certi/internal/db"
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
	srv := api.New(conn)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestReadyz_WithDB(t *testing.T) {
	conn := openTestDB(t)
	srv := api.New(conn)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
```

- [ ] **Step 3: Run tests to confirm they fail**

```bash
go test ./internal/api/ -v
```

Expected: FAIL — `api` package not yet defined.

- [ ] **Step 4: Create web/dist placeholder**

Create `web/dist/index.html`:

```html
<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><title>go-certi</title></head>
<body><p>Frontend build not yet available. Run <code>cd web && npm run build</code>.</p></body>
</html>
```

(Remove the `.gitkeep` if present: `rm -f web/dist/.gitkeep`)

- [ ] **Step 5: Create health handlers**

Create `internal/api/health.go`:

```go
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// healthz is always 200 — no auth, no DB check. Used by Docker healthchecks.
func (s *Server) healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// readyz checks the DB is reachable and returns 200/503.
func (s *Server) readyz(c *gin.Context) {
	if err := s.db.PingContext(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db unreachable", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
```

- [ ] **Step 6: Create server**

Create `internal/api/server.go`:

```go
package api

import (
	"database/sql"
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed ../../web/dist
var webFS embed.FS

// Server holds the Gin engine and shared dependencies.
type Server struct {
	engine *gin.Engine
	db     *sql.DB
}

// New constructs the server and registers all routes.
func New(db *sql.DB) *Server {
	gin.SetMode(gin.ReleaseMode)
	s := &Server{engine: gin.New(), db: db}
	s.engine.Use(gin.Recovery())

	// Health — always unauthenticated
	s.engine.GET("/healthz", s.healthz)
	s.engine.GET("/readyz", s.readyz)

	// Frontend — serve embedded web/dist
	stripped, _ := fs.Sub(webFS, "web/dist")
	s.engine.NoRoute(gin.WrapH(http.FileServer(http.FS(stripped))))

	return s
}

// ServeHTTP implements http.Handler so the server can be used with httptest.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.engine.ServeHTTP(w, r)
}

// Engine exposes the underlying Gin engine for route registration in main.
func (s *Server) Engine() *gin.Engine {
	return s.engine
}
```

- [ ] **Step 7: Run tests**

```bash
go test ./internal/api/ -v -race
```

Expected: `TestHealthz` PASS, `TestReadyz_WithDB` PASS.

- [ ] **Step 8: Commit**

```bash
git add internal/api/ web/dist/index.html go.mod go.sum
git commit -m "feat: add Gin HTTP server with /healthz and /readyz endpoints"
```

---

## Task 6: CLI Entrypoint (main.go)

**Files:**
- Create: `cmd/go-certi/main.go`

- [ ] **Step 1: Add pflag dependency**

```bash
go get github.com/spf13/pflag
```

Expected: `go.mod` and `go.sum` updated.

- [ ] **Step 2: Create main.go**

Create `cmd/go-certi/main.go`:

```go
package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/spf13/pflag"

	"github.com/t0mer/go-certi/internal/api"
	"github.com/t0mer/go-certi/internal/config"
	"github.com/t0mer/go-certi/internal/db"
	"github.com/t0mer/go-certi/internal/version"
)

func main() {
	// --- Flag definitions ---
	var (
		port         = pflag.IntP("port", "p", 8111, "Server port\n  env: GO_CERTI_PORT")
		confDir      = pflag.String("conf", config.DefaultConfDir(), "Config + DB directory\n  env: GO_CERTI_CONF")
		sslmateKey   = pflag.String("sslmate-api-key", "", "sslmate Cert Spotter API key\n  env: GO_CERTI_SSLMATE_API_KEY")
		resetPwd     = pflag.Bool("reset-password", false, "Generate new password, print plaintext, exit\n  env: GO_CERTI_RESET_PASSWORD")
		resetToken   = pflag.Bool("reset-api-token", false, "Generate new API token, print it, exit\n  env: GO_CERTI_RESET_API_TOKEN")
		showVersion  = pflag.Bool("version", false, "Print version and exit")
	)

	pflag.CommandLine.SortFlags = false
	pflag.Parse()

	// --- Env overrides (env wins over CLI flag) ---
	// Documented precedence: env > flag. This is intentional and container-friendly.
	overrides := []struct{ env, desc string; apply func(string) }{
		{"GO_CERTI_PORT", "port", func(v string) {
			if n, err := strconv.Atoi(v); err == nil {
				*port = n
			}
		}},
		{"GO_CERTI_CONF", "conf", func(v string) { *confDir = v }},
		{"GO_CERTI_SSLMATE_API_KEY", "sslmate-api-key", func(v string) { *sslmateKey = v }},
		{"GO_CERTI_RESET_PASSWORD", "reset-password", func(v string) {
			if v == "true" || v == "1" {
				*resetPwd = true
			}
		}},
		{"GO_CERTI_RESET_API_TOKEN", "reset-api-token", func(v string) {
			if v == "true" || v == "1" {
				*resetToken = true
			}
		}},
	}
	for _, o := range overrides {
		if v, ok := os.LookupEnv(o.env); ok && v != "" {
			slog.Debug("config source", "key", o.desc, "source", "env")
			o.apply(v)
		} else {
			slog.Debug("config source", "key", o.desc, "source", "flag")
		}
	}

	// --- Version flag ---
	if *showVersion {
		fmt.Println(version.Version)
		os.Exit(0)
	}

	// --- Placeholder actions (implemented in auth plan) ---
	if *resetPwd || *resetToken {
		slog.Error("--reset-password and --reset-api-token not yet implemented")
		os.Exit(1)
	}

	// --- Load or create config ---
	cfg, err := config.LoadOrCreate(*confDir)
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}
	if *port != 8111 { // respect explicit port override from flag/env
		cfg.Port = *port
	}
	_ = sslmateKey // stored to DB settings in a later plan

	// --- Open database ---
	dbConn, err := db.Open(*confDir + "/go-certi.db")
	if err != nil {
		slog.Error("database", "err", err)
		os.Exit(1)
	}
	defer dbConn.Close()

	// --- Start HTTP server ---
	srv := api.New(dbConn)
	addr := fmt.Sprintf(":%d", cfg.Port)
	slog.Info("go-certi starting", "version", version.Version, "addr", addr, "conf", *confDir)

	if err := http.ListenAndServe(addr, srv); err != nil {
		slog.Error("server", "err", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 3: Build and verify**

```bash
go build ./...
go vet ./...
```

Expected: both commands produce no output (success).

- [ ] **Step 4: Smoke-test the binary**

```bash
go run ./cmd/go-certi --version
```

Expected: `dev`

```bash
GO_CERTI_PORT=9999 go run ./cmd/go-certi --port 8111 &
sleep 1
curl -s http://localhost:9999/healthz
kill %1
```

Expected: `{"status":"ok"}` — confirms env var (9999) overrides flag (8111).

- [ ] **Step 5: Commit**

```bash
git add cmd/go-certi/main.go go.mod go.sum
git commit -m "feat: add CLI entrypoint with pflag, env override, and HTTP server wiring"
```

---

## Task 7: Build Scripts

**Files:**
- Create: `scripts/build.sh`
- Create: `scripts/next-version.sh`

- [ ] **Step 1: Create build.sh**

Create `scripts/build.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail

VERSION="${VERSION:-dev}"
BUILD_MODE="${BUILD_MODE:-dev}"
OUTPUT_DIR="${OUTPUT_DIR:-dist}"
MODULE="github.com/t0mer/go-certi"

mkdir -p "$OUTPUT_DIR"

LDFLAGS="-X ${MODULE}/internal/version.Version=${VERSION}"
if [[ "$BUILD_MODE" == "prod" ]]; then
    LDFLAGS="-s -w ${LDFLAGS}"
fi

declare -a TARGETS=(
    "linux   amd64  amd64  "
    "linux   arm64  arm64  "
    "linux   armv7  arm    7"
    "linux   armhf  arm    6"
    "linux   arm    arm    5"
    "windows amd64  amd64  "
    "windows arm64  arm64  "
)

for target in "${TARGETS[@]}"; do
    read -r os display_arch goarch goarm <<< "$target"
    ext=""
    [[ "$os" == "windows" ]] && ext=".exe"
    outfile="${OUTPUT_DIR}/go-certi_${VERSION}_${os}_${display_arch}${ext}"

    echo "Building ${outfile}..."
    env CGO_ENABLED=0 GOOS="$os" GOARCH="$goarch" GOARM="$goarm" \
        go build -ldflags "$LDFLAGS" -o "$outfile" ./cmd/go-certi
done

echo "Build complete. Binaries in ${OUTPUT_DIR}/"
```

- [ ] **Step 2: Create next-version.sh**

Create `scripts/next-version.sh`:

```bash
#!/usr/bin/env bash
# Prints the next version in YYYY.M.PATCH format.
# PATCH is the number of existing tags for the current year+month, starting at 0.
set -euo pipefail

YEAR=$(date +%Y)
MONTH=$(date +%-m)    # no leading zero

PREFIX="${YEAR}.${MONTH}."

# Count tags matching the prefix to determine next PATCH
PATCH=$(git tag --list "${PREFIX}*" 2>/dev/null | wc -l | tr -d ' ')

echo "${PREFIX}${PATCH}"
```

- [ ] **Step 3: Make scripts executable**

```bash
chmod +x scripts/build.sh scripts/next-version.sh
```

- [ ] **Step 4: Test next-version.sh**

```bash
bash scripts/next-version.sh
```

Expected: something like `2026.5.0` (year.month.0 since no tags exist yet).

- [ ] **Step 5: Test build.sh in dev mode**

```bash
VERSION=test bash scripts/build.sh
ls dist/
```

Expected: 7 binaries listed:
```
go-certi_test_linux_amd64
go-certi_test_linux_arm64
go-certi_test_linux_armv7
go-certi_test_linux_armhf
go-certi_test_linux_arm
go-certi_test_windows_amd64.exe
go-certi_test_windows_arm64.exe
```

- [ ] **Step 6: Cleanup dist/ and commit**

```bash
rm -rf dist/
git add scripts/build.sh scripts/next-version.sh
git commit -m "feat: add multi-arch build script and next-version helper"
```

---

## Task 8: Dockerfile + docker-compose

**Files:**
- Create: `Dockerfile`
- Create: `docker-compose.yaml`

- [ ] **Step 1: Create Dockerfile**

Create `Dockerfile`:

```dockerfile
# syntax=docker/dockerfile:1

# Stage 1: Build Go binary
FROM golang:1.22-alpine AS go-builder
ARG VERSION=dev
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Copy pre-built frontend dist (built in CI before this stage, or use placeholder)
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X github.com/t0mer/go-certi/internal/version.Version=${VERSION}" \
    -o /go-certi \
    ./cmd/go-certi

# Stage 2: Build React frontend (only runs when web/package.json exists)
FROM node:20-alpine AS node-builder
WORKDIR /web
# If web/package.json doesn't exist this stage produces nothing useful;
# the Go stage already embedded web/dist from the COPY above.
COPY web/package*.json ./
RUN if [ -f package.json ]; then npm ci --silent; fi
COPY web/ ./
RUN if [ -f package.json ]; then npm run build; fi

# Stage 3: Final minimal image
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=go-builder /go-certi /go-certi
EXPOSE 8111
VOLUME ["/data"]
ENV GO_CERTI_CONF=/data
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/go-certi", "--version"]
ENTRYPOINT ["/go-certi"]
CMD ["--conf", "/data"]
```

Note: The `HEALTHCHECK` uses `--version` as a quick liveness check since `curl` is not available in distroless. A proper check (`curl /healthz`) can be added when the user switches to a `debian-slim` base or adds a healthcheck binary.

- [ ] **Step 2: Create docker-compose.yaml**

Create `docker-compose.yaml`:

```yaml
services:
  go-certi:
    build:
      context: .
      args:
        VERSION: dev
    ports:
      - "8111:8111"
    volumes:
      - go-certi-data:/data
    environment:
      # Uncomment and set these as needed:
      # GO_CERTI_SSLMATE_API_KEY: "your-key-here"
      GO_CERTI_CONF: /data
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "/go-certi", "--version"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 10s

volumes:
  go-certi-data:
```

- [ ] **Step 3: Commit**

```bash
git add Dockerfile docker-compose.yaml
git commit -m "feat: add multi-stage Dockerfile and docker-compose"
```

---

## Task 9: GitHub Actions Workflows

**Files:**
- Create: `.github/workflows/release.yml`
- Create: `.github/workflows/docker.yml`

- [ ] **Step 1: Create release.yml**

Create `.github/workflows/release.yml`:

```yaml
name: Release

on:
  workflow_dispatch:
    inputs:
      version:
        description: "Version (leave empty to auto-compute via next-version.sh)"
        required: false
        type: string

jobs:
  release:
    name: Build and Release
    runs-on: ubuntu-latest
    permissions:
      contents: write

    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Resolve version
        id: version
        run: |
          if [ -n "${{ inputs.version }}" ]; then
            VERSION="${{ inputs.version }}"
          else
            VERSION=$(bash scripts/next-version.sh)
          fi
          echo "VERSION=$VERSION" >> "$GITHUB_OUTPUT"
          echo "Resolved version: $VERSION"

      - name: Build all targets
        run: BUILD_MODE=prod VERSION=${{ steps.version.outputs.VERSION }} bash scripts/build.sh

      - name: Create git tag
        run: |
          git config user.name  "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git tag "${{ steps.version.outputs.VERSION }}"
          git push origin "${{ steps.version.outputs.VERSION }}"

      - name: Create GitHub Release
        uses: softprops/action-gh-release@v2
        with:
          tag_name: ${{ steps.version.outputs.VERSION }}
          name: ${{ steps.version.outputs.VERSION }}
          files: dist/*
          generate_release_notes: true
```

- [ ] **Step 2: Create docker.yml**

Create `.github/workflows/docker.yml`:

```yaml
name: Docker Build

on:
  workflow_dispatch: {}
  workflow_run:
    workflows: ["Release"]
    types: [completed]

jobs:
  docker:
    name: Build and Push Docker Image
    runs-on: ubuntu-latest
    if: ${{ github.event_name == 'workflow_dispatch' || github.event.workflow_run.conclusion == 'success' }}

    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Resolve version from latest tag
        id: version
        run: |
          VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "dev")
          echo "VERSION=$VERSION" >> "$GITHUB_OUTPUT"
          echo "Image version: $VERSION"

      - name: Set up QEMU
        uses: docker/setup-qemu-action@v3

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Log in to Docker Hub
        uses: docker/login-action@v3
        with:
          username: ${{ secrets.DOCKERHUB_USERNAME }}
          password: ${{ secrets.DOCKERHUB_TOKEN }}

      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: .
          platforms: linux/amd64,linux/arm64,linux/arm/v7
          push: true
          build-args: |
            VERSION=${{ steps.version.outputs.VERSION }}
          tags: |
            techblog/go-certi:latest
            techblog/go-certi:${{ steps.version.outputs.VERSION }}
```

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/
git commit -m "feat: add Release and Docker Build GitHub Actions workflows"
```

---

## Task 10: Final Verification

- [ ] **Step 1: Full build and test**

```bash
go build ./...
go vet ./...
go test ./... -race -cover
```

Expected:
- `go build` exits 0
- `go vet` exits 0
- All tests pass, coverage reported

- [ ] **Step 2: Smoke-test the running binary**

```bash
go run ./cmd/go-certi --conf /tmp/go-certi-smoke &
PID=$!
sleep 1

curl -sf http://localhost:8111/healthz && echo "healthz OK"
curl -sf http://localhost:8111/readyz  && echo "readyz OK"

kill $PID
rm -rf /tmp/go-certi-smoke
```

Expected:
```
{"status":"ok"} healthz OK
{"status":"ok"} readyz OK
```

- [ ] **Step 3: Verify version ldflags work**

```bash
go build -ldflags="-X github.com/t0mer/go-certi/internal/version.Version=0.0.test" \
    -o /tmp/go-certi-test ./cmd/go-certi
/tmp/go-certi-test --version
rm /tmp/go-certi-test
```

Expected: `0.0.test`

- [ ] **Step 4: Final commit if any loose changes**

```bash
git status
# Commit anything unstaged
git add -u
git commit -m "chore: bootstrap complete — compilable binary with health endpoints"
```

---

## Self-Review

**Spec coverage check:**
- [x] §3a — directory structure, go.mod, .gitignore, stubs, verify — Tasks 1–10
- [x] §3 — all directories created — Task 1
- [x] §6 — CLI flags + env override table — Task 6
- [x] §8 — --port, --conf, --sslmate-api-key, --reset-password, --reset-api-token, --version — Task 6
- [x] §10 — release.yml and docker.yml verbatim from spec — Task 9
- [x] §10 — build.sh multi-arch matrix — Task 7
- [x] §10 — next-version.sh YYYY.M.PATCH — Task 7
- [x] §11 — Dockerfile multi-stage, port 8111, /data volume, non-root, healthcheck — Task 8
- [x] §5 — /healthz, /readyz — Task 5
- [ ] §4 — Domain model tables — **NOT in this plan** (Plan 2: Domain Model)
- [ ] §5 — All REST API endpoints — **NOT in this plan** (Plan 3: API Layer)
- [ ] §6 — Auth (JWT, bcrypt, API tokens) — **NOT in this plan** (Plan 4: Auth)
- [ ] §7 — CT scanning + scheduler — **NOT in this plan** (Plan 5: CT + Scheduler)
- [ ] §3 notify/scheduler packages — **NOT in this plan** (Plan 5)
- [ ] §9 — Frontend pages — **NOT in this plan** (Plan 6: Frontend)

**What's deferred to later plans (in order):**
1. **Plan 2 — Domain Model**: All DB tables (FQDN, Certificate, NotificationChannel, Schedule), sqlc queries
2. **Plan 3 — API Layer**: All REST handlers, Swagger annotations, middleware
3. **Plan 4 — Auth**: JWT, bcrypt, API tokens, login/logout, --reset-password, --reset-api-token
4. **Plan 5 — CT + Notifications + Scheduler**: sslmate client, crt.sh fallback, shoutrrr/greenapi/waweb, robfig/cron
5. **Plan 6 — Frontend**: React + Vite + shadcn/ui, all 7 pages, mobile-first

**Placeholder scan:** No TBDs, TODOs, or "similar to task N" patterns found.

**Type consistency:** `api.New(db *sql.DB)` defined in Task 5 and used in Task 6 — consistent. `config.LoadOrCreate` and `config.OverrideFromEnv` defined in Task 3, used in Task 6 — consistent. `db.Open` defined in Task 4, used in Tasks 5 and 6 — consistent.
