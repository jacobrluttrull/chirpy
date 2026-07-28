# Chirpy

A simple Go HTTP server, built as part of the Boot.dev backend course.

## Running

```sh
go run .
```

The server listens on `:8080`.

## Routes

- `/app/` — serves static files (the app itself)
- `/assets/` — serves static assets (e.g. images)
- `/healthz` — readiness endpoint, returns `200 OK`
