# Backend integration test results (T26 / #52)

## Scope
Key backend workflows (US-1..US-6) exercised end to end via the HTTP API against a real PostgreSQL test database (build tag `integration`).

## Commands
- Integration suite (documented): `(cd backend && go test -tags=integration ./...)`
- Unit tests (no database): `(cd backend && go test ./...)`
- Checks: `(cd backend && gofmt -l . && go vet ./...)`
- Contract: `bunx @redocly/cli lint openapi.yaml`

## Environment used for this run
- Test DB: `postgres://postgres:***@localhost:15432/learn_ai_engineering_test` (`pg-test`, postgres:16-alpine); `TEST_DATABASE_URL` set to that isolated test database.
- `CGO_ENABLED=0` was required in this execution environment (no `gcc`); the documented commands are unchanged.

## Results
- `(cd backend && CGO_ENABLED=0 TEST_DATABASE_URL=... go test -tags=integration -p 1 ./...)` — **exit 0**. All integration-tagged packages passed: `handler`, `middleware` (cached/no integration in this run), `migrate`, `repo`, `seed`, `service`.
- `TestKeyWorkflows_EndToEnd` and `TestKeyWorkflows_CanRunTwice` (`internal/handler`) — **pass** (cover browsing published content, register/login/session/logout, quiz start/submit, progress save/read, quiz history read, unpublished content hidden in public responses, foreign-question submission rejected with `validation_error`).
- `-p 1` (serial package execution) used for the full-suite run because integration tests in multiple packages reset/apply schema on the same test database; a parallel package run against a shared test DB produced schema-interference failures unrelated to #52. Handler integration tests (including `TestMaintainerSeedAndDraftPublishFlow_EndToEnd` and key workflows) pass together.
- `(cd backend && CGO_ENABLED=0 go test ./...)` — **exit 0** (unit tests run without a database).
- `(cd backend && gofmt -l .)` — **exit 0** (no output). `(cd backend && CGO_ENABLED=0 go vet ./...)` — **exit 0**.
- `bunx @redocly/cli lint openapi.yaml` — **exit 0** (valid; 4 pre-existing warnings).

## Notes / skipped
- No load/performance tests (out of scope).
- Repository-level SQL tests remain in `T21b/T22b` packages; this artifact covers the HTTP-API end-to-end workflows added for T26.
