## 2026-10-30 — Slim AGENTS.md and process.md, one log file per entry (chore)
- Tool: Claude (cloud session); the patch is applied by the maintainer.
- Task given: cut the context the roles read; give each rule one owner; split the single AI log into one file per entry.
- Context given (files by name): AGENTS.md, docs/process.md, docs/ai-workflow.md, docs/ai-workflow-archive.md, the three role files, opencode.jsonc, openapi.yaml (structure only).
- Result: root AGENTS.md 11.1 KB to 2.9 KB, stack commands and gotchas moved to backend/ and frontend/AGENTS.md; process.md 4.5 KB to 2.6 KB; docs/ai-workflow.md is only rules and a template; entries 0-23 split unchanged into docs/ai-log/; scripts/story.sh and scripts/contract.sh read one user story or part of the contract; nixos MCP off and gopls written in full in opencode.jsonc.
- Artifacts: docs/ai-log/, scripts/.
- Accepted / rejected / changed by hand, and why: left for the maintainer.
- Verification: both scripts run by hand on openapi.yaml and product-spec.md; the patch applies to 3ecba9e with git apply --check. Not run: redocly lint, a full "do issue N" cycle.
- Links: branch chore/slim-agents-process.
