# AI Engineering Learning Platform

A web application where developers learn AI engineering through structured lessons and short multiple-choice quizzes. Built as the final project of the [DataTalksClub AI Dev Tools Zoomcamp](https://github.com/DataTalksClub/ai-dev-tools-zoomcamp) using AI-assisted software development.

- Product requirements: [`product-spec.md`](product-spec.md)
- Instructions for AI tools: [`AGENTS.md`](AGENTS.md)
- API contract: [`openapi.yaml`](openapi.yaml)
- Live application: [https://learn-ai-engineering.onrender.com/](https://learn-ai-engineering.onrender.com/)

## Project status

> Work in progress. The specification and project instructions are written; implementation follows the order in `AGENTS.md`. The status table matches the repository as it is.

| Area | Status |
|---|---|
| Product specification | done (`product-spec.md`, draft v0.2) |
| Project instructions (`AGENTS.md`) | done |
| Task tracker (GitHub Issues with acceptance criteria) | in use ([issues](https://github.com/mrrobot911/learn-ai-engineering/issues)) |
| Frontend prototype with mock API client | done (T02-T09) |
| OpenAPI contract | done (`openapi.yaml`: content, quiz, authentication, progress, history and maintainer operations) |
| Backend and tests | done (Go API and unit tests in `backend/`) |
| Database and integration tests | done (PostgreSQL, migrations in `backend/migrations/`, integration tests run separately from unit tests) |
| Docker Compose | done (`docker-compose.yml`: `db`, `backend`, `frontend`; `make up` / `make down`) |
| CI/CD and deployment | done for CI; deployment configured for Render (single service, manual promotion from `dev` to `main`; see section 10 and `docs/deployment.md`) |
| Local observability stack | local only, not deployed (`observability/` Compose profile, Grafana dashboard, alert and receiver, on-call drill; see section 8) |
| Agent extension pack, security and audit | not done (course modules not added to this repository) |
| Agent roles, task template and process | done (`.opencode/agents/`, `docs/process.md`, `docs/task-template.md`) |

## 1. Problem

Developers who want to move into AI engineering have to piece together scattered articles, videos and documentation, and have no simple way to check what they actually understood.

This application organises learning material by topic and lets users test their knowledge with short quizzes. Everything works without registering. Registered users can additionally keep their lesson progress and quiz history.

## 2. What the system does

- **Topics and lessons.** Content is organised into topics, each with ordered lessons written in Markdown. Initial topics: LLMs, AI agents, MCP, RAG, embeddings, evaluation, deployment.
- **Quizzes.** Each quiz has 5 multiple-choice questions drawn at random from the topic's published question bank. Correct answers are revealed only after submission, with a short explanation.
- **Anonymous access.** Visitors can read all lessons and take quizzes without an account. Anonymous results are shown but not stored.
- **Accounts.** Users can register and log in (email and password, cookie session). Logged-in users can mark lessons as completed and see their past quiz attempts.
- **Content preparation.** Lessons and questions are drafted with AI assistance as _drafts_, reviewed by the maintainer and published explicitly. Users only ever see published content.
- **Seed content.** On first start every initial topic has at least 1 published lesson and 6 published questions.

Not in version 1: AI-generated curriculum in the UI, social features, an admin web interface, payments, password reset by email, mobile apps. Full list and acceptance criteria: [`product-spec.md`](product-spec.md).

## 3. Technologies and architecture

| Part | Technology | Role |
|---|---|---|
| Frontend | React, TypeScript, Bun | User interface for lessons, quizzes, accounts and history. All backend calls go through a single API client module. |
| Backend | Go | REST API that implements `openapi.yaml`: content, quiz grading, authentication, progress and history. |
| Database | PostgreSQL | Persistent storage for content, users, sessions, progress and quiz attempts. Schema changes through numbered migrations. |
| API contract | OpenAPI | Single source of truth for the frontend-backend interface. |
| Containerisation | Docker, Docker Compose | Runs frontend, backend and database together with one command. |
| CI/CD | GitHub Actions | Runs tests on every change; deploys only when tests pass. |

How the parts fit together:

```text
Browser ── React app ── API client ──HTTP/JSON──> Go API ──> PostgreSQL
                                          ▲
                              described by openapi.yaml
```

Backend layers: thin HTTP handlers, services with the logic, and a repository layer that holds all SQL.

## 4. AI-assisted development workflow

The project follows a specification-first workflow:

1. **Specification first.** [`product-spec.md`](product-spec.md) defines the problem, user stories, acceptance criteria and non-goals before any code.
2. **Project instructions.** [`AGENTS.md`](AGENTS.md) gives AI tools the commands, layout, conventions and gotchas of this project, on top of a global set of personal rules.
3. **Task tracking.** Implementation tasks are tracked as GitHub Issues in this repository. Each issue references the relevant requirements in the specification and defines testable acceptance criteria. AI tools receive the issue as the implementation task, and commits and verification results are linked back to it. Work follows [`docs/process.md`](docs/process.md): a product manager role grooms each issue, a software engineer role implements it and a QA engineer role verifies it against the acceptance criteria (subagents in `.opencode/agents/`), which the `orchestrator` agent runs as one cycle per issue, started with `/issue <N>`; the maintainer starts no role and reviews once at the end, after the QA engineer reports PASS. The cycle ends with a pull request into `dev`: the orchestrator pushes the branch and opens the pull request from [`.github/PULL_REQUEST_TEMPLATE.md`](.github/PULL_REQUEST_TEMPLATE.md) after the QA verdict, and the maintainer merges it; because `dev` is the default branch, GitHub then closes the issue. The maintainer promotes `dev` to `main` by hand.
4. **Implementation with AI tools.** Code is generated task by task, in the order: frontend prototype with mocks, OpenAPI contract, backend with its tests, persistent storage with integration tests, connect the frontend to the backend, deployment pipeline. Tools: OpenCode in the repository (configuration in `opencode.jsonc`), Bolt for the UI prototype _(planned)_, Claude for planning and documents. Project skills for UI and React work are kept in `.agents/skills/`: `frontend-design`, `web-design-guidelines`, `react-best-practices` and `composition-patterns`, copied from their upstream repositories at the commit recorded in [`docs/artifacts/frontend-prototype/skills.md`](docs/artifacts/frontend-prototype/skills.md).
5. **Stage artifacts.** Each stage ends with an artifact committed to the repository and linked to its issue (specification, prototype screenshots or notes, `openapi.yaml`, test results). The next stage receives it as context by path. Documents and screenshots are stored in `docs/artifacts/`; the specification and the API contract stay in the repository root.
6. **Verification.** Generated code is checked by tests (unit and separate integration tests), a lint step, and manual review of the diff.
7. **Logging.** For each significant task, one file in [`docs/ai-log/`](docs/ai-log/) (rules and template in [`docs/ai-workflow.md`](docs/ai-workflow.md)) records the issue and its acceptance criteria, the instructions and specification given to the AI tool, what it implemented, which tests were run and their result, and what was reviewed or changed manually. Secrets and private data are never included.

AI-prepared lesson and question drafts always go through human review before publication.

### AI workflow evidence

- Public task tracker: [GitHub Issues](https://github.com/mrrobot911/learn-ai-engineering/issues)
- AI workflow log: [`docs/ai-log/`](docs/ai-log/), rules in [`docs/ai-workflow.md`](docs/ai-workflow.md)
- Implementation history: [commits](https://github.com/mrrobot911/learn-ai-engineering/commits), linked to their issues

## 5. Repository structure

```text
README.md                           this file
product-spec.md                     product specification
AGENTS.md                           instructions for AI tools
opencode.jsonc                      OpenCode project configuration (tools, permissions)
openapi.yaml                        API contract
Dockerfile                          single deployment image: frontend build + Go backend
docker-compose.yml                  application stack (db, backend, frontend) and observability profile
Makefile                            test, check and stack targets
.env.example                        documented environment variables (see below)
render.yaml                         Render Blueprint (single web service and database)
.github/PULL_REQUEST_TEMPLATE.md    pull request template, filled after the QA verdict
.github/workflows/                  CI pipeline (ci.yml)
.agents/skills/                     project skills for AI tools
.opencode/agents/                   role subagents (product manager, software engineer, QA engineer)
.opencode/commands/                 OpenCode commands
backend/
  Dockerfile                        backend container image
  AGENTS.md, cmd/, internal/, migrations/, README.md, seed/, tests/
frontend/
  Dockerfile                        frontend container image (served via Caddy/proxy in Compose)
  AGENTS.md, src/, tests/, package.json, ...
observability/                      local observability stack (not deployed)
  README.md                         observability usage, alert and checks
  otel-collector-config.yaml, prometheus.yml, loki.yaml, tempo.yaml
  alertmanager/                     Alertmanager config template
  prometheus/                       alert rules and tests
  grafana/                          provisioned datasources and dashboard JSON
  receiver/                         alert receiver and on-call drill service (Dockerfile, Go code)
  smoke-check.sh, alert-*.sh, alert-lib.sh, alert-test.compose.yml
scripts/
  frontend-integration.sh           frontend integration test command used by `make test-integration`
  contract.sh, story.sh             helper scripts
docs/
  ai-workflow.md                    rules for AI workflow log entries
  ai-log/                           per-issue/task AI workflow log files
  process.md, task-template.md      process and task template
  deployment.md                     Render deployment steps
  artifacts/
    backlog/                        task tracker artifacts
    frontend-prototype/             frontend prototype notes
    integration/                    integration artifacts
    planning/                       planning artifacts
    incidents/                      on-call drill incident bundles and evidence (local, no secrets/personal data)
```

`.env.example` documents every environment variable used by the repository: `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `DATABASE_URL`, `TEST_DATABASE_URL`, `HOST`, `PORT`, `STATIC_DIR` (commented example), `ENV`, `SEED_MAINTAINER_EMAIL`, `SEED_MAINTAINER_PASSWORD`, `VITE_API_MODE`, `VITE_API_BASE_URL`, `API_UPSTREAM`, `OTEL_SERVICE_NAME`, `OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_EXPORTER_OTLP_PROTOCOL`, `ALERT_WEBHOOK_URL`, `OPENCODE_PROVIDER`, `OPENCODE_MODEL`, `OPENCODE_API_KEY`, `OPENAI_API_KEY` (commented), `ANTHROPIC_API_KEY` (commented), `LOKI_URL`, `TEMPO_URL`, `INCIDENT_BUNDLE_DIR`, `REPO_DIR`, `REPO_SRC`, `OPENCODE_BIN`, and `FAULT_INJECT_QUIZ_SUBMISSION_5XX` (commented, test-only).

## 6. API contract

The API contract is defined in [`openapi.yaml`](openapi.yaml) before any real API call is implemented. The frontend and the backend communicate only through this interface, and the backend follows it. It describes the content and quiz endpoints, authentication and the session, the lesson progress, the quiz history and the maintainer draft and publish operations.

## 7. Database

PostgreSQL stores topics, lessons, questions with answer options, users, sessions, lesson progress and quiz attempts (with a snapshot of the question and answers at the time of the attempt). High-level model: section 9 of [`product-spec.md`](product-spec.md).

Configuration comes from environment variables, with separate settings for development, test and production (see `.env.example`: `POSTGRES_*`, `DATABASE_URL`, `TEST_DATABASE_URL`, `ENV`). Schema changes are managed by numbered migrations in `backend/migrations/`. Integration tests use an isolated test database, never the development database.

## 8. Running the project

Requirements: Docker with Docker Compose. For local development and tests also Go and [Bun](https://bun.sh). Versions of Go, Bun, base images and PostgreSQL are pinned in `go.mod`, the Dockerfiles and `docker-compose.yml`; use those versions.

```bash
cp .env.example .env
make up
```

`make up` runs `docker compose up --build` (see [`Makefile`](Makefile)). The application, API and database start together. Ports and URLs are defined in [`docker-compose.yml`](docker-compose.yml).

- Frontend: `http://localhost:3000`
- Backend API (inside the Compose network): `http://backend:8080` (exposed to the frontend container; not published on the host)

Stop the stack with:

```bash
make down
```

`make down` runs `docker compose down`.

### Observability stack (local only, not deployed)

The optional `observability` Compose profile starts an OpenTelemetry Collector, Prometheus, Loki, Tempo, Grafana, Alertmanager and the alert receiver for local inspection and the on-call drill. It is not started by default, uses local volumes only, contains no secrets, and is not used by Render or CI/CD. Details: [`observability/README.md`](observability/README.md). On-call drill incident bundles and evidence are stored under [`docs/artifacts/incidents/`](docs/artifacts/incidents/) (local only; no secrets or personal data).

Start the application together with observability (the backend exports telemetry only when `OTEL_EXPORTER_OTLP_ENDPOINT` is set; without it the backend sends no telemetry). Command from [`observability/README.md`](observability/README.md):

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318 OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf docker compose --profile observability up --build
```

Stop the profile with:

```bash
docker compose --profile observability down
```

Never use `docker compose --profile observability down -v`: it removes volumes, including the application `db_data` volume.

#### Quiz journey dashboard

Grafana is provisioned from files (no manual dashboard or datasource setup):

- URL: `http://127.0.0.1:3001/d/quiz-journey/quiz-journey`
- UID: `quiz-journey`
- Dashboard JSON: [`observability/grafana/dashboards/quiz-journey.json`](observability/grafana/dashboards/quiz-journey.json)

The dashboard covers `GET /topics/{topicSlug}/quiz` and `POST /topics/{topicSlug}/quiz/submissions` with request rate, 5xx rate and p95 latency (Prometheus), quiz submissions and trace search (Tempo), and journey logs with `trace_id` (Loki).

#### Alert and receiver

Prometheus evaluates the quiz submission 5xx alert (`QuizSubmission5xxRateHigh`) defined in [`observability/prometheus/alerts.yml`](observability/prometheus/alerts.yml). Alertmanager delivers notifications to `ALERT_WEBHOOK_URL` (see `.env.example`; by default the profile delivers to the local receiver service at `http://receiver:8081/alerts`). The receiver stores incident data (logs and traces) and, when configured, starts the on-call agent for the local drill as described in [`observability/README.md`](observability/README.md), [`AGENTS.md`](AGENTS.md) (On-call agent) and incident evidence in [`docs/artifacts/incidents/`](docs/artifacts/incidents/).

Localhost-bound ports (observability profile), as defined in [`docker-compose.yml`](docker-compose.yml):

| Service | URL / port |
|---|---|
| Grafana UI | `http://127.0.0.1:3001` |
| Prometheus | `http://127.0.0.1:9090` |
| Loki | `http://127.0.0.1:3100` |
| Tempo | `http://127.0.0.1:3200` |
| Alertmanager | `http://127.0.0.1:9093` |
| Receiver (health) | `http://127.0.0.1:8081` |
| OTel Collector (OTLP gRPC) | `127.0.0.1:4317` |
| OTel Collector (OTLP HTTP) | `127.0.0.1:4318` |
| OTel Collector (health) | `http://127.0.0.1:13133` |

Observability-related variables (documented in [`.env.example`](.env.example)): `OTEL_SERVICE_NAME`, `OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_EXPORTER_OTLP_PROTOCOL`, `ALERT_WEBHOOK_URL`, `OPENCODE_PROVIDER`, `OPENCODE_MODEL`, `OPENCODE_API_KEY` (or provider-specific equivalents such as `OPENAI_API_KEY`, `ANTHROPIC_API_KEY`), `LOKI_URL`, `TEMPO_URL`, `INCIDENT_BUNDLE_DIR`, `REPO_DIR`, `REPO_SRC`, `OPENCODE_BIN`, and test-only `FAULT_INJECT_QUIZ_SUBMISSION_5XX` (commented; not for production/Render).

Smoke check (from [`observability/README.md`](observability/README.md)): starts the profile, verifies service and OTLP health, generates quiz journey traffic, and queries the Grafana API to confirm dashboard `quiz-journey` exists, is provisioned, and contains the required panels; then stops and removes only `otel-collector`, `prometheus`, `alertmanager`, `loki`, `tempo` and `grafana` with their volumes, not the receiver.

```bash
./observability/smoke-check.sh
```

## 9. Testing

Makefile targets (run from the repository root, see [`Makefile`](Makefile)):

- `make test` — all unit tests (backend `go test ./...`, receiver `CGO_ENABLED=0 go test ./...` in `observability/receiver`, frontend `bun install --frozen-lockfile && bun run test`).
- `make test-integration` — backend and frontend integration tests on an isolated test database (backend runs with `-p 1` because packages share one database; frontend integration runs via `bash scripts/frontend-integration.sh`).
- `make check` — checks that do not change files (`gofmt -l` and `go vet` for `backend` and `observability/receiver`, frontend `typecheck`, `lint` and `format:check`, and `bunx @redocly/cli lint openapi.yaml`).
- `make up` / `make down` — start and stop the Compose stack (`docker compose up --build` / `docker compose down`).

CI runs on every push and on every pull request (see [`.github/workflows/ci.yml`](.github/workflows/ci.yml)). The `check-and-test` job runs `make check` and `make test`; a separate `integration-tests` job starts `postgres:16-alpine` (same major as `docker-compose.yml`), creates an isolated `learn_ai_engineering_test` database, supplies `TEST_DATABASE_URL` and `DATABASE_URL` from the workflow only, and runs `make test-integration`. Integration tests always use an isolated test database (never the development database). Failing tests fail the job; no deploy step runs.

```bash
# unit tests
make test

# integration tests (requires Docker; uses an isolated test database, never dev)
make test-integration

# checks (no file changes)
make check

# formatting (changes files; commands from AGENTS.md / Makefile checks)
(cd backend && gofmt -w .)
(cd frontend && bun run format)
```

Unit tests live next to the code. Integration tests are kept separate from unit tests and cover key workflows such as browsing content, taking and submitting a quiz, registering and logging in, and saving progress and quiz history.

## 10. CI/CD and deployment

- **CI:** GitHub Actions ([`.github/workflows/ci.yml`](.github/workflows/ci.yml)) runs `make check` and `make test` on every push and on every pull request, and runs `make test-integration` in a separate job with an isolated PostgreSQL test database.
- **CD / promotion:** the maintainer promotes `dev` to `main` by hand (see [`docs/process.md`](docs/process.md)); Render deploys from `main` when checks pass (`autoDeployTrigger: checksPass` in [`render.yaml`](render.yaml)), as described in [`docs/deployment.md`](docs/deployment.md).
- **Hosting:** [Render](https://render.com) free plan, one Docker web service named `learn-ai-engineering` built from the root [`Dockerfile`](Dockerfile) through [`render.yaml`](render.yaml) (Blueprint). The image contains the frontend build and the Go backend, which serves the app and the API at the service public URL.
- **Public URL:** [https://learn-ai-engineering.onrender.com/](https://learn-ai-engineering.onrender.com/) (Render service URL).
- **Cold start:** a free web service sleeps after 15 minutes without traffic; the first request after that takes about a minute to wake it.

Deployment steps, Blueprint setup and rollback are in [`docs/deployment.md`](docs/deployment.md). Variables from [`render.yaml`](render.yaml): `DATABASE_URL` is provided by the linked Render Postgres database; `ENV`, `HOST` and `PORT` are set in the Blueprint; `SEED_MAINTAINER_EMAIL` and `SEED_MAINTAINER_PASSWORD` are marked `sync: false` and must be set in the Render dashboard (no values in the repository). `STATIC_DIR` is set to `/app/static` in the root `Dockerfile`. If an older two-service Blueprint is still deployed, delete the old `learn-ai-engineering-backend` and `learn-ai-engineering-frontend` services in the Render dashboard after the new `learn-ai-engineering` service is live.

## 11. Reproducibility

A reviewer should be able to run the project from a clean machine by following this README:

1. Clone the repository ([https://github.com/mrrobot911/learn-ai-engineering](https://github.com/mrrobot911/learn-ai-engineering)) and copy `.env.example` to `.env`.
2. Start the application stack with `make up` (`docker compose up --build`).
3. Open the frontend at `http://localhost:3000` (as defined in `docker-compose.yml`).
4. Run the tests and checks with the commands in section 9 (`make test`, `make test-integration`, `make check`).
5. Deploy with the steps in [`docs/deployment.md`](docs/deployment.md).

The status table in this README marks observability as local only and not deployed.

## 12. Course context

Final project of the DataTalksClub AI Dev Tools Zoomcamp. The requirements for the 2026 cohort are a draft and may change; this project follows the draft criteria: problem description, AI workflow, architecture, frontend, API contract, backend, database, containerisation, integration tests, deployment, CI/CD, agent extension pack, security and audit, and reproducibility.
