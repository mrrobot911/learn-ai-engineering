# Task template

Every GitHub issue uses this structure. The product manager role rewrites an issue into it before implementation (see [`process.md`](process.md)).

```markdown
## Goal
One or two sentences: what the user or the system can do after this task that it could not do before.
Source: <user story or section of product-spec.md>

## Acceptance criteria
- [ ] A testable statement: given <state or input>, when <action>, then <observable result>.
- [ ] Another testable statement.
- [ ] Tests for the new behaviour exist and pass (unit, and integration when it crosses layers).

## Out of scope
- What this task deliberately does not do, so it is not added by accident.

## Constraints
- Rules that apply, from `AGENTS.md` (for example: calls to the backend only through `frontend/src/api/`; contract first).
- Depends on: #<issue number>, or none.
```

## Rules for a good task
- Small enough to finish in one session, and understandable without reading the other issues.
- Every criterion can be checked by running something or looking at something specific.
- The first task of the project sets up an empty frontend and backend, each with one passing test, so the commands in `AGENTS.md` exist.
- Tasks follow the project order in `AGENTS.md`: prototype, contract, backend, integration, persistence, containers, CI/CD, deployment.
