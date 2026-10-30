# Process

An issue goes from the backlog to a pull request in one cycle, started with `/issue <N>` and run by the `orchestrator` agent. The roles are the subagents in `.opencode/agents/`; the orchestrator starts them and does not do their work. The GitHub Issues are the backlog.

1. **Pick** the issue given to `/issue <N>`, or the next open one. Its dependencies must be closed (GitHub closes an issue when its pull request is merged into `main`); if one is open, stop and name it. Create the branch `issue-<N>` from `main`.
2. **Groom:** `pm` rewrites the issue into the task template. If `pm` reports that part of the issue belongs to another task, stop and tell the maintainer, who changes the backlog.
3. **Implement:** `software-engineer` works on `issue-<N>`.
4. **Verify:** `qa-engineer` checks the issue against its acceptance criteria.
5. **FAIL:** give the QA comment to the engineer and repeat steps 3 and 4; after three failed rounds, stop and ask the maintainer.
6. **PASS:** push the branch and open the pull request into `main`. Title: the issue title. Body: `.github/PULL_REQUEST_TEMPLATE.md` filled with the engineer's comment and the QA verdict, plus `Closes #<N>` so that GitHub closes the issue on merge; pass it to `--body` as a here-document, because the orchestrator cannot create files. Then stop and report the changed files, the verdict, the test results and the pull request link.
7. **Review:** the maintainer reviews the diff once and merges.

One issue per session: after the pull request the maintainer starts a new session.
