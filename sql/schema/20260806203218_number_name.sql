-- +goose Up
create table users (
    id uuid primary key,
    created_at timestamp with time zone not null default now(),
    updated_at timestamp with time zone not null default now(),
    email text not null unique);

-- +goose Down
drop table users;
