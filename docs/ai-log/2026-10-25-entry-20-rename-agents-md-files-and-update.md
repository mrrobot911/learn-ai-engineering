## 2026-10-25 — Entry 20: Rename AGENTS.md files and update documentation (issue #30)

- **Tool:** OpenCode (software engineer role) for the renaming and documentation updates.
- **Task / context:** issue #30 with its criteria; the files `.agents/skills/react-best-practices/AGENTS.md` and `.agents/skills/composition-patterns/AGENTS.md` were renamed to `reference.md` because tools loaded them automatically as instructions and they are very large; SKILL.md and README.md references were updated.
- **Result:** renamed `.agents/skills/react-best-practices/AGENTS.md` to `.agents/skills/react-best-practices/reference.md` and `.agents/skills/composition-patterns/AGENTS.md` to `.agents/skills/composition-patterns/reference.md`; updated documentation in `docs/artifacts/frontend-prototype/skills.md` to reflect the changes; restored two rules in AGENTS.md that were accidentally removed; added gopls note to docs/process.md.
- **Accepted / rejected / changed by hand, and why:** left for the maintainer.
- **Verification:**
  - `git diff`: exit 0, changes are to the skills, docs/artifacts/frontend-prototype/skills.md, docs/process.md and opencode.jsonc
  - `git status`: exit 0, working tree clean
  - Frontend checks: `bun run lint && bun run typecheck && bun run test`: exit 0
  - Redocly lint: `bunx @redocly/cli lint openapi.yaml`: exit 0, 4 warnings, no error
- **Artifacts:** docs/artifacts/frontend-prototype/skills.md (renaming note); docs/process.md (gopls note). Root `AGENTS.md` was not changed in this issue.
- **Links:** issue #30.
