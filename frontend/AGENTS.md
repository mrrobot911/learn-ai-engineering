# Frontend rules (React + TypeScript)

## Commands
Run from the repository root as `(cd frontend && ...)`.
- JavaScript tooling is Bun only (`bun install` / `bun add`), never npm, npx, yarn or pnpm.
- test: `bun run test [pattern]` (not `bun test`, which starts Bun's own runner); integration tests are in `frontend/tests/integration/`
- checks: `bun run lint && bun run typecheck && bun run format:check`; format: `bun run format`

## Rules
- Components never call `fetch`; every backend call goes through the client in `frontend/src/api/`.
- Build only the screens in `product-spec.md`, with loading, error and empty states, keyboard access and labels. Add no features or dependencies for visual reasons.
- Skills (`.agents/skills/`): for UI work load `frontend-design` and `web-design-guidelines`; for React code use `react-best-practices` and `composition-patterns`, apply only the rules that fit a plain React app, ignore the Next.js ones, and load single `rules/*.md` files. A skill never overrides `AGENTS.md` or `product-spec.md`.
- Install or update skills only when asked, and read the `SKILL.md` before committing.
