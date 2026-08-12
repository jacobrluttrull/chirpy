# Chirpy

Chirpy is a REST API for a Twitter-style social feed, written in Go. Users sign up, log in, and post short "chirps"; chirps can be listed, filtered by author, sorted, and deleted, and a mock payments webhook upgrades users to a paid "Chirpy Red" tier.

It was built as the capstone project for [Boot.dev](https://www.boot.dev)'s backend course, as a way to practice building a production-shaped HTTP API in Go from scratch: JWT auth with refresh tokens, password hashing, a Postgres-backed data layer generated with `sqlc`, versioned migrations with `goose`, and a resource-oriented REST design — without a web framework, using only the standard library's `net/http`.

## Features

- Email/password auth with bcrypt-hashed passwords
- JWT access tokens + long-lived, revocable refresh tokens
- CRUD on chirps (140-character posts), with profanity filtering
- Listing chirps with optional `author_id` filtering and `asc`/`desc` sorting
- Webhook endpoint for a mock third-party payment processor to upgrade users
- API-key- and JWT-protected routes, with ownership checks (you can only delete your own chirps)

## Tech stack

- Go (`net/http`, no framework)
- PostgreSQL
- [`sqlc`](https://sqlc.dev/) for type-safe generated queries
- [`goose`](https://github.com/pressly/goose) for schema migrations

## Getting started

### Prerequisites

- Go 1.26+
- PostgreSQL
- [`goose`](https://github.com/pressly/goose) (only needed if you want to run/modify migrations)

### Setup

1. Clone the repo and create a Postgres database for it.
2. Copy the example env file and fill in your own values:

   ```sh
   cp .env.example .env
   ```

   | Variable     | Description                                                              |
   | ------------ | ------------------------------------------------------------------------- |
   | `DB_URL`     | Postgres connection string, e.g. `postgres://user:pass@localhost:5432/chirpy?sslmode=disable` |
   | `PLATFORM`   | Set to `dev` to enable the `/admin/reset` endpoint                        |
   | `JWT_SECRET` | Secret used to sign access tokens (generate with `openssl rand -base64 64`) |
   | `POLKA_KEY`  | API key expected from the payments webhook caller                         |

3. Run migrations:

   ```sh
   goose -dir sql/schema postgres "$DB_URL" up
   ```

4. Start the server:

   ```sh
   go run .
   ```

   The server listens on `:8080`.

### Running tests

```sh
go test ./...
```

## API overview

| Method | Path                     | Description                                      |
| ------ | ------------------------ | ------------------------------------------------- |
| GET    | `/api/healthz`           | Health check                                       |
| POST   | `/api/users`              | Create a user                                      |
| PUT    | `/api/users`              | Update the authenticated user's email/password     |
| POST   | `/api/login`              | Log in, returns access + refresh tokens            |
| POST   | `/api/refresh`            | Exchange a refresh token for a new access token    |
| POST   | `/api/revoke`             | Revoke a refresh token                             |
| POST   | `/api/chirps`             | Create a chirp (auth required)                     |
| GET    | `/api/chirps`             | List chirps — supports `?author_id=` and `?sort=asc\|desc` |
| GET    | `/api/chirps/{chirpID}`   | Get a single chirp                                 |
| DELETE | `/api/chirps/{chirpID}`   | Delete a chirp you own (auth required)              |
| POST   | `/api/polka/webhooks`     | Payments webhook that upgrades a user to Chirpy Red |

Admin/dev-only:

| Method | Path              | Description                              |
| ------ | ----------------- | ----------------------------------------- |
| GET    | `/admin/metrics`  | View app file-server hit count            |
| POST   | `/admin/reset`    | Reset the database (only when `PLATFORM=dev`) |
