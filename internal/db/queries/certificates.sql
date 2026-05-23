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
    (id, fqdn_id, serial, issuer_ca, issuer_name, subject_cn, sans, not_before, not_after, discovered_at, source, revoked)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: ListDistinctCAs :many
SELECT DISTINCT issuer_ca FROM certificates WHERE issuer_ca != '' ORDER BY issuer_ca;
