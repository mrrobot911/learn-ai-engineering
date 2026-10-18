# AI workflow log

This file records how AI tools were used to build this project: what was asked, what context was given, what was produced, what was changed by hand and how it was verified. One entry per significant task. Secrets and private data are never recorded here.

Related: [`AGENTS.md`](../AGENTS.md) (instructions for AI tools), [`product-spec.md`](../product-spec.md) (specification), GitHub Issues (tasks and acceptance criteria).

## Entry template

```text
## <date> — <task name> (issue #N)
- Tool:
- Prompt / task given:
- Context given (files, issue, spec):
- Result:
- Artifacts produced (path) and where the next stage used them:
- Accepted / rejected / changed by hand, and why:
- Verification (tests run and result, manual checks):
- Links (commit, issue, tool session):
```

## Entries

## 2026-10-17 — Entry 0: project planning, specification and instructions (no issue yet)

- **Tool:** Claude (chat), plus a second AI assistant used only to review the written files.
- **Prompt / task given:**
  - Find out what the final project requires and check my draft plan against the course rubric.
  - Turn the checked plan into `product-spec.md`, a project `AGENTS.md` (using my template and taking my global `AGENTS.md` into account), and `README.md` written to the rubric.
  - Review the files for completeness, conflicts with my global rules, and consistency with the course requirements.
- **Context given:**
  - The draft 2026 final-project requirements and criteria from the course repository.
  - My first project plan.
  - My global `AGENTS.md` and the project `AGENTS.md` template.
  - The text of a third-party "lazy senior developer" skill, to decide which of its rules to adopt.
  - A review of the files written by the second AI assistant.
- **Result:**
  - A gap analysis of the plan against the 14 rubric criteria (spreadsheet, kept outside the repository).
  - `product-spec.md` (draft v0.1, then v0.2).
  - `AGENTS.md` with commands, layout, conventions, a "smallest complete change" section and gotchas.
  - `README.md` structured by the rubric, with status markers for what is not built yet.
- **Decisions made by me, sometimes against what the AI proposed:**
  - Kept user accounts, saved progress and quiz history in version 1 (the AI had suggested deferring them to reduce scope).
  - Took the "reuse existing code, do not write long solutions, shortest working diff" rules from the third-party skill, and rejected the `shortcut:` comment rule and the session-wide persona mode.
  - Corrected the AI's use of `npm`: the project uses Bun only.
  - Decided that an unanswered quiz question counts as incorrect, instead of rejecting the submission.
  - Agreed that quiz size is fixed at 5 and a maintainer is a user with an `is_maintainer` flag; AI-generated curriculum stays out of version 1.
  - Took from the second assistant's review: the project order with a frontend prototype before the contract, commands run from the repository root in subshells, separate check and write formatting commands, and task tracking through GitHub Issues. Did not take: its request to remove CI/CD, the extension pack and security directories (the draft rubric scores them), and its file-name claim for the AI report, which I could not confirm.
  - Added project skills in `.agents/skills/`: `frontend-design`, `web-design-guidelines`, `react-best-practices` and `composition-patterns`, all installed in the project and not globally; no extra skills for Go and PostgreSQL. A skill never overrides `AGENTS.md` or the specification.
  - Added a rule that each stage ends with a committed artifact that the next stage receives as context by path.
  - Added a version policy: pin versions in `go.mod`, Dockerfiles, compose image tags and workflows, commit lockfiles, never `latest`.
  - Configured OpenCode for the project (`opencode.jsonc`): Go language server enabled, `npm`, `npx`, `yarn` and `pnpm` denied (Bun only), `gh workflow run` requires approval.
  - Checked the repository structure against the course materials: the draft project requirements put `README.md`, `product-spec.md`, `AGENTS.md` and `openapi.yaml` in the root and the rest in `docs/`, `security/` and `ops/`; stage artifacts that are documents or screenshots go to `docs/artifacts/`.
- **Verification:**
  - Files were checked against the 14 draft rubric criteria and for conflicts with my global `AGENTS.md`.
  - Found and fixed by review: a `cd` sequence that would break when commands run one after another, a reserved SQL word used as a column name (`order`), a missing protection for cookie sessions, and statements in the README written as done while still planned.
  - No code or tests exist yet, so nothing was run.
- **Artifacts:** `product-spec.md`, `AGENTS.md`, `README.md`; used as context for the roles and for the backlog breakdown.
- **Links:** commits e039aa5, 0c0537d, c062760. Export of the planning conversation to `docs/artifacts/planning/`.

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

## 2026-10-18 — Entry 3: the orchestrator runs the whole cycle, one review at the end (no issue yet)

- **Tool:** OpenCode (main session).
- **Prompt / task given:** Change the lifecycle of an issue so the main session, as orchestrator, runs the whole cycle for one issue by itself (pick, groom, implement, verify) without the maintainer starting any role, repeats implement and verify after a QA FAIL for at most three rounds, and stops on PASS. Keep the wording of the six files in step with that change: `docs/process.md` (lifecycle, branches, backlog changes, end of cycle), `README.md` (section 4 item 3), `docs/ai-workflow.md` (the decision line in Entry 1), `AGENTS.md` (the Conventions line), `.opencode/agents/pm.md` (step 4) and `.opencode/agents/software-engineer.md` (step 6). Do not commit.
- **Context given:** `docs/process.md`, `README.md`, `AGENTS.md`, `.opencode/agents/pm.md`, `.opencode/agents/software-engineer.md`, `docs/ai-workflow.md`.
- **Result:** the lifecycle in `docs/process.md` lost the `@role` commands and the check after grooming, and gained the automatic cycle, a `Branches` section and a `Changes to the backlog` section; README section 4 item 3 describes the same cycle; the Conventions line in `AGENTS.md` says that a request to work on an issue means the full cycle with the main session as orchestrator; `pm.md` step 4 and `software-engineer.md` step 6 were extended as asked. Nothing else in those files was changed, and nothing was committed.
- **Decisions given by the maintainer:** the two manual checkpoints of Entry 1 are dropped, because the orchestrator now runs the cycle itself and the maintainer reviews once at the end (the diff and the log entry) before the orchestrator closes the issue. The orchestrator creates `issue-<N>` from `main` before the cycle, never pushes, merges or runs a workflow, stops and names the issue when a dependency is not yet closed and merged, and works on one issue at a time. When the product manager reports that part of an issue belongs to another task, the orchestrator itself moves the criteria into `docs/artifacts/backlog/tasks.md`.
- **Verification:** no code or tests exist, so nothing was run. The diff of the six files was read back and checked change by change against the request. The other uncommitted changes in the working tree (the lines of Entry 2 above and `docs/artifacts/backlog/tasks.md`) belong to the previous task and are not part of this one.
- **Artifacts:** `docs/process.md`, `README.md`, `AGENTS.md`, `.opencode/agents/pm.md`, `.opencode/agents/software-engineer.md`, `docs/ai-workflow.md`; `docs/process.md` is what the next cycle follows.
- **Links:** no commit yet; nothing has been accepted or changed by hand so far.

<!-- Add new entries below, newest last. -->
