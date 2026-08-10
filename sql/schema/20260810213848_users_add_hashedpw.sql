-- +goose Up
SELECT 'up SQL query';
ALTER TABLE users ADD COLUMN hashed_password TEXT NOT NULL DEFAULT 'unset';

-- +goose Down
SELECT 'down SQL query';
ALTER TABLE users DROP COLUMN hashed_password;
