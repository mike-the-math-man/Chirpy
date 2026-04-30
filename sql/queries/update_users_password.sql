-- name: UpdateUserPassword :exec
UPDATE users SET updated_at = NOW(), hashed_password = $1, email = $2 WHERE id = $3;