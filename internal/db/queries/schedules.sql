-- name: ListSchedules :many
SELECT * FROM schedules ORDER BY name;

-- name: GetSchedule :one
SELECT * FROM schedules WHERE id = ?;

-- name: GetDefaultSchedule :one
SELECT * FROM schedules WHERE is_default = 1 LIMIT 1;

-- name: CreateSchedule :one
INSERT INTO schedules (id, name, cron_expr, is_default, enabled)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateSchedule :one
UPDATE schedules
SET name      = ?,
    cron_expr = ?,
    is_default = ?,
    enabled   = ?,
    updated_at = strftime('%Y-%m-%dT%H:%M:%SZ','now')
WHERE id = ?
RETURNING *;

-- name: DeleteSchedule :exec
DELETE FROM schedules WHERE id = ?;

-- name: UnsetDefaultSchedules :exec
UPDATE schedules SET is_default = 0 WHERE is_default = 1;
