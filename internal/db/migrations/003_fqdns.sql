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
