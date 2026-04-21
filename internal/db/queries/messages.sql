-- name: CreateMessage :one
INSERT INTO messages (group_id, sender_id, recipient_id, direction, body) VALUES (?, ?, ?, ?, ?) RETURNING *;

-- name: ListChatMessages :many
SELECT * FROM messages
WHERE group_id = ?
  AND ((sender_id = ? AND recipient_id = ?) OR (sender_id = ? AND recipient_id = ?))
ORDER BY created_at ASC
LIMIT 50;
