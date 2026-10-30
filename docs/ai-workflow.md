# AI workflow log

Rules and template for the log of AI-assisted work. The entries are in [`ai-log/`](ai-log/), one file each; this file never grows. Secrets and private data are never recorded.

## Rules
- One file per issue (or per task without an issue): `docs/ai-log/<YYYY-MM-DD>-issue-<N>.md`, or `docs/ai-log/<YYYY-MM-DD>-<short-slug>.md` when there is no issue. When QA returns FAIL and the engineer fixes the issue, the same file is updated.
- The author of the work writes the entry (the software engineer, or the main session for a direct change). The line about what was accepted or changed by hand is left for the maintainer.
- `ls docs/ai-log/` is the index; entries are history for the maintainer. When a project skill is installed or updated, its entry records the source and the commit.
- Keep an entry within about 1.5 KB: the task, the context files by name, the result in five to eight lines, the verification commands with their exit codes, the maintainer decision and the links. No list of changed files and no repeated diff.
- Entries 0-23 were moved here unchanged from the earlier single log and keep their `entry-NN` file names.

## Template

```text
## <date> — <task name> (issue #N)
- Tool:
- Task given:
- Context given (files by name):
- Result (5-8 lines):
- Artifacts (path) and where the next stage used them:
- Accepted / rejected / changed by hand, and why:
- Verification (commands and exit codes, manual checks):
- Links (commit, issue, pull request, tool session):
```
