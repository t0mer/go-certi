CREATE TABLE IF NOT EXISTS notification_channels (
    id         TEXT NOT NULL PRIMARY KEY,
    name       TEXT NOT NULL,
    type       TEXT NOT NULL CHECK (type IN ('shoutrrr','greenapi','waweb')),
    config     TEXT NOT NULL DEFAULT '{}',
    enabled    BOOLEAN NOT NULL DEFAULT 1 CHECK (enabled IN (0,1)),
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now'))
);
