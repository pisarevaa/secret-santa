-- name: CreateMagicLink :exec
INSERT INTO magic_links (token, email, expires_at) VALUES (?, ?, ?);

-- name: GetMagicLink :one
SELECT * FROM magic_links WHERE token = ? AND used_at IS NULL AND expires_at > CURRENT_TIMESTAMP;

-- name: MarkMagicLinkUsed :exec
UPDATE magic_links SET used_at = CURRENT_TIMESTAMP WHERE token = ?;

-- name: CreateSession :exec
INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?);

-- name: GetSession :one
SELECT * FROM sessions WHERE token = ? AND expires_at > CURRENT_TIMESTAMP;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE token = ?;
