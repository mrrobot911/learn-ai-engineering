# Frontend integration test results (T31b / #86)

## Scope
Frontend integration tests for accounts, progress and quiz history (US-4, US-5, US-6) exercised via the HTTP API client against a real backend and PostgreSQL test database. Covers T31a (content/quiz, `contentAndQuiz.test.ts`) and T31b (accounts/progress/history, `accountsProgressHistory.test.ts`).

## Commands
- Frontend integration suite (documented, README section 9): `bash scripts/frontend-integration.sh`
  - Starts an isolated test database (`learn_ai_fe_test` by default, never the dev database), builds/starts a test backend, applies migrations and seed, inserts required fixtures, then runs `(cd frontend && bun run test:integration)`.
- Frontend checks (no backend required): `(cd frontend && bun run lint && bun run typecheck && bun run format:check)`
- Frontend unit tests: `(cd frontend && bun run test)`

## Environment used for this run
- Docker 29.9.2 with Docker Compose 5.6.0; Bun 1.4.2.
- Test backend: `http://127.0.0.1:18080` (isolated container `learn-ai-fe-test-api`), test DB reset by the harness script before each run.

## Results
- Run 1 (`bash scripts/frontend-integration.sh`, clean DB) — **exit 0**.
  - `tests/integration/contentAndQuiz.test.ts` — **9 passed**.
  - `tests/integration/accountsProgressHistory.test.ts` — **5 passed**.
  - Total: **2 files, 14 tests passed**.
- Run 2 (`bash scripts/frontend-integration.sh`, clean DB) — **exit 0**.
  - `tests/integration/contentAndQuiz.test.ts` — **9 passed**.
  - `tests/integration/accountsProgressHistory.test.ts` — **5 passed**.
  - Total: **2 files, 14 tests passed**.
- The suite starts from a clean database on each run (the harness drops/recreates the isolated test database) and **passed twice in a row**.
- `(cd frontend && bun run lint)` — **exit 0**.
- `(cd frontend && bun run typecheck)` — **exit 0**.
- `(cd frontend && bun run format:check)` — **exit 0** (Prettier).

## Coverage (T31b / #86 acceptance criteria)
- Register, login, current session and logout, including duplicate email (409) and CSRF refresh and retry (in-memory CSRF cleared with session cookie preserved; state-changing request retried after `/auth/session` refresh).
- Progress mark and unmark survive a reload (client state reset with session cookie preserved) and are isolated between two users.
- Logged-in quiz submission stores an attempt; history list and detail show the stored snapshot (question text, selected/correct answer text, correctness); another user receives 404 for the attempt.

## Notes / skipped
- Browser automation and visual regression tests are out of scope (per issue).
- Backend integration tests are covered by #52.
- The CSRF refresh-and-retry test exercises the client retry path against a real backend by clearing in-memory CSRF while keeping the session cookie; it does not forge backend CSRF internals.
