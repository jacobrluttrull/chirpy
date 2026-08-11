-- name: CreateRefreshToken :one
insert into refresh_tokens (token, created_at, updated_at, user_id, expires_at)
values (
    $1,
    NOW(),
    NOW(),
    $2,
    $3
)
RETURNING *;

-- name: GetUserFromRefreshToken :one
select users.* from users
join refresh_tokens on refresh_tokens.user_id = users.id
where refresh_tokens.token = $1
  and refresh_tokens.revoked_at is null
  and refresh_tokens.expires_at > NOW();

-- name: RevokeRefreshToken :exec
update refresh_tokens
set revoked_at = NOW(), updated_at = NOW()
where token = $1;
