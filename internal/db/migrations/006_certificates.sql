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
