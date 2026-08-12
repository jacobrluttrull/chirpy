-- name: GetAllChirps :many
Select * from chirps
order by created_at asc;

-- name: GetChirpsByAuthor :many
Select * from chirps
where user_id = $1
order by created_at asc;