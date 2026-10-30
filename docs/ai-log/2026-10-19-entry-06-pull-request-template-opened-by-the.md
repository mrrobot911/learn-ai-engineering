## 2026-10-19 — Entry 6: pull request template, opened by the orchestrator (no issue yet)

- **Tool:** OpenCode (main session) with the `brainstorming` skill.
- **Prompt / task given:** "Add a pull request template to the project", then two decisions from the maintainer: take the pull request that already exists as the model, and let the agent fill the template in while never pushing to `main`. Scope: the template, the process, the roles and this log.
- **Context given:** pull request #14 (merged, issue #2) as the model; [`docs/process.md`](process.md); `.opencode/agents/pm.md`, `software-engineer.md` and `qa-engineer.md`; `README.md` sections 4 and 5; `opencode.jsonc`; the push rules of `AGENTS.md`.
- **Result:**
  - `.github/PULL_REQUEST_TEMPLATE.md` holds the six sections of pull request #14 in the same order: Summary, Related issue, Verification, Independent QA, Limitations and follow-ups, Author checklist. Each has an HTML comment saying what belongs in it, and the `Independent QA` section carries the verdict the QA engineer reported.
  - [`docs/process.md`](process.md): step 6 of the lifecycle now pushes the branch and opens the pull request before the report, and the `Branches` section states who pushes (`git push origin issue-<N>` plus `gh pr create` with the filled template), that the engineer commits but does not push, and that the orchestrator never merges into `main`.
  - `README.md`: section 4 item 3 says the cycle ends with the pull request the orchestrator opens and the maintainer merges; section 5 lists the template next to `.github/workflows/`.
  - `opencode.jsonc` gained `"git push origin main*": "ask"`, so the orchestrator can push `issue-<N>` but has to ask before anything reaches `main`. The maintainer chose `ask` over `deny`: the rule stops an automatic push to `main` without taking the choice away.
  - No role file changed: `software-engineer.md` already forbids the engineer to push, and `qa-engineer.md` already denies it in its permissions.
- **Decisions made by me:**
  - The orchestrator pushes the branch and opens the pull request, not the software engineer and not the QA engineer.
  - The pull request opens after the QA verdict, so `Independent QA` is never filled with "pending"; one `gh pr create` replaces a draft plus a later edit.
  - The title is the issue title and the body references the issue with `Refs #<N>`, as pull request #14 did.
- **Verification:** no code exists and none was added, so no test was run. The template was read back after writing, the whole diff was reviewed file by file, and a search for the old wording found one place left, the historical record of Entry 3, which now carries a note about the change.
- **Artifacts:** `.github/PULL_REQUEST_TEMPLATE.md`, the body the orchestrator fills at step 6 of every issue cycle in [`docs/process.md`](process.md).
- **Links:** pull request #14 as the model; branch `pr-template`, not pushed without the maintainer's word.
