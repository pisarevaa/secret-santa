-- name: CreateGroup :one
INSERT INTO groups (invite_code, title, organizer_id, status) VALUES (?, ?, ?, 'open') RETURNING *;

-- name: GetGroupByInviteCode :one
SELECT * FROM groups WHERE invite_code = ?;

-- name: GetGroupByID :one
SELECT * FROM groups WHERE id = ?;

-- name: DrawGroup :execresult
UPDATE groups SET status = 'drawn', drawn_at = CURRENT_TIMESTAMP WHERE id = ? AND status = 'open';
