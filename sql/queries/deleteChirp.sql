-- name: DeleteChirp :exec
delete from chirps
where id = $1;
