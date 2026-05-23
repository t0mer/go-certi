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
