# AI Engineering Learning Platform

Web app: lessons and multiple-choice quizzes about AI engineering; anonymous access, optional accounts for progress and quiz history. `product-spec.md` (behaviour) and `openapi.yaml` (API) are the sources of truth.

Frontend work: also follow `frontend/AGENTS.md`. Backend work: also follow `backend/AGENTS.md`.

## Commands
Run from the repository root; wrap directory changes in a subshell, as `(cd <dir> && <command>)`. The backend and frontend commands are in their own `AGENTS.md`.
- everything: `docker compose up --build`
- contract: `bunx @redocly/cli lint openapi.yaml`
- Install packages per project, never globally. A missing script (for example `format` in `frontend/package.json`) is reported and added deliberately.

## Rules
- Order: spec, frontend prototype with a mock behind the API client, `openapi.yaml`, backend, storage, frontend connected to the backend, deployment. Contract first: new behaviour changes the spec, then `openapi.yaml`, then code; nothing that is not in `openapi.yaml`.
- Every change that adds behaviour has a test (integration too when it crosses frontend, backend or database); run the relevant tests before saying it is done. Integration tests stay separate from unit tests.
- Configuration and secrets come from environment variables; document new ones in `.env.example`. Never commit secrets or put them, or real user data, into prompts, logs or fixtures.
- Pin versions (`go.mod`, Dockerfiles, the compose image tag, workflow actions), never `latest`; commit lockfiles; add or upgrade a dependency only when the task asks.
- Keep `README.md` runnable when a command or step changes.
- Smallest complete change: reuse what exists, write no code nobody asked for, but update the callers, tests and fixtures the change breaks. Never cut validation, error handling, security or accessibility. End a reply that changes code with one or two lines on what was skipped or not verified, and any risk.
- Do not run deploy commands or workflows (`gh workflow run`) unless asked.
- Work from an issue. "do issue #N" means the cycle in `docs/process.md`, with the main session as orchestrator; implement directly only when the maintainer explicitly asks for a direct change.
- Log every notable AI-assisted step as its own file in `docs/ai-log/` (rules and template: `docs/ai-workflow.md`). Never read existing entries.

## Context
- Never read whole: `openapi.yaml`, `docs/artifacts/backlog/tasks.md` and `done.md`, `docs/artifacts/planning/`, `.agents/skills/*/reference.md`. Read a user story with `scripts/story.sh <n>`, the contract with `scripts/contract.sh` (usage in its first lines), a task by its heading in `tasks.md`.
- Do not read the same range again without a change in between; save long command output to a file once and read the file.
- This file is already in every role's context; a role does not open it again.
