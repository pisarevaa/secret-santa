-- name: CreateUser :one
INSERT INTO users (email, name) VALUES (?, ?) RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ?;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = ?;

-- name: UpdateUserName :exec
UPDATE users SET name = ? WHERE id = ?;
