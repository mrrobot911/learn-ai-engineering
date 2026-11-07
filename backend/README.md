# Backend

Go API server for the AI Engineering Learning Platform.

## Structure

- `cmd/api/` - Application entry point
- `internal/config/` - Configuration loading from environment
- `internal/handler/` - HTTP handlers (thin)
- `internal/middleware/` - HTTP middleware
- `internal/service/` - Business logic layer
- `internal/repo/` - Data access layer (SQL only here)

## Running

```bash
go run ./cmd/api
```

When `DATABASE_URL` is set, the server applies migrations and loads versioned seed content from `backend/seed/` on start. Seed loading is repeatable: existing records are skipped.

## Seed content

Seed files live under `backend/seed/` and are reviewed in diffs. To load seed into a configured database (migrations are applied first):

```bash
go run ./cmd/seed
```

## Testing

```bash
go test ./...
```
