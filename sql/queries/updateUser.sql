-- name: UpdateUser :one
update users
set email = $1, hashed_password = $2, updated_at = NOW()
where id = $3
returning *;