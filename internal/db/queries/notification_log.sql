-- name: LogNotification :exec
INSERT OR IGNORE INTO notification_log (id, fqdn_id, cert_id, event)
VALUES (?, ?, ?, ?);

-- name: WasNotificationSent :one
SELECT COUNT(*) > 0 FROM notification_log
WHERE fqdn_id = ? AND cert_id = ? AND event = ?
  AND sent_at > strftime('%Y-%m-%dT%H:%M:%SZ', datetime('now', '-24 hours'));

-- name: GetCertsExpiringBefore :many
SELECT * FROM certificates
WHERE fqdn_id = ? AND revoked = 0
  AND not_after > ? AND not_after <= ?;

-- name: GetExpiredCerts :many
SELECT * FROM certificates
WHERE fqdn_id = ? AND revoked = 0 AND not_after <= ?;

-- name: GetRevokedCerts :many
SELECT * FROM certificates
WHERE fqdn_id = ? AND revoked = 1;

-- name: GetPreviousCertForFQDN :one
SELECT * FROM certificates
WHERE fqdn_id = ? AND serial != ?
ORDER BY discovered_at DESC LIMIT 1;
