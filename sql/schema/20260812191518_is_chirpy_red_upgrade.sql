-- +goose Up
SELECT 'up SQL query';

alter table users
add column is_chirpy_red boolean not null default false;


-- +goose Down
SELECT 'down SQL query';

alter table users
drop column is_chirpy_red;
