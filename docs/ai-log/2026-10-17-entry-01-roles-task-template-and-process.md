## 2026-10-17 — Entry 1: roles, task template and process (no issue yet)

- **Tool:** Claude (chat).
- **Prompt / task given:** Set up the roles from the course workflow (product manager, software engineer, QA engineer), a task template and a process file, for OpenCode.
- **Context given:** The course Module 1 article on specifications, loop and graph engineering; the OpenCode agents documentation; the project `AGENTS.md` and `opencode.jsonc`.
- **Result:** `.opencode/agents/pm.md`, `software-engineer.md`, `qa-engineer.md`; `docs/task-template.md`; `docs/process.md`; matching lines in `AGENTS.md` and `README.md`.
- **Decisions made by me:**
  - The product manager only grooms one issue at a time, as in the course. Breaking the specification into tasks stays a main-session task reviewed by me, not a role (the AI had offered a two-mode product manager; I declined it).
  - Roles are defined in `.opencode/agents/` because that is where OpenCode finds subagents, instead of a `docs/team/` folder.
  - Changed on 2026-10-18: the orchestrator runs the full cycle for one issue automatically and the maintainer reviews once at the end. The manual checkpoints had been proposed by Claude as a safety measure for untested roles.
  - Role permissions repeat the Bun-only denials, because a broad `"*": allow` on a role could otherwise override the project-level rules.
- **Verification:** Not run yet. The roles have not been used, and the permission rules and the `@role` calls are untested. TODO: run the product manager on the first issue and record the result and any fixes here.
- **Artifacts:** `.opencode/agents/*.md`, `docs/task-template.md`, `docs/process.md`; the task template is used by the backlog breakdown, the roles by the first grooming run.
- **Links:** commit c8f6884.
