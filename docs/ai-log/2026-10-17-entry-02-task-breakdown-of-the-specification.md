## 2026-10-17 — Entry 2: task breakdown of the specification (issues #1-#13 for T01-T13)

- **Tool:** OpenCode (main session).
- **Prompt / task given:** Break `product-spec.md` into an ordered task list following `docs/task-template.md`, then apply the maintainer's corrections: fix the header links, the T17, T20 and T33 dependencies, merge T22 and T23, narrow the CSRF rule to state-changing requests that carry a session cookie, and record the open questions as decisions.
- **Context given:** `product-spec.md` (draft v0.2), `AGENTS.md`, `README.md`, `docs/task-template.md`, `docs/process.md`, `docs/ai-workflow.md`.
- **Result:** `docs/artifacts/backlog/tasks.md` with T01-T37 in the project order from `AGENTS.md`; the project order corrected in `AGENTS.md` and `README.md` (section 4). The issues for the first three stages (T01-T13) were then created from that file, one label per stage: `foundation`, `frontend-prototype`, `api-contract`.
- **Decisions made by me, confirmed by the maintainer:**
  - T22 and T23 are merged into one persistence task, so the list ends at T37; every other task stays as it is.
  - CI (T33) depends on T01 alone and runs lint, format, unit tests and the contract lint from the start; its integration job against PostgreSQL is added after T21.
  - The CSRF token is required on state-changing requests that carry a session cookie; an anonymous quiz submission works without it.
  - Passwords: Argon2id, at least 8 and at most 128 characters. Session cookie `SameSite=Lax`.
  - Seed content is drafted with AI, reviewed by hand and committed as seed files; one lesson and six questions per topic is the minimum.
  - The frontend is served by its own static server image; the hosting target is chosen by the maintainer before the deployment task, and the image task is not blocked by it.
- **Verification:** no code exists, so nothing was run. The dependency graph, the Order table and the header links were checked with a script over the file; one dependency check was found to be vacuous on the first run and was redone. The created issues were verified through `gh issue list`/`gh issue view` (title, label, dependency line); they were not opened in a browser.
- **Artifacts:** `docs/artifacts/backlog/tasks.md`; the source the maintainer turns into GitHub issues.
- **Links:** issues #1-#13 for T01-T13 (stages Foundation, Frontend prototype, API contract); the files above are in commit `a59ae3f`.
