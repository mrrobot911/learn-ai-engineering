# Planning conversation (condensed)

Condensed record of the planning conversation with Claude, 2026-10-17. See `summary.md` for the outcome.

## 1. Requirements
- **Me:** What does the final project require? (link to the course `project` folder)
- **Claude:** The page could not be opened automatically. From the course description: build, deploy and document a full-stack app made with AI tools.
- **Me:** Shared the draft 2026 project README (14 criteria, 30 points, expected repository contents).
- **Claude:** Summarised the criteria, the repository structure and a work plan. Noted that the requirements are a draft.

## 2. Check of my plan
- **Me:** Shared my first plan; asked for a check against the course materials.
- **Claude:** Too much: chat leftovers in the text, accounts, AI-generated curriculum, a stack that contradicted "no stack in the plan". Missing: CI/CD, agent extension pack, security and audit, separate integration tests, peer review.
- **Me:** Asked for a spreadsheet of that answer. **Claude:** Built one with rubric coverage, scope review and a repository checklist.
- **Me:** Keep user accounts.

## 3. Specification
- **Me:** I wanted to write the specification.
- **Claude:** Wrote `product-spec.md` v0.1 (problem, users, scope, user stories, data model).

## 4. AGENTS.md
- **Me:** Write the project `AGENTS.md` from my template and take my global `AGENTS.md` into account.
- **Claude:** Wrote it without repeating global rules.
- **Me:** Should I use rules from a "lazy senior developer" skill?
- **Claude:** Some yes, some no. **Me:** Drop the `shortcut:` comment rule; keep "reuse, no long solutions, shortest working diff". Review the rules again.
- **Claude:** Rewrote the section as "Smallest complete change".
- **Me:** Check for gaps and conflicts with my global rules.
- **Claude:** Fixed search tools, directory changes, deploy and migration rules, and added missing security gotchas.
- **Me:** There is no `npm`, only Bun. **Claude:** Replaced all JavaScript commands.

## 5. Review by a second AI assistant
- **Me:** Pasted its review of the two files.
- **Claude:** Accepted: project order with a prototype first, open questions in the specification, no empty directories in advance. Not accepted: removing CI/CD, the extension pack and security directories (the rubric scores them); the file name for the AI report could not be confirmed.
- **Me:** Apply the changes. **Claude:** Specification v0.2 (quiz size 5, cookie session, maintainer flag, seed content, answer snapshots).
- **Me:** An unanswered question must count as incorrect. **Claude:** Changed.

## 6. README and log
- **Me:** Write the README to the rubric and course materials. **Claude:** Wrote it with status markers for what is not built yet.
- **Me:** Pasted review notes on the README. **Claude:** Accepted relative links, task tracking through GitHub Issues, commands from the repository root; declined one point it could not confirm.
- **Me:** What should I document and how? **Claude:** Prompts, context files, what was accepted or changed, verification; created `docs/ai-workflow.md` with entry 0.

## 7. Repository
- **Me:** Chose a new repository. **Claude:** Suggested a name, a description and topics.
- **Me:** How do I commit, and when do I install skills? **Claude:** Three small commits; skills at the start of the prototype stage.

## 8. Skills, artifacts, versions, structure
- **Me:** Add design skills for the agent, an artifact for each stage, and check the README; shared my OpenCode config.
- **Claude:** Project skills for UI and React, a rule that each stage leaves an artifact used as context by the next stage, a version policy (pin versions, commit lockfiles, never `latest`), and an OpenCode config with Bun-only permissions.
- **Me:** Install React skills in the project too. **Claude:** Done.
- **Me:** Check where the course puts documents. **Claude:** Structure matches the draft requirements; stage artifacts go to `docs/artifacts/`. Found that the course FAQ mentions 19 criteria and 41 points, which differs from my file.

## 9. Roles and backlog
- **Me:** Should a product manager role split the specification into tasks?
- **Claude:** In the course article the product manager only grooms one issue; the task list is made in the main session.
- **Me:** Write all roles; no two-mode product manager; do as the course does.
- **Claude:** Wrote the product manager, software engineer and QA engineer roles, a task template and a process file, with two manual checkpoints.
- **Me:** Is it right to add them now, what do I pass for the breakdown, how do the tasks reach Issues?
- **Claude:** Yes; pass the specification, `AGENTS.md`, the task template and the process file with a prompt; create issues with `gh` after a dry run.
- **Me:** Committed the files to `main`. **Claude:** Checked the repository: all files match; remaining TODOs are the log links, planning artifacts and README links.
