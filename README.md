# AI Engineering Learning Platform

A web application where developers learn AI engineering through structured lessons and short multiple-choice quizzes. Built as the final project of the [DataTalksClub AI Dev Tools Zoomcamp](https://github.com/DataTalksClub/ai-dev-tools-zoomcamp) using AI-assisted software development.

- Product requirements: [`product-spec.md`](product-spec.md)
- Instructions for AI tools: [`AGENTS.md`](AGENTS.md)
- API contract: [`openapi.yaml`](openapi.yaml)
- Live application: _TODO: add the deployed URL after the first deployment_

## Project status

> Work in progress. The specification and project instructions are written; implementation follows the order in `AGENTS.md`. Sections below describe the intended system. Items marked _planned_ do not exist yet and are updated as the project advances.

| Area | Status |
|---|---|
| Product specification | done (draft v0.2) |
| Project instructions (`AGENTS.md`) | done |
| Task tracker (GitHub Issues with acceptance criteria) | planned |
| Frontend prototype with mock API client | planned |
| OpenAPI contract | planned |
| Backend and tests | planned |
| Database and integration tests | planned |
| Docker Compose | planned |
| CI/CD and deployment | planned |
| Agent extension pack, security and audit | planned (course modules) |
| Agent roles, task template and process | done (not yet tried on an issue) |

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
3. **Task tracking.** Implementation tasks are tracked as GitHub Issues in this repository. Each issue references the relevant requirements in the specification and defines testable acceptance criteria. AI tools receive the issue as the implementation task, and commits and verification results are linked back to it. Work follows [`docs/process.md`](docs/process.md): a product manager role grooms each issue, a software engineer role implements it and a QA engineer role verifies it against the acceptance criteria (subagents in `.opencode/agents/`), and the maintainer reviews after grooming and before closing. _Planned: issues are created before implementation starts._
4. **Implementation with AI tools.** Code is generated task by task, in the order: frontend prototype with mocks, OpenAPI contract, backend with its tests, persistent storage with integration tests, connect the frontend to the backend, deployment pipeline. Tools: OpenCode in the repository (configuration in `opencode.jsonc`), Bolt for the UI prototype _(planned)_, Claude for planning and documents. Project skills for UI and React work are kept in `.agents/skills/` _(planned: installed at the prototype stage)_.
5. **Stage artifacts.** Each stage ends with an artifact committed to the repository and linked to its issue (specification, prototype screenshots or notes, `openapi.yaml`, test results). The next stage receives it as context by path. Documents and screenshots are stored in `docs/artifacts/`; the specification and the API contract stay in the repository root.
6. **Verification.** Generated code is checked by tests (unit and separate integration tests), a lint step, and manual review of the diff.
7. **Logging.** For each significant task, [`docs/ai-workflow.md`](docs/ai-workflow.md) records the issue and its acceptance criteria, the instructions and specification given to the AI tool, what it implemented, which tests were run and their result, and what was reviewed or changed manually. Secrets and private data are never included. _Planned: created when implementation starts._

AI-prepared lesson and question drafts always go through human review before publication.

### AI workflow evidence

- Public task tracker: GitHub Issues of this repository _(TODO: add the link once the repository URL is final)_
- AI workflow log: [`docs/ai-workflow.md`](docs/ai-workflow.md)
- Implementation history: commits of this repository, each linked to its issue _(TODO: add the link)_

## 5. Repository structure

```text
README.md              this file
product-spec.md        product specification
AGENTS.md              instructions for AI tools
opencode.jsonc         OpenCode project configuration (tools, permissions)
.agents/skills/        project skills for AI tools (added at the prototype stage)
.opencode/agents/      role subagents: product manager, software engineer, QA engineer
.gitignore             files kept out of the repository
openapi.yaml           API contract
frontend/              React + TypeScript app
backend/               Go API server, migrations, tests
docker-compose.yml     frontend, backend and PostgreSQL
.env.example           documented environment variables
.github/workflows/     CI and CD pipelines
docs/                  AI workflow log, process and roles, task template, stage artifacts (docs/artifacts/), deployment, extension pack and permissions notes
```

Directories for the agent extension pack (`agent-capabilities/`, `agent-hooks/`, `mcp-server/`), `security/` and `ops/` are added when the matching course tasks are done.

## 6. API contract

The API contract will be defined in `openapi.yaml` before any real API call is implemented. The frontend and the backend will communicate only through this interface, and the backend will follow it. _Planned._

## 7. Database

PostgreSQL stores topics, lessons, questions with answer options, users, sessions, lesson progress and quiz attempts (with a snapshot of the question and answers at the time of the attempt). High-level model: section 9 of [`product-spec.md`](product-spec.md).

Configuration comes from environment variables, with separate settings for development, test and production. See `.env.example`. _Planned._

## 8. Running the project

Requirements: Docker with Docker Compose. For local development and tests also Go and [Bun](https://bun.sh). Versions of Go, Bun, base images and PostgreSQL are pinned in `go.mod`, the Dockerfiles and `docker-compose.yml`; use those versions.

```bash
cp .env.example .env
docker compose up --build
```

The application, API and database start together. Ports and URLs are defined in `docker-compose.yml`. _Planned: instructions are completed when the compose file exists._

## 9. Testing

```bash
# run every command from the repository root

# backend unit tests
(cd backend && go test ./...)

# start the database for integration tests
docker compose up -d db

# backend integration tests
(cd backend && go test -tags=integration ./...)

# frontend tests
(cd frontend && bun install && bun run test)

# checks that do not change files
(cd backend && gofmt -l . && go vet ./...)
(cd frontend && bun run lint && bun run format:check)

# apply formatting (changes files)
(cd backend && gofmt -w .)
(cd frontend && bun run format)

# validate the API contract
bunx @redocly/cli lint openapi.yaml
```

Unit tests live next to the code. Integration tests are kept separate from unit tests and cover the key workflows: browsing content, taking and submitting a quiz, registering and logging in, saving progress and quiz history. _Planned: commands are checked once the project is scaffolded._

## 10. CI/CD and deployment

- **CI:** a GitHub Actions workflow will run the backend and frontend tests on every push and pull request.
- **CD:** when tests pass on the main branch, the workflow will deploy the application.
- **Hosting:** _TODO: choose the hosting target (the course deploys to AWS)._
- **Proof of deployment:** _TODO: public URL and a short description of how it was verified._

Deployment steps will be documented in `docs/deployment.md`. _Planned._

## 11. Reproducibility

A reviewer should be able to run the project from a clean machine by following this README:

1. Clone the repository and copy `.env.example` to `.env`.
2. Run `docker compose up --build`.
3. Open the application URL from `docker-compose.yml`.
4. Run the tests with the commands in section 9.
5. Deploy with the steps in `docs/deployment.md`.

## 12. Course context

Final project of the DataTalksClub AI Dev Tools Zoomcamp. The requirements for the 2026 cohort are a draft and may change; this project follows the draft criteria: problem description, AI workflow, architecture, frontend, API contract, backend, database, containerisation, integration tests, deployment, CI/CD, agent extension pack, security and audit, and reproducibility.
