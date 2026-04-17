-- name: CreateMessage :one
INSERT INTO messages (group_id, sender_id, recipient_id, direction, body) VALUES (?, ?, ?, ?, ?) RETURNING *;

-- name: ListMessages :many
SELECT * FROM messages
WHERE group_id = ? AND sender_id = ? AND recipient_id = ? AND direction = ?
ORDER BY created_at DESC
LIMIT 50;

-- name: ListMessagesBefore :many
SELECT * FROM messages
WHERE group_id = ? AND sender_id = ? AND recipient_id = ? AND direction = ? AND id < ?
ORDER BY created_at DESC
LIMIT 50;
