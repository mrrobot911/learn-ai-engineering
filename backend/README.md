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

## Testing

```bash
go test ./...
```
