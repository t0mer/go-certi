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
