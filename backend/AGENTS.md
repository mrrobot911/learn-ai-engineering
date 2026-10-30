# Backend rules (Go)

## Commands
Run from the repository root as `(cd backend && ...)`.
- test: `go test ./...` (one test: `-run TestName`); integration: `go test -tags=integration ./...` (needs `docker compose up -d db`)
- checks: `gofmt -l . && go vet ./...`; format: `gofmt -w .`

## Rules
- Handlers stay thin, logic lives in services, SQL only in the repository layer. Errors have the shape defined in `openapi.yaml`.
- Database changes go through numbered migration files, never by hand. Ask before destructive migrations (drop, rename, data rewrite).
- Integration tests (build tag `integration`) need a running test database; never point them at the development database.
- No extra skills. Use `gopls` for Go navigation and `context7` for library docs at the pinned versions.

## Gotchas
- Quiz endpoints must not return the correct option before the attempt is submitted.
- Content has `draft` and `published` status; public endpoints return only `published`. Agents and tools may create drafts but never publish.
- Users read and change only their own progress and attempts; check ownership in every protected endpoint.
- Quiz grading happens on the server. An unanswered question, or an option that does not belong to its question, counts as incorrect; reject only a submission that contains a question not belonging to the topic.
- Anonymous quiz results are returned but not stored. Submission works with or without a session and stores the attempt only when a valid session exists.
- Seed loading must be repeatable and must not duplicate existing content.
- Passwords are stored only as salted hashes; never return or log `password_hash` or passwords.
- The session cookie is `HttpOnly` and `SameSite`; state-changing requests need cross-site request forgery protection.
