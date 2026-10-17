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
1. **Pick** the next open issue whose dependencies are closed.
2. **Groom:** `@pm groom issue #N`.
3. **Maintainer check:** read the groomed issue and approve it or send it back.
4. **Implement:** `@software-engineer implement issue #N`.
5. **Verify:** `@qa-engineer test issue #N`.
6. **FAIL:** give the QA comment to the software engineer and repeat steps 4 and 5. After three failed rounds, stop and ask the maintainer.
7. **PASS:** the maintainer reviews the diff and fills in the manual review in the log entry; the orchestrator closes the issue.

The maintainer checks twice: after grooming and before closing. This keeps the manual review that [`ai-workflow.md`](ai-workflow.md) records.

## Notes
- Each role works on one issue per run and leaves its result as an issue comment.
- Role permissions in the agent files are guard rails, not a guarantee. Review what each run did.
- More roles cost more time and tokens. For a small change a direct request to the main session is enough; record it in the log in the same way.
