# Domain Model Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create all domain tables (schedules, fqdns, notification_channels, fqdn_channels, certificates) via SQLite migrations and generate type-safe Go query code with sqlc.

**Architecture:** New migration files extend the existing hand-rolled migration runner (internal/db). sqlc reads the schema from migration files and query files from internal/db/queries/, generating a type-safe `*models.Queries` in internal/models/. All API handlers will use `models.New(db)` for DB access.

**Tech Stack:** modernc.org/sqlite, sqlc v1.28+ (via `go tool`), github.com/google/uuid v1.6+

---

## File Map

| File | Purpose |
|---|---|
| `sqlc.yaml` | sqlc config: sqlite engine, queries dir, schema dir, output to internal/models |
| `internal/db/migrations/002_schedules.sql` | schedules table + seed default schedule |
| `internal/db/migrations/003_fqdns.sql` | fqdns table |
| `internal/db/migrations/004_channels.sql` | notification_channels table |
| `internal/db/migrations/005_fqdn_channels.sql` | fqdn_channels junction table |
| `internal/db/migrations/006_certificates.sql` | certificates table |
| `internal/db/queries/schedules.sql` | sqlc queries for schedules |
| `internal/db/queries/fqdns.sql` | sqlc queries for fqdns |
| `internal/db/queries/channels.sql` | sqlc queries for channels + fqdn_channel links |
| `internal/db/queries/certificates.sql` | sqlc queries for certificates |
| `internal/db/queries/settings.sql` | sqlc queries for the settings singleton |
| `internal/models/` | sqlc-generated Go types (committed, not gitignored) |
| `internal/models/models_test.go` | integration tests for all CRUD operations |

---

## Task 1: sqlc Tool Setup

**Files:**
- Modify: `go.mod` (add sqlc as tool)
- Create: `sqlc.yaml`

- [ ] **Step 1: Add sqlc as a Go tool**

```bash
cd /opt/dev/go-certi
go get -tool github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

Expected: go.mod gains a `tool` section, go.sum updated.

- [ ] **Step 2: Verify sqlc is runnable**

```bash
go tool sqlc version
```

Expected: prints sqlc version like `v1.28.0`.

- [ ] **Step 3: Create sqlc.yaml**

Create `/opt/dev/go-certi/sqlc.yaml`:

```yaml
version: "2"
sql:
  - engine: "sqlite"
    queries: "internal/db/queries/"
    schema: "internal/db/migrations/"
    gen:
      go:
        package: "models"
        out: "internal/models"
        emit_json_tags: true
        emit_pointers_for_null_types: true
        overrides:
          - db_type: "BOOLEAN"
            go_type: "bool"
          - db_type: "boolean"
            go_type: "bool"
```

- [ ] **Step 4: Commit**

```bash
git add go.mod go.sum sqlc.yaml
git commit -m "chore: add sqlc tool and config"
```

---

## Task 2: Migrations 002–003 (schedules + fqdns)

**Files:**
- Create: `internal/db/migrations/002_schedules.sql`
- Create: `internal/db/migrations/003_fqdns.sql`

- [ ] **Step 1: Create 002_schedules.sql**

Create `internal/db/migrations/002_schedules.sql`:

```sql
CREATE TABLE IF NOT EXISTS schedules (
    id         TEXT NOT NULL PRIMARY KEY,
    name       TEXT NOT NULL,
    cron_expr  TEXT NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT 0 CHECK (is_default IN (0,1)),
    enabled    BOOLEAN NOT NULL DEFAULT 1 CHECK (enabled IN (0,1)),
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now'))
);

-- Seed the default schedule
INSERT OR IGNORE INTO schedules (id, name, cron_expr, is_default, enabled)
VALUES ('00000000-0000-0000-0000-000000000001', 'Every 2 hours', '@every 2h', 1, 1);

-- Point the settings row to the default schedule
UPDATE settings
SET default_schedule_id = '00000000-0000-0000-0000-000000000001'
WHERE id = 1 AND default_schedule_id IS NULL;
```

- [ ] **Step 2: Create 003_fqdns.sql**

Create `internal/db/migrations/003_fqdns.sql`:

```sql
CREATE TABLE IF NOT EXISTS fqdns (
    id                    TEXT NOT NULL PRIMARY KEY,
    fqdn                  TEXT NOT NULL UNIQUE,
    include_subdomains    BOOLEAN NOT NULL DEFAULT 0 CHECK (include_subdomains IN (0,1)),
    enabled               BOOLEAN NOT NULL DEFAULT 1 CHECK (enabled IN (0,1)),
    notifications_enabled BOOLEAN NOT NULL DEFAULT 1 CHECK (notifications_enabled IN (0,1)),
    schedule_id           TEXT REFERENCES schedules(id) ON DELETE SET NULL,
    created_at            TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
    updated_at            TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now'))
);

CREATE INDEX IF NOT EXISTS idx_fqdns_enabled ON fqdns(enabled);
```

- [ ] **Step 3: Verify migrations run**

```bash
cd /opt/dev/go-certi
go test ./internal/db/ -v -run TestOpen
```

Expected: both `TestOpen_CreatesAndMigrates` and `TestOpen_IdempotentMigrations` PASS.

- [ ] **Step 4: Commit**

```bash
git add internal/db/migrations/002_schedules.sql internal/db/migrations/003_fqdns.sql
git commit -m "feat: add schedules and fqdns migrations"
```

---

## Task 3: Migrations 004–006 (channels, fqdn_channels, certificates)

**Files:**
- Create: `internal/db/migrations/004_channels.sql`
- Create: `internal/db/migrations/005_fqdn_channels.sql`
- Create: `internal/db/migrations/006_certificates.sql`

- [ ] **Step 1: Create 004_channels.sql**

Create `internal/db/migrations/004_channels.sql`:

```sql
CREATE TABLE IF NOT EXISTS notification_channels (
    id         TEXT NOT NULL PRIMARY KEY,
    name       TEXT NOT NULL,
    type       TEXT NOT NULL CHECK (type IN ('shoutrrr','greenapi','waweb')),
    config     TEXT NOT NULL DEFAULT '{}',
    enabled    BOOLEAN NOT NULL DEFAULT 1 CHECK (enabled IN (0,1)),
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now'))
);
```

- [ ] **Step 2: Create 005_fqdn_channels.sql**

Create `internal/db/migrations/005_fqdn_channels.sql`:

```sql
CREATE TABLE IF NOT EXISTS fqdn_channels (
    fqdn_id    TEXT NOT NULL REFERENCES fqdns(id) ON DELETE CASCADE,
    channel_id TEXT NOT NULL REFERENCES notification_channels(id) ON DELETE CASCADE,
    PRIMARY KEY (fqdn_id, channel_id)
);

CREATE INDEX IF NOT EXISTS idx_fqdn_channels_channel ON fqdn_channels(channel_id);
```

- [ ] **Step 3: Create 006_certificates.sql**

Create `internal/db/migrations/006_certificates.sql`:

```sql
CREATE TABLE IF NOT EXISTS certificates (
    id            TEXT NOT NULL PRIMARY KEY,
    fqdn_id       TEXT NOT NULL REFERENCES fqdns(id) ON DELETE CASCADE,
    serial        TEXT NOT NULL,
    issuer_ca     TEXT NOT NULL DEFAULT '',
    subject_cn    TEXT NOT NULL DEFAULT '',
    sans          TEXT NOT NULL DEFAULT '[]',
    not_before    TEXT NOT NULL,
    not_after     TEXT NOT NULL,
    discovered_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
    source        TEXT NOT NULL DEFAULT 'sslmate',
    UNIQUE (fqdn_id, serial)
);

CREATE INDEX IF NOT EXISTS idx_certs_fqdn_id    ON certificates(fqdn_id);
CREATE INDEX IF NOT EXISTS idx_certs_not_after  ON certificates(not_after);
CREATE INDEX IF NOT EXISTS idx_certs_issuer_ca  ON certificates(issuer_ca);
CREATE INDEX IF NOT EXISTS idx_certs_discovered ON certificates(discovered_at);
```

- [ ] **Step 4: Verify all 6 migrations run**

```bash
cd /opt/dev/go-certi
go test ./internal/db/ -v -run TestOpen
```

Expected: PASS. The schema_migrations table should now have 6 rows.

Add a quick verification:
```bash
cd /opt/dev/go-certi && cat > /tmp/verify_schema.go << 'EOF'
//go:build ignore

package main

import (
    "database/sql"
    "fmt"
    "os"
    _ "modernc.org/sqlite"
)

func main() {
    db, err := sql.Open("sqlite", "file:/tmp/schema_test.db?_pragma=foreign_keys(ON)")
    if err != nil { panic(err) }
    defer db.Close()
    defer os.Remove("/tmp/schema_test.db")

    tables := []string{"settings","schedules","fqdns","notification_channels","fqdn_channels","certificates","schema_migrations"}
    for _, t := range tables {
        var n int
        db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", t).Scan(&n)
        if n == 0 { fmt.Println("MISSING:", t); os.Exit(1) }
        fmt.Println("OK:", t)
    }
}
EOF
# This just confirms the migration SQL is valid; actual test is via go test above
```

- [ ] **Step 5: Commit**

```bash
git add internal/db/migrations/
git commit -m "feat: add channels, fqdn_channels, and certificates migrations"
```

---

## Task 4: sqlc Query Files

**Files:**
- Create: `internal/db/queries/schedules.sql`
- Create: `internal/db/queries/fqdns.sql`
- Create: `internal/db/queries/channels.sql`
- Create: `internal/db/queries/certificates.sql`
- Create: `internal/db/queries/settings.sql`
- Delete: `internal/db/queries/.gitkeep` (no longer needed once .sql files exist)

- [ ] **Step 1: Create schedules.sql**

Create `internal/db/queries/schedules.sql`:

```sql
-- name: ListSchedules :many
SELECT * FROM schedules ORDER BY name;

-- name: GetSchedule :one
SELECT * FROM schedules WHERE id = ?;

-- name: GetDefaultSchedule :one
SELECT * FROM schedules WHERE is_default = 1 LIMIT 1;

-- name: CreateSchedule :one
INSERT INTO schedules (id, name, cron_expr, is_default, enabled)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateSchedule :one
UPDATE schedules
SET name      = ?,
    cron_expr = ?,
    is_default = ?,
    enabled   = ?,
    updated_at = strftime('%Y-%m-%dT%H:%M:%SZ','now')
WHERE id = ?
RETURNING *;

-- name: DeleteSchedule :exec
DELETE FROM schedules WHERE id = ?;

-- name: UnsetDefaultSchedules :exec
UPDATE schedules SET is_default = 0 WHERE is_default = 1;
```

- [ ] **Step 2: Create fqdns.sql**

Create `internal/db/queries/fqdns.sql`:

```sql
-- name: ListFQDNs :many
SELECT * FROM fqdns ORDER BY fqdn;

-- name: GetFQDN :one
SELECT * FROM fqdns WHERE id = ?;

-- name: GetFQDNByName :one
SELECT * FROM fqdns WHERE fqdn = ?;

-- name: CreateFQDN :one
INSERT INTO fqdns (id, fqdn, include_subdomains, enabled, notifications_enabled, schedule_id)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateFQDN :one
UPDATE fqdns
SET fqdn                  = ?,
    include_subdomains    = ?,
    enabled               = ?,
    notifications_enabled = ?,
    schedule_id           = ?,
    updated_at            = strftime('%Y-%m-%dT%H:%M:%SZ','now')
WHERE id = ?
RETURNING *;

-- name: DeleteFQDN :exec
DELETE FROM fqdns WHERE id = ?;

-- name: ListEnabledFQDNs :many
SELECT * FROM fqdns WHERE enabled = 1 ORDER BY fqdn;
```

- [ ] **Step 3: Create channels.sql**

Create `internal/db/queries/channels.sql`:

```sql
-- name: ListChannels :many
SELECT * FROM notification_channels ORDER BY name;

-- name: GetChannel :one
SELECT * FROM notification_channels WHERE id = ?;

-- name: CreateChannel :one
INSERT INTO notification_channels (id, name, type, config, enabled)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateChannel :one
UPDATE notification_channels
SET name       = ?,
    type       = ?,
    config     = ?,
    enabled    = ?,
    updated_at = strftime('%Y-%m-%dT%H:%M:%SZ','now')
WHERE id = ?
RETURNING *;

-- name: DeleteChannel :exec
DELETE FROM notification_channels WHERE id = ?;

-- name: GetFQDNChannels :many
SELECT nc.* FROM notification_channels nc
INNER JOIN fqdn_channels fc ON fc.channel_id = nc.id
WHERE fc.fqdn_id = ?
ORDER BY nc.name;

-- name: GetFQDNChannelIDs :many
SELECT channel_id FROM fqdn_channels WHERE fqdn_id = ?;

-- name: AddFQDNChannel :exec
INSERT OR IGNORE INTO fqdn_channels (fqdn_id, channel_id) VALUES (?, ?);

-- name: DeleteFQDNChannels :exec
DELETE FROM fqdn_channels WHERE fqdn_id = ?;
```

- [ ] **Step 4: Create certificates.sql**

Create `internal/db/queries/certificates.sql`:

```sql
-- name: ListCertificates :many
SELECT * FROM certificates ORDER BY discovered_at DESC LIMIT ? OFFSET ?;

-- name: ListCertificatesByFQDN :many
SELECT * FROM certificates WHERE fqdn_id = ?
ORDER BY discovered_at DESC LIMIT ? OFFSET ?;

-- name: CountCertificates :one
SELECT COUNT(*) FROM certificates;

-- name: CountCertificatesByFQDN :one
SELECT COUNT(*) FROM certificates WHERE fqdn_id = ?;

-- name: GetCertificate :one
SELECT * FROM certificates WHERE id = ?;

-- name: GetCertificateBySerial :one
SELECT * FROM certificates WHERE fqdn_id = ? AND serial = ?;

-- name: InsertCertificate :one
INSERT OR IGNORE INTO certificates
    (id, fqdn_id, serial, issuer_ca, subject_cn, sans, not_before, not_after, discovered_at, source)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: ListDistinctCAs :many
SELECT DISTINCT issuer_ca FROM certificates WHERE issuer_ca != '' ORDER BY issuer_ca;
```

- [ ] **Step 5: Create settings.sql**

Create `internal/db/queries/settings.sql`:

```sql
-- name: GetSettings :one
SELECT * FROM settings WHERE id = 1;

-- name: UpdateSettings :one
UPDATE settings
SET auth_enabled                 = ?,
    username                     = ?,
    password_hash                = ?,
    api_token_protection_enabled = ?,
    api_token                    = ?,
    theme                        = ?,
    sslmate_api_key              = ?,
    default_schedule_id          = ?,
    updated_at                   = strftime('%Y-%m-%dT%H:%M:%SZ','now')
WHERE id = 1
RETURNING *;
```

- [ ] **Step 6: Remove .gitkeep from queries dir**

```bash
rm /opt/dev/go-certi/internal/db/queries/.gitkeep
```

- [ ] **Step 7: Commit query files**

```bash
git add internal/db/queries/
git commit -m "feat: add sqlc query definitions for all entities"
```

---

## Task 5: Generate sqlc Code

**Files:**
- Create: `internal/models/` (generated by sqlc — all files)

- [ ] **Step 1: Run sqlc generate**

```bash
cd /opt/dev/go-certi
go tool sqlc generate
```

Expected: Creates files in `internal/models/`:
- `db.go` — DBTX interface + New() constructor
- `models.go` — all struct types
- `schedules.sql.go` — schedule query implementations
- `fqdns.sql.go` — fqdn query implementations
- `channels.sql.go` — channel + fqdn_channel query implementations
- `certificates.sql.go` — certificate query implementations
- `settings.sql.go` — settings query implementations
- `querier.go` — Querier interface

If sqlc errors on query parsing, check that the SQL syntax matches SQLite (not Postgres). Common fix: `?` placeholders (not `$1`).

- [ ] **Step 2: Remove models/.gitkeep if present**

```bash
rm -f /opt/dev/go-certi/internal/models/.gitkeep
```

- [ ] **Step 3: Build to verify generated code compiles**

```bash
cd /opt/dev/go-certi && go build ./...
```

Expected: no errors. If sqlc generated code imports packages not yet in go.mod, run `go mod tidy`.

- [ ] **Step 4: Add google/uuid as direct dependency**

```bash
cd /opt/dev/go-certi
go get github.com/google/uuid@v1.6.0
```

Expected: go.mod updated, uuid appears as direct dep (needed for generating UUIDs in the API layer).

- [ ] **Step 5: Commit generated code**

```bash
git add internal/models/ go.mod go.sum
git commit -m "feat: generate sqlc models and query code for all entities"
```

---

## Task 6: Integration Tests for Domain Model

**Files:**
- Create: `internal/models/models_test.go`

- [ ] **Step 1: Write failing test**

Create `internal/models/models_test.go`:

```go
package models_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/t0mer/go-certi/internal/db"
	"github.com/t0mer/go-certi/internal/models"
)

func openDB(t *testing.T) *models.Queries {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return models.New(conn)
}

func TestScheduleCRUD(t *testing.T) {
	q := openDB(t)
	ctx := context.Background()

	// Create
	sched, err := q.CreateSchedule(ctx, models.CreateScheduleParams{
		ID:        uuid.NewString(),
		Name:      "Test Schedule",
		CronExpr:  "@every 1h",
		IsDefault: false,
		Enabled:   true,
	})
	if err != nil {
		t.Fatalf("CreateSchedule: %v", err)
	}
	if sched.Name != "Test Schedule" {
		t.Errorf("Name = %q, want Test Schedule", sched.Name)
	}

	// Get
	got, err := q.GetSchedule(ctx, sched.ID)
	if err != nil {
		t.Fatalf("GetSchedule: %v", err)
	}
	if got.ID != sched.ID {
		t.Errorf("ID mismatch")
	}

	// List
	list, err := q.ListSchedules(ctx)
	if err != nil {
		t.Fatalf("ListSchedules: %v", err)
	}
	if len(list) < 2 { // seeded default + our new one
		t.Errorf("expected >=2 schedules, got %d", len(list))
	}

	// Update
	updated, err := q.UpdateSchedule(ctx, models.UpdateScheduleParams{
		Name:      "Updated",
		CronExpr:  "@every 4h",
		IsDefault: false,
		Enabled:   true,
		ID:        sched.ID,
	})
	if err != nil {
		t.Fatalf("UpdateSchedule: %v", err)
	}
	if updated.Name != "Updated" {
		t.Errorf("Updated name = %q", updated.Name)
	}

	// Delete
	if err := q.DeleteSchedule(ctx, sched.ID); err != nil {
		t.Fatalf("DeleteSchedule: %v", err)
	}
}

func TestFQDNCRUD(t *testing.T) {
	q := openDB(t)
	ctx := context.Background()

	fqdn, err := q.CreateFQDN(ctx, models.CreateFQDNParams{
		ID:                   uuid.NewString(),
		Fqdn:                 "example.com",
		IncludeSubdomains:    false,
		Enabled:              true,
		NotificationsEnabled: true,
		ScheduleID:           nil,
	})
	if err != nil {
		t.Fatalf("CreateFQDN: %v", err)
	}
	if fqdn.Fqdn != "example.com" {
		t.Errorf("Fqdn = %q", fqdn.Fqdn)
	}

	got, err := q.GetFQDN(ctx, fqdn.ID)
	if err != nil {
		t.Fatalf("GetFQDN: %v", err)
	}
	if got.ID != fqdn.ID {
		t.Error("ID mismatch")
	}

	if err := q.DeleteFQDN(ctx, fqdn.ID); err != nil {
		t.Fatalf("DeleteFQDN: %v", err)
	}
}

func TestChannelCRUD(t *testing.T) {
	q := openDB(t)
	ctx := context.Background()

	ch, err := q.CreateChannel(ctx, models.CreateChannelParams{
		ID:      uuid.NewString(),
		Name:    "Test Telegram",
		Type:    "shoutrrr",
		Config:  `{"url":"telegram://token@telegram?chats=123"}`,
		Enabled: true,
	})
	if err != nil {
		t.Fatalf("CreateChannel: %v", err)
	}
	if ch.Name != "Test Telegram" {
		t.Errorf("Name = %q", ch.Name)
	}

	if err := q.DeleteChannel(ctx, ch.ID); err != nil {
		t.Fatalf("DeleteChannel: %v", err)
	}
}

func TestCertificateInsert(t *testing.T) {
	q := openDB(t)
	ctx := context.Background()

	fqdn, _ := q.CreateFQDN(ctx, models.CreateFQDNParams{
		ID: uuid.NewString(), Fqdn: "cert-test.com",
		Enabled: true, NotificationsEnabled: true,
	})

	cert, err := q.InsertCertificate(ctx, models.InsertCertificateParams{
		ID:          uuid.NewString(),
		FqdnID:      fqdn.ID,
		Serial:      "01:02:03",
		IssuerCa:    "Let's Encrypt",
		SubjectCn:   "cert-test.com",
		Sans:        `["cert-test.com"]`,
		NotBefore:   "2026-01-01T00:00:00Z",
		NotAfter:    "2026-04-01T00:00:00Z",
		DiscoveredAt: "2026-01-01T12:00:00Z",
		Source:      "sslmate",
	})
	if err != nil {
		t.Fatalf("InsertCertificate: %v", err)
	}
	if cert == nil {
		t.Fatal("expected cert row, got nil (INSERT OR IGNORE conflict?)")
	}
	if cert.Serial != "01:02:03" {
		t.Errorf("Serial = %q", cert.Serial)
	}

	// Idempotent insert
	dup, err := q.InsertCertificate(ctx, models.InsertCertificateParams{
		ID: uuid.NewString(), FqdnID: fqdn.ID, Serial: "01:02:03",
		NotBefore: "2026-01-01T00:00:00Z", NotAfter: "2026-04-01T00:00:00Z",
		DiscoveredAt: "2026-01-01T12:00:00Z", Source: "sslmate",
		Sans: `[]`,
	})
	if err != nil {
		t.Fatalf("duplicate InsertCertificate should not error: %v", err)
	}
	if dup != nil {
		t.Error("duplicate insert should return nil (INSERT OR IGNORE)")
	}
}

func TestSettingsGetUpdate(t *testing.T) {
	q := openDB(t)
	ctx := context.Background()

	s, err := q.GetSettings(ctx)
	if err != nil {
		t.Fatalf("GetSettings: %v", err)
	}
	if s.Theme != "system" {
		t.Errorf("default theme = %q", s.Theme)
	}

	updated, err := q.UpdateSettings(ctx, models.UpdateSettingsParams{
		AuthEnabled:               false,
		ApiTokenProtectionEnabled: false,
		Theme:                     "dark",
		SslmateApiKey:             "test-key",
	})
	if err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	if updated.Theme != "dark" {
		t.Errorf("updated theme = %q", updated.Theme)
	}
}
```

- [ ] **Step 2: Run test to confirm it fails**

```bash
cd /opt/dev/go-certi
go test ./internal/models/ -v -run TestScheduleCRUD 2>&1 | head -20
```

Expected: FAIL (package not yet testable, or type mismatches before models generated).
After sqlc generate in Task 5, re-run — should PASS.

- [ ] **Step 3: Run all model tests**

```bash
cd /opt/dev/go-certi
go test ./internal/models/ -v -race
```

Expected: all 5 tests PASS.

If you get a type mismatch (e.g., `ScheduleID` expects `*string` but sqlc generated `sql.NullString`), update the test to use the generated type. Check `internal/models/models.go` for the actual struct field types and adjust test accordingly.

- [ ] **Step 4: Run full test suite**

```bash
cd /opt/dev/go-certi
go test ./... -race
```

Expected: all packages PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/models/models_test.go
git commit -m "test: add domain model integration tests"
```

---

## Self-Review

**Spec coverage (CLAUDE.md §4):**
- [x] FQDN table — Task 2 (003_fqdns.sql)
- [x] Certificate table with dedup key (fqdn_id, serial) — Task 3 (006_certificates.sql)
- [x] NotificationChannel table with type enum — Task 3 (004_channels.sql)
- [x] Schedule table with is_default — Task 2 (002_schedules.sql)
- [x] fqdn_channels junction (channel_ids []uuid) — Task 3 (005_fqdn_channels.sql)
- [x] Settings already exists (001_init.sql) + query in Task 4
- [x] Default schedule seeded — Task 2
- [x] UUIDs as TEXT — all migrations
- [x] SANs as JSON TEXT column — Task 3
- [x] channel config as JSON TEXT column — Task 3
- [x] source field on certs ('sslmate', 'crt.sh') — Task 3

**Not in this plan (deferred to API plan):**
- REST endpoints
- UUID generation in handlers (uses github.com/google/uuid added in Task 5)
