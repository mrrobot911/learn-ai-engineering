# AI Engineering Learning Platform

Web app where developers learn AI engineering through lessons and multiple-choice quizzes; anonymous access, optional accounts for progress and quiz history. Full requirements: `product-spec.md`.

## Commands
- run everything: `docker compose up --build`
- Run every command from the repository root; wrap directory changes in a subshell, for example `(cd backend && go test ./...)`, because the shell may keep its working directory between commands.
- backend test: `(cd backend && go test ./...)` (single test: `(cd backend && go test ./... -run TestName)`)
- backend integration test: `(cd backend && go test -tags=integration ./...)` (needs Postgres from `docker compose up -d db`)
- frontend test: `(cd frontend && bun run test)` (single test: `(cd frontend && bun run test <file-or-pattern>)`); use `bun run test`, not `bun test`, which starts Bun's own runner instead of the project's
- checks that do not change files: `(cd backend && gofmt -l . && go vet ./...)`; `(cd frontend && bun run lint && bun run format:check)`
- apply formatting (changes files): `(cd backend && gofmt -w .)`; `(cd frontend && bun run format)`
- validate API contract: `bunx @redocly/cli lint openapi.yaml`
- JavaScript tooling is Bun only: do not use `npm`, `npx`, `yarn` or `pnpm`; install with `bun install` / `bun add`
- These commands are planned; scripts and tools may not exist until the project is scaffolded. If one is missing, say so and add it deliberately instead of guessing.

## Layout
- `README.md`, `.env.example` — setup, run, test and deploy instructions; documented environment variables
- `product-spec.md` — product requirements; source of truth for behaviour
- `openapi.yaml` — frontend/backend contract; source of truth for the API
- `frontend/` — React + TypeScript app; all backend calls go through `frontend/src/api/`
- `backend/` — Go API server, migrations, tests
- `docker-compose.yml` — frontend, backend and PostgreSQL for local run
- `.github/workflows/` — CI (tests) and CD (deploy after tests pass)
- `docs/` — AI workflow log, `process.md` (issue lifecycle and roles), `task-template.md`, stage artifacts (`docs/artifacts/`), deployment notes, `agent-extension-pack.md`, `permissions.md`
- `security/`, `ops/` — scan results, audit output, diagnosis notes, AI tool/data policy
- `agent-capabilities/`, `agent-hooks/`, `mcp-server/` — agent extension pack
- `opencode.jsonc` — OpenCode project configuration (tools and permissions)
- `.agents/skills/` — project skills for AI tools (see Skills)
- `.opencode/agents/` — role subagents: `pm`, `software-engineer`, `qa-engineer`
- `security/`, `ops/`, the agent extension pack directories and `.agents/skills/` are created only when the matching course task starts, not in advance.

## Conventions
- Project order: spec, then a frontend prototype with a mock implementation behind the central API client, then `openapi.yaml`, then backend and its tests, then connect the frontend to the backend, then persistent storage and integration checks. The contract must exist before any real API call is written.
- After the contract exists, change it first for any new behaviour: spec, `openapi.yaml`, then code. Do not add endpoints, fields or behaviour that are not in `openapi.yaml`; update the contract in the same change instead.
- Frontend: components never call `fetch` directly; use the client in `frontend/src/api/`.
- Backend: handlers stay thin; logic lives in services; SQL only in the repository layer. Return errors in the shape defined in `openapi.yaml`.
- Unit tests live next to the code; integration tests are separate (build tag `integration` in Go, `frontend/tests/integration/` for the frontend) and must stay separate.
- Every change that adds behaviour comes with a test (unit; integration too when it crosses frontend, backend or database). Run the relevant tests before saying it is done.
- Database changes go through numbered migration files, never by editing the schema by hand.
- Configuration and secrets come from environment variables, with separate settings for development, test and production; document new ones in `.env.example`. Never commit real secrets.
- Keep `README.md` runnable: if a command or step changes, update the README in the same change.
- Work from a GitHub issue that has testable acceptance criteria; reference the issue in commits and in the log entry.
- Follow `docs/process.md` for the issue lifecycle and the roles. When acting as a role from `.opencode/agents/`, do only what that role's file allows. Acceptance criteria are changed only by the product manager role or the maintainer.
- Each stage ends with an artifact committed to the repository and linked to its issue (for example the specification, prototype screenshots or notes, `openapi.yaml`, test results). The next stage receives the artifact as context by path, not by pasting the whole file. Note the artifact and where it was used in `docs/ai-workflow.md`. Documents, screenshots and reports go in `docs/artifacts/<stage>/` (create it with the first file); the specification and the API contract stay where Layout puts them.
- Versions: pin runtime and tool versions in the files that use them: Go in `go.mod`, Bun and base images in the Dockerfiles, the PostgreSQL major version as an image tag in `docker-compose.yml`, action versions in the workflows. Never use `latest`. Commit the lockfiles (`go.sum`, the Bun lockfile). Do not upgrade a version or add a dependency at another version unless the task asks for it, and read library docs for the pinned version.
- Log notable AI-assisted steps (issue and its acceptance criteria, instructions and context files given, what was generated, tests run and their result, manual review and changes made) in `docs/ai-workflow.md`.

## Skills
- Project skills live in `.agents/skills/` and are committed to the repository. Install or update them only when asked, read their `SKILL.md` before committing, and record the source and commit in `docs/ai-workflow.md`.
- UI work in `frontend/`: load `frontend-design` and `web-design-guidelines`. Build only the screens in `product-spec.md`, with loading, error and empty states, keyboard access and labels; add no features or dependencies for visual reasons.
- React code: use `react-best-practices` and `composition-patterns`. They target React and Next.js in general; apply only the rules that fit this plain React app and ignore Next.js-specific ones.
- Go, PostgreSQL and REST work: no extra skills. The rules are in this file and `openapi.yaml`; use the `gopls` tool for Go navigation and `context7` for library docs at the pinned versions.
- A skill never overrides this file or `product-spec.md`; if they conflict, these two win.

## Smallest complete change
Before writing:
- Find the code the task touches by searching first (`rg`, `ast-grep`), then read only what you need. List everything the change must reach: callers, tests, fixtures, config, exports. Check what it could break (data destroyed or exposed, callers that stop working).
- Check whether the feature needs to exist; `product-spec.md` decides. Skip options and flexibility nobody asked for and name them in one line. A vague request gets the smallest version that does the core job.

Writing, take the first option that fully works:
1. Something already in this project (helper, component, service, pattern): use it the way the surrounding code does.
2. Standard library or platform feature: use it, unless the project has its own.
3. A dependency that is already installed: use it. Never add a dependency for a few lines.
4. Otherwise write the minimum code that works.

Rules:
- Do not write logic that already exists elsewhere. No abstraction, wrapper, option, config or "for later" code that the task or `product-spec.md` does not ask for, beyond the layers and the API client required in Conventions. Deletion beats addition.
- The shortest working diff wins, once you know everything it must touch. A one-liner that needs decoding is not short.
- Be lazy about the solution, never about completeness: update the callers, tests and fixtures your change breaks. If that makes the change large, split it into reviewable steps.
- Keep the existing structure and layers. Code you move or merge keeps its error handling and validation.
- Between options of equal size, take the one that is correct on edge cases.
- Bug fix: find every caller of the function you touch (`rg` or `ast-grep`), then fix the root cause once in the shared code.
- Comments only for the why that the code cannot show, in one line.
- Never cut: validation at trust boundaries, error handling that prevents data loss, security, accessibility, anything the user asked for.
- End each reply that changes code with one or two lines: what was skipped or not verified, and any risk.

## Gotchas
- Quiz endpoints must not return the correct option before the attempt is submitted.
- Content has `draft` and `published` status; public endpoints return only `published`. Agents and tools may create drafts but must never publish.
- Users may only read and change their own progress and attempts; check ownership in every protected endpoint.
- Quiz grading happens on the server. An unanswered question, or an option that does not belong to its question, counts as incorrect; reject only a submission that contains a question not belonging to the topic.
- Anonymous quiz results are returned but not stored. Quiz submission works with or without a session and stores the attempt only when a valid session exists.
- Seed loading must be repeatable and must not duplicate existing content.
- Passwords are stored only as salted hashes; never return or log `password_hash` or passwords.
- The session cookie is `HttpOnly` and `SameSite`; state-changing requests need cross-site request forgery protection.
- Never put secrets or real user data into prompts, logs, fixtures or AI tool inputs (follow the AI tool/data policy in `security/` once it exists).
- Integration tests need a running test database; do not point them at the development database.
- Ask before destructive migrations (drop, rename, data rewrite).
- Do not run deploy commands or workflows (including `gh workflow run`) unless explicitly asked.
