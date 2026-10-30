## 2026-10-30 — Start issues with a command and an orchestrator agent (chore)
- Tool: Claude (cloud session); the files are copied by the maintainer.
- Task given: "do issue N" often fails (the agent writes code itself, skips QA, does not open the pull request, or does not read docs/process.md); make the cycle start reliably, keep the cycle in docs/, and let GitHub close the issue on merge.
- Context given (files by name): docs/process.md, AGENTS.md, README.md, .github/PULL_REQUEST_TEMPLATE.md, the three role files, opencode.jsonc; the OpenCode v2 documentation for commands, skills, agents and the v1 migration.
- Result: new command `.opencode/commands/issue.md` (`/issue <N>`) only starts the agent `orchestrator` on the cycle in docs/process.md; new primary agent `.opencode/agents/orchestrator.md` cannot edit files, can start only pm, software-engineer and qa-engineer, and cannot merge, close issues or run workflows. docs/process.md is cut to the cycle itself; the backlog-move rule is dropped (the maintainer changes the backlog), and the pull request carries `Closes #N`, so GitHub closes the issue on merge (PR template changed from `Refs #` to `Closes #`). The old "do issue #N" line is removed from AGENTS.md; README names the orchestrator instead of the main session.
- Artifacts: .opencode/commands/issue.md, .opencode/agents/orchestrator.md.
- Accepted / rejected / changed by hand, and why: left for the maintainer.
- Verification: frontmatter of both new files parsed as YAML; no Cyrillic in the changed files. Not run: `/issue N` in OpenCode, so the permission rules (subagent, shell, edit) are unverified on the maintainer's version.
- Links: branch chore/issue-command.
