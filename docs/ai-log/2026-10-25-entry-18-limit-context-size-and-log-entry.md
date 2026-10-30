## 2026-10-25 — Entry 18: limit context size and log entry length (chore-context-limits)

- **Tool:** OpenCode (main session) implementing chore-context-limits branch.
- **Prompt / task given:** limit context size and log entry length: (1) docs/process.md needs the new "Context size" section added before "Notes"; (2) AGENTS.md: add to AI workflow log convention that entries stay short (task, context files by name, result in 5-8 lines, verification commands with exit codes, maintainer decision, links; no file lists or diffs) and add (cd frontend && bun run typecheck) to run with bun run test and bun run lint before frontend changes are reported as done; (3) .opencode/agents/software-engineer.md: add that log entries keep the length set in AGENTS.md and hold no file lists or diffs.
- **Context given:** docs/process.md (partial), AGENTS.md, .opencode/agents/software-engineer.md, docs/ai-workflow.md (last entry).
- **Result:** Added "Context size" section to docs/process.md before "Notes"; updated AGENTS.md to specify short log entries and added typecheck command; updated software-engineer.md to specify log entry length; added short entry to docs/ai-workflow.md.
- **Accepted / rejected / changed by hand, and why:** left for the maintainer.
- **Verification:**
  - `git diff`: exit 0, changes are to docs/process.md, AGENTS.md and .opencode/agents/software-engineer.md
  - `grep -n "Context size" docs/process.md`: exit 0, section found at line 49
  - `git diff AGENTS.md`: exit 0, added log entry length specification and typecheck command
  - `git diff .opencode/agents/software-engineer.md`: exit 0, added log entry length specification
- **Artifacts:** docs/process.md updated with "Context size" section; AGENTS.md updated with context size and log entry length guidelines; .opencode/agents/software-engineer.md updated with log entry length specification.
- **Links:** branch chore-context-limits.
