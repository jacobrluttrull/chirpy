-- name: GetAllChirps :many
Select * from chirps
order by created_at asc;