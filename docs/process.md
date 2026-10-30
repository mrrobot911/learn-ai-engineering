# Process

How work moves from the specification to closed issues. Code rules and file-reading rules are in [`AGENTS.md`](../AGENTS.md).

## Backlog
Made in the main session, without a role: it breaks `product-spec.md` into small tasks in the project order with [`task-template.md`](task-template.md); the maintainer merges, splits and removes anything outside version 1; the tasks become GitHub Issues, which are the backlog from then on. When a role reports that part of an issue belongs to another task, the orchestrator moves those criteria into that task in `docs/artifacts/backlog/tasks.md` before it continues.

## Roles
Subagents in `.opencode/agents/`: `pm`, `software-engineer`, `qa-engineer`; each file says what the role does and does not do. The main session is the orchestrator: it starts a role with the issue number and file paths, not pasted text, and does not do the role's work.

## Lifecycle of an issue
The orchestrator runs the whole cycle for one issue; the maintainer starts no role.
1. **Pick** the next open issue whose dependencies are closed and merged into `main` by the maintainer; if one is missing, stop and name it. One issue at a time. Create the branch `issue-<N>` from `main`.
2. **Groom:** `pm` rewrites the issue into the task template.
3. **Implement:** `software-engineer` works on `issue-<N>`.
4. **Verify:** `qa-engineer` checks the issue against its acceptance criteria.
5. **FAIL:** give the QA comment to the engineer and repeat steps 3 and 4; after three failed rounds, stop and ask the maintainer.
6. **PASS:** push the branch and open the pull request (title: the issue title; body: [`.github/PULL_REQUEST_TEMPLATE.md`](../.github/PULL_REQUEST_TEMPLATE.md) filled with the engineer's comment and the QA verdict; `Refs #<N>`):

   ```bash
   git push origin issue-<N>
   gh pr create --base main --head issue-<N> --title "<issue title>" --body-file <filled template>
   ```

   Then stop and report the changed files, the verdict, the test results and the pull request link.
7. **Review:** the maintainer reviews the diff once, merges, and only then does the orchestrator close the issue. The orchestrator never merges into `main`.

## Notes
- Each stage ends with an artifact in the repository (`docs/artifacts/<stage>/`, or where the spec and contract live). One issue per session: after the pull request the maintainer starts a new session, and the next stage gets its context from the artifacts by path.
- Role permissions in the agent files are guard rails, not a guarantee; review what each run did.
