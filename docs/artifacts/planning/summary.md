# Planning summary

Stage: planning and project documents, before any code. Date: 2026-10-17.
Tools: Claude (chat) for planning and drafting; a second AI assistant for a review of the written files.
The raw record of the conversation is in `planning-conversation.md`. This file only summarises it.

## What was done
1. Read the draft 2026 final-project requirements (14 criteria, 30 points) and compared my first plan with them: 5 criteria covered, 5 partial, 4 missing (CI/CD, agent extension pack, security and audit, separate integration tests).
2. Wrote `product-spec.md` (v0.1, then v0.2 after review).
3. Wrote `AGENTS.md` from my template, taking my global `AGENTS.md` into account, then reviewed it for conflicts and gaps.
4. Wrote `README.md` to the rubric, `docs/ai-workflow.md`, `opencode.jsonc` and `.gitignore`.
5. Checked the repository structure against the course materials.
6. Added the roles (product manager, software engineer, QA engineer), the task template and the process file, following the course Module 1 workflow.

## Decisions I made
Product:
- User accounts, saved progress and quiz history stay in version 1 (the AI suggested deferring them).
- A quiz always has 5 questions; sessions use a cookie with a 7-day expiry; the maintainer is a user with an `is_maintainer` flag.
- An unanswered question counts as incorrect; a question from another topic rejects the submission.
- Seed content is loaded on first start; AI-generated curriculum is out of version 1.

Process and tools:
- JavaScript tooling is Bun only; OpenCode denies `npm`, `npx`, `yarn` and `pnpm`.
- Project order: specification, frontend prototype with mocks, OpenAPI contract, backend, integration, persistence.
- Skills are installed in the project (`.agents/skills/`) at the prototype stage; versions are pinned in the files that use them.
- Each stage ends with an artifact that the next stage receives as context by path.
- The product manager role only grooms one issue; breaking the specification into tasks stays a main-session task that I review.
- Two manual checkpoints (after grooming, before closing) instead of the fully automatic loop.

Taken and not taken:
- Took from the third-party skill: reuse existing code, no long solutions, shortest working diff. Did not take the `shortcut:` comment rule or the session-wide persona mode.
- Did not take the second assistant's advice to drop CI/CD, the extension pack and the security directories (the draft rubric scores them). Its file name for the AI report could not be confirmed, so `docs/ai-workflow.md` stays.

## Problems found in review and fixed
- `cd` commands that break when run one after another; now run from the repository root in subshells.
- `order` used as a SQL column name; renamed to `position`.
- Cookie sessions without CSRF protection; added.
- README statements written as done while still planned; reworded.
- Wrong `npm` commands; replaced with Bun.
- Quiz grading and unanswered questions were undefined; defined.

## Still open
- The course criteria may be newer than the file I used (the course FAQ mentions 19 criteria and 41 points); re-check against the current `project/README.md`.
- The order "frontend prototype before contract" is taken from a review and not confirmed in the lesson.
- The roles and their permissions have not been run in OpenCode.
- Hosting is not chosen.
- The specification is not yet broken down into issues.
