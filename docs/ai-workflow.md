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
- **Links:** _TODO: add the commit that adds these files; save an export of the planning conversation in `docs/artifacts/planning/` and link it here._

<!-- Add new entries below, newest last. -->
