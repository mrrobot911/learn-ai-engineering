## 2026-10-29 — Entry 23: domain error type and HTTP mapping (issue #35)

- **Tool:** OpenCode (software engineer role) implementing groomed issue #35 on branch `issue-35`.
- **Task / context:** issue #35 (T14b) with its criteria and out-of-scope list; `AGENTS.md` (layering, smallest complete change, gotchas); `openapi.yaml` `Error` schema; `product-spec.md` US-3, US-4, US-7; the existing `backend/internal/{config,handler,middleware}` and `cmd/api` layout.
- **Result:** Added `backend/internal/apperr` with the seven `Code` constants and the domain `Error` (`Code`, `Message`, `Details`, wrapped `Err`, `Error`/`Unwrap`). Added `handler.WriteError`, the one place mapping a code to status and JSON body: `validation_error` 400, `authentication_error` 401, `forbidden` and `authorization_error` both 403 with their own code, `not_found` 404, `conflict` 409, `internal_error` 500. Bodies carry only `code`, a non-empty `message` and `details` when present; an unknown code or a non-domain error becomes a generic 500 whose body never repeats the wrapped text. No endpoint, CSRF or authorisation logic added.
- **Accepted / rejected / changed by hand, and why:** left for the maintainer.
- **Verification:**
  - `(cd backend && gofmt -l . && go vet ./... && go test ./...)`: exit 0 (all packages pass; the shell had no gcc, so the check ran with the nix gcc wrapper on PATH)
  - `bunx @redocly/cli lint openapi.yaml`: exit 0, 4 warnings, no error
  - Frontend checks and integration tests skipped: no frontend, database or HTTP handler change.
- **Artifacts:** `backend/internal/apperr` and `backend/internal/handler/errors.go`; issues #40, #43 and #44 use them for 403 responses.
- **Links:** issue #35, branch `issue-35`.
