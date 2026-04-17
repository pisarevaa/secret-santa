-- name: CreateMembership :one
INSERT INTO memberships (group_id, user_id, wishlist) VALUES (?, ?, ?) RETURNING *;

-- name: GetMembership :one
SELECT * FROM memberships WHERE id = ?;

-- name: GetMembershipByGroupAndUser :one
SELECT * FROM memberships WHERE group_id = ? AND user_id = ?;

-- name: ListMembershipsByGroup :many
SELECT * FROM memberships WHERE group_id = ?;

-- name: UpdateWishlist :exec
UPDATE memberships SET wishlist = ? WHERE id = ?;

-- name: SetRecipient :exec
UPDATE memberships SET recipient_id = ? WHERE group_id = ? AND user_id = ?;

-- name: GetMyRecipient :one
SELECT u.name, m.wishlist
FROM memberships m
JOIN users u ON u.id = m.recipient_id
WHERE m.group_id = ? AND m.user_id = ? AND m.recipient_id IS NOT NULL;

-- name: CountMembersByGroup :one
SELECT COUNT(*) FROM memberships WHERE group_id = ?;
