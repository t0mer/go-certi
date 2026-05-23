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
