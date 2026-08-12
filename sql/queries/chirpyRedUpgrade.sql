-- name: UpgradeUserToChirpyRed :one
update users
set is_chirpy_red = true, updated_at = NOW()
where id = $1
returning *;