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
- Do not run deploy commands or workflows (`gh workflow run`) unless asked.
- Log every notable AI-assisted step as its own file in `docs/ai-log/` (rules and template: `docs/ai-workflow.md`). Never read existing entries.

## Writing code
- Before writing, search (`rg`, `ast-grep`) and list everything the change must reach: callers, tests, fixtures, config, exports. Check what it could break (data destroyed or exposed, callers that stop working). Reuse in this order: existing project code, the standard library, an already-installed dependency, new code — use it the way the surrounding code does. Never add a dependency for a few lines; deletion beats addition.
- If the feature's need is unclear, `product-spec.md` decides; a vague request gets the smallest version that does the core job.
- Smallest complete change: the shortest working diff, readable, not clever. No abstraction, wrapper, option or "for later" code that nobody asked for; name skipped options in one line. Update the callers, tests and fixtures the change breaks; split a large change into reviewable steps.
- Moved or merged code keeps its structure, layers, error handling and validation. Between options of equal size, take the one that is correct on edge cases.
- Bug fix: find all callers, then fix the root cause once in shared code.
- Comments only for the why that code cannot show, one line at most.
- Never cut validation at trust boundaries, error handling that prevents data loss, security, accessibility or anything the user asked for. End a reply that changes code with one or two lines on what was skipped or not verified, and any risk.

## Context
- Never read whole: `openapi.yaml`, `docs/artifacts/backlog/tasks.md` and `done.md`, `docs/artifacts/planning/`, `.agents/skills/*/reference.md`. Read a user story with `scripts/story.sh <n>`, the contract with `scripts/contract.sh` (usage in its first lines), a task by its heading in `tasks.md`.
- Do not read the same range again without a change in between; save long command output to a file once and read the file.
- This file is already in every role's context; a role does not open it again.
