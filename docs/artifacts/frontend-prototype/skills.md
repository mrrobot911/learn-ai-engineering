# Project skills installed for UI and React work (issue #2, T02)

Stage artifact of the frontend-prototype stage, task T02. It records what was installed in
`.agents/skills/`, from where, at which commit, under which licence, so that the tasks that follow
(#3 to #9) load the skills by name and know what is vendored. Context for this artifact: issue #2,
`AGENTS.md` > Skills, [`docs/artifacts/backlog/tasks.md`](../backlog/tasks.md) (T02).

Installed per project only, never globally: `~/.config/opencode/skills`, `~/.claude/skills` and
`~/.agents/skills` do not hold any of them, and `.gitignore` does not exclude `.agents/`.

Date of the install: 2026-10-18.

## Sources

| Skill | Repository | Path in the repository | Commit fetched and copied | Licence file that came with it |
|---|---|---|---|---|
| `frontend-design` | `anthropics/skills` | `skills/frontend-design` | `dbd4588f9e1033efb41dad4bef2f7947c8993d44` | `LICENSE.txt` in the directory, Apache License 2.0 |
| `web-design-guidelines` | `vercel-labs/agent-skills` | `skills/web-design-guidelines` | `063bee94c3f4df8453406c830b0a7df0f2860278` | none: no licence file ships with the directory and the repository has none |
| `react-best-practices` | `vercel-labs/agent-skills` | `skills/react-best-practices` | `063bee94c3f4df8453406c830b0a7df0f2860278` | none: no licence file ships with the directory and the repository has none; the `SKILL.md` frontmatter declares `license: MIT` |
| `composition-patterns` | `vercel-labs/agent-skills` | `skills/composition-patterns` | `063bee94c3f4df8453406c830b0a7df0f2860278` | none: no licence file ships with the directory and the repository has none; the `SKILL.md` frontmatter declares `license: MIT` |

Both shas are the head of the `main` branch at install time and are recorded so the copy can be
compared with `diff -r` against a fresh checkout. Pinning to a later commit is a separate change
with its own log entry.

## What each directory holds

- `frontend-design/`: `SKILL.md`, `LICENSE.txt`.
- `web-design-guidelines/`: `SKILL.md` only.
- `react-best-practices/` and `composition-patterns/`: `SKILL.md`, `AGENTS.md`, `README.md`,
  `metadata.json` and `rules/`. The nested `AGENTS.md` is compiled output from upstream and is kept
  verbatim; it never overrides the project `AGENTS.md`.

93 files in total. Everything is copied file for file: nothing is renamed, summarised, translated or
edited, with the single exception below.

## The one permitted deviation

Criterion 2 of issue #2 and the OpenCode documentation both require the frontmatter `name` of a
`SKILL.md` to be equal to the directory that holds it. Upstream ships a different value for the two
React skills, so only that one line was changed in each file:

| File | Upstream `name:` | Installed `name:` |
|---|---|---|
| `react-best-practices/SKILL.md` | `vercel-react-best-practices` | `react-best-practices` |
| `composition-patterns/SKILL.md` | `vercel-composition-patterns` | `composition-patterns` |

Nothing else differs: `diff -r` against the checkouts prints nothing for `frontend-design` and
`web-design-guidelines`, and only those two lines for the two React skills. All four `SKILL.md`
files carry a non-empty `description`.

## How the next tasks use them

Load the skill before writing UI or React code, as `AGENTS.md` > Skills says: `frontend-design` and
`web-design-guidelines` for the screens, `react-best-practices` and `composition-patterns` for React
code. Both React skills target React and Next.js in general, so only the rules that fit this plain
React app apply and the Next.js-specific ones are ignored. A skill never overrides `AGENTS.md` or
`product-spec.md`.

## Verification done at install time

- `diff -r` of each installed directory against a checkout of the recorded commit: identical except
  the two `name:` lines above.
- Frontmatter of the four `SKILL.md` files read: `name` equal to the directory name, `description`
  present.
- The global skill locations listed: none of the four skills is there; `git check-ignore .agents`
  reports nothing.
- `rg` over `.agents/skills` for `sk-`, `ghp_`, `-----BEGIN`, `api_key` and `password`: no match. No
  `.env` file was added.
- The `skill` tool was called once with each of the four names. Every call answered "not found"
  because the skill list of that session was built at session start, before these files existed, so
  loading could not be confirmed there; a session started after this commit sees them.
- No test was run and none was added: this change adds no behaviour (criterion 11 of issue #2).