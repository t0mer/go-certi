-- name: GetSettings :one
SELECT * FROM settings WHERE id = 1;

-- name: UpdateSettings :one
UPDATE settings
SET auth_enabled                 = ?,
    username                     = ?,
    password_hash                = ?,
    api_token_protection_enabled = ?,
    api_token                    = ?,
    theme                        = ?,
    sslmate_api_key              = ?,
    default_schedule_id          = ?,
    updated_at                   = strftime('%Y-%m-%dT%H:%M:%SZ','now')
WHERE id = 1
RETURNING *;
