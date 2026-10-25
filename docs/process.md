# Process

How work moves from the specification to closed issues. Rules for the code itself are in [`AGENTS.md`](../AGENTS.md).

## Backlog
The backlog is created in the main session, without a role:
1. The main session breaks `product-spec.md` into small tasks in the project order, using [`task-template.md`](task-template.md).
2. The maintainer reviews the list: merges, splits, and removes anything outside version 1.
3. The tasks become GitHub Issues. From then on the issues are the backlog.

## Roles
The roles are subagents defined in `.opencode/agents/`. The main session is the orchestrator: it starts the roles and does not do their work itself.

| Role | File | Does | Does not |
|---|---|---|---|
| Product manager | `pm.md` | Grooms one issue into the task template | Write code, change the specification |
| Software engineer | `software-engineer.md` | Implements one groomed issue with tests | Change criteria, close the issue |
| QA engineer | `qa-engineer.md` | Verifies against the criteria, reports PASS or FAIL | Fix code, close the issue |

## Lifecycle of an issue
The orchestrator runs the whole cycle for one issue by itself. The maintainer starts no role.
1. **Pick** the next open issue whose dependencies are closed and merged into `main`. One issue at a time.
2. **Groom:** the product manager rewrites issue #N into the task template.
3. **Implement:** the software engineer implements issue #N on its branch.
4. **Verify:** the QA engineer verifies issue #N against its acceptance criteria.
5. **FAIL:** give the QA comment to the software engineer and repeat steps 3 and 4. After three failed rounds, stop and ask the maintainer.
6. **PASS:** push the branch, open the pull request from [`.github/PULL_REQUEST_TEMPLATE.md`](../.github/PULL_REQUEST_TEMPLATE.md) filled with the engineer's comment and the QA verdict, then stop and report the changed files, the QA verdict, the test results, the log entry the software engineer added to [`ai-workflow.md`](ai-workflow.md), and the link to the pull request.
7. **Maintainer review:** the maintainer reviews the diff and the log entry. Only after that review does the orchestrator close the issue.

The maintainer reviews once, at the end. This keeps the manual review that [`ai-workflow.md`](ai-workflow.md) records.

## Branches
- The orchestrator creates the branch `issue-<N>` from `main` before it starts the cycle.
- Every issue the chosen issue depends on must be closed and its branch merged into `main` by the maintainer. Otherwise the orchestrator stops and names the issue that is missing.
- The software engineer commits to that branch, with `#<N>` in the commit message, and does not push.
- The orchestrator pushes `issue-<N>` and opens the pull request after the QA engineer reported PASS, so the body carries the final verdict:

  ```bash
  git push origin issue-<N>
  gh pr create --base main --head issue-<N> --title "<issue title>" --body-file <filled template>
  ```

- The pull request title is the issue title; the body references the issue with `Refs #<N>`.
- The orchestrator never merges into `main` and never runs a workflow. Merging is the maintainer's.

## Changes to the backlog
When the product manager reports that part of the issue belongs to another task, the orchestrator updates `docs/artifacts/backlog/tasks.md` before it continues: the receiving task gets the criteria.

## Context size
Do not read `docs/ai-workflow.md`, `docs/artifacts/backlog/tasks.md` or `openapi.yaml` whole, in the main session or in a role. Find the entry for the issue with `rg -n "issue #<N>"` in `ai-workflow.md` and read only it (or the last one when adding a new entry); read only the section of the task in `tasks.md`; read only the paths and schemas of `openapi.yaml` that the issue touches. One issue per session: after the pull request is opened the maintainer starts a new session, and the next stage gets its context from the artifacts by path, not from the conversation.

## Notes
- Each role works on one issue per run and leaves its result as an issue comment.
- Role permissions in the agent files are guard rails, not a guarantee. Review what each run did.
- More roles cost more time and tokens. For a small change a direct request to the main session is enough; record it in the log in the same way.
