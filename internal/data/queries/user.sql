-- name: CreateUser :one
INSERT INTO users (name, email, password, status, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = ? LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ? LIMIT 1;

-- name: UpdateUser :one
UPDATE users
SET name = ?, email = ?, status = ?, updated_at = ?
WHERE id = ?
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;

-- name: ListUsers :many
SELECT * FROM users
WHERE (? IS NULL OR id = ?)
  AND (? IS NULL OR name LIKE CONCAT('%', ?, '%'))
  AND (? IS NULL OR email LIKE CONCAT('%', ?, '%'))
  AND (? IS NULL OR status = ?)
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: CountUsers :one
SELECT COUNT(*) as count FROM users
WHERE (? IS NULL OR id = ?)
  AND (? IS NULL OR name LIKE CONCAT('%', ?, '%'))
  AND (? IS NULL OR email LIKE CONCAT('%', ?, '%'))
  AND (? IS NULL OR status = ?);

-- name: ExistsByEmail :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = ?) as exists;
