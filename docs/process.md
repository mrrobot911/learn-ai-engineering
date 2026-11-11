# Process

An issue goes from the backlog to a pull request in one cycle, started with `/issue <N>` and run by the `orchestrator` agent. The roles are the subagents in `.opencode/agents/`; the orchestrator starts them and does not do their work. The GitHub Issues are the backlog.

`dev` is the repository default branch and `main` stays the production branch. The maintainer makes `dev` the GitHub default branch by hand (Settings -> Branches); the agent does not change GitHub settings. The maintainer protects `main` by hand with a ruleset: pull requests required, the CI checks required, force push and deletion blocked, only the maintainer can bypass, and no auto-merge. The agent does not change GitHub settings.

GitHub closes an issue only when the pull request merges into the default branch, so a merge into `dev` closes the issue only because `dev` is the default branch. A pull request into `dev` with `Closes #N` closes the issue on merge. The maintainer promotes by merging `dev` into `main` by hand.

1. **Pick** the issue given to `/issue <N>`, or the next open one (`gh issue list`). Read it with `gh issue view <N>`: the issues in its `Depends on` line must be closed (GitHub closes an issue when its pull request is merged into the default branch), checked with `gh issue view <D> --json number,state,title`, never with the full text; if one is open, stop and name it. Create the branch from `dev` with two separate commands: `git fetch origin`, then `git switch -c issue-<N> origin/dev` (if the branch already exists, `git switch issue-<N>`).
2. **Groom:** if the body read in step 1 has the headings Goal, Acceptance criteria and Out of scope, skip this step. Otherwise note `gh issue view <N> --json updatedAt --jq .updatedAt`, start `pm` to rewrite the issue into the task template, and read `updatedAt` again: if it did not change, `pm` wrote nothing, so stop and tell the maintainer. If `pm` reports that part of the issue belongs to another task, stop and tell the maintainer, who changes the backlog.
3. **Implement:** `software-engineer` works on `issue-<N>`.
4. **Verify:** `qa-engineer` checks the issue against its acceptance criteria.
5. **FAIL:** give the QA comment to the engineer and repeat steps 3 and 4; after three failed rounds, stop and ask the maintainer.
6. **PASS:** read the engineer's and QA comments with `gh issue view <N> --json comments`, then push the branch and open the pull request into `dev`. Run each command on its own: do not chain commands or use `$(...)`, because every part must be allowed separately.

   ```bash
   git push origin issue-<N>
   gh pr create --base dev --head issue-<N> --title "<issue title>" --body-file - <<'EOF'
   <filled template>
   EOF
   ```

   Title: the issue title. Body: `.github/PULL_REQUEST_TEMPLATE.md` filled with the engineer's comment and the QA verdict, plus `Closes #<N>` so that GitHub closes the issue on merge. The body goes in on stdin because the orchestrator cannot create files. Then stop and report the changed files, the verdict, the test results and the pull request link.
7. **Review:** the maintainer reviews the diff once and merges the pull request into `dev`. The maintainer promotes `dev` to `main` by hand.

One issue per session: after the pull request the maintainer starts a new session.
