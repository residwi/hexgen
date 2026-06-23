# __PROJECT_NAME__

A Go API service generated from
[go-api-project-template](https://github.com/residwi/go-api-project-template)
via `gen`. Feature-based clean architecture (vertical slices) over
`net/http`, PostgreSQL (pgx/v5), and Redis.

## What's included

- `cmd/api/` — HTTP server entry point
- `internal/core/`, `internal/platform/`, `internal/middleware/` — shared infrastructure
- `internal/features/auth/` + `internal/features/user/` — a working auth + user slice
- `db/migrations/` — `users` table + shared triggers (goose)
- Full unit and integration tests (`make test`, integration tests require Docker)

## Quick start

```bash
go mod tidy
make setup            # installs mockery, goose, air, golangci-lint
cp .env.example .env  # then edit DATABASE_URL etc.
make migrate-up
make seed             # admin user: admin@example.com / admin123  (change this!)
make run
```

## Add a feature

Create a package under `internal/features/<name>/` (handler, service, repository,
model, dto, routes), register it in `internal/server/router.go`, and add a
migration with `make migrate-create name=create_<name>`.
