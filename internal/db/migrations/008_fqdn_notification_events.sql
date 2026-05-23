-- Add per-FQDN notification event configuration
ALTER TABLE fqdns ADD COLUMN notification_events TEXT NOT NULL DEFAULT '["new_cert"]';
ALTER TABLE fqdns ADD COLUMN expiry_threshold_days INTEGER NOT NULL DEFAULT 10;

-- Log sent notifications to prevent duplicates within a 24h window
CREATE TABLE IF NOT EXISTS notification_log (
    id      TEXT NOT NULL PRIMARY KEY,
    fqdn_id TEXT NOT NULL,
    cert_id TEXT NOT NULL,
    event   TEXT NOT NULL,
    sent_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now'))
);
CREATE INDEX IF NOT EXISTS idx_notif_log_lookup ON notification_log(fqdn_id, cert_id, event);
