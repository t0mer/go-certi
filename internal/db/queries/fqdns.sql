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
