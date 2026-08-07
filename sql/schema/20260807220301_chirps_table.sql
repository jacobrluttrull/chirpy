-- +goose Up
create table chirps
(
    id         uuid primary key,
    created_at timestamp with time zone not null default now(),
    updated_at timestamp with time zone not null default now(),
    body       text                     not null,
    user_id    uuid                     not null references users (id) on delete cascade
);
-- +goose Down
drop table chirps;
