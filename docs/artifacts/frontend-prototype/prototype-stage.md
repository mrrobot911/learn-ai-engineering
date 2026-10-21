# Frontend prototype stage: screens, routes and fixtures (issue #9, T09)

Stage artifact of the frontend-prototype stage, written when the last task of the stage (T09, issue
#9) closed the stage. It lists the screens of the prototype with their routes, how to run it and the
fixture difference the seeded attempts demonstrate. Context for this artifact: issue #9,
`AGENTS.md` > Conventions (each stage ends with an artifact), [`product-spec.md`](../../../product-spec.md)
section 4 US-3/US-6 and section 5, [`docs/artifacts/backlog/tasks.md`](../backlog/tasks.md) (T09).

The prototype is a React app with a mock API client: every screen reads through
`frontend/src/api/` and nothing calls a server. The mock stands in for the backend, so the
behaviour that the contract (`openapi.yaml`, task T12) will describe is already exercised here.

## The screens and their routes

| Route | Screen | File | What it does |
|---|---|---|---|
| `#/topics` | Topics | `frontend/src/pages/TopicsPage.tsx` | The topic list, read with `api.listTopics()`. |
| `#/topic/<slug>` | Topic | `frontend/src/pages/TopicPage.tsx` | The lessons of one topic, the completed count and the quiz entry. |
| `#/lesson/<id>` | Lesson | `frontend/src/pages/LessonPage.tsx` | The lesson body as Markdown and the completion control. |
| `#/quiz/<slug>` | Quiz start | `frontend/src/pages/QuizPages.tsx` | The entry to the quiz of a topic. |
| `#/quiz/<slug>/play` | Quiz | `frontend/src/pages/QuizPages.tsx` | The five questions of the topic, answered and submitted. |
| `#/quiz/<slug>/result/<attemptId>` | Quiz result | `frontend/src/pages/QuizPages.tsx` | The graded result of the submission. |
| `#/history` | Quiz history | `frontend/src/pages/HistoryPages.tsx` | The attempts of the logged-in account, newest first, read with `api.listAttempts()`. |
| `#/history/<attemptId>` | Attempt | `frontend/src/pages/HistoryPages.tsx` | One attempt with the stored question, selected answer and correct answer of every question, read with `api.getAttempt()`. |
| `#/login` | Log in | `frontend/src/pages/AuthPage.tsx` | The session of a registered account. |
| `#/register` | Register | `frontend/src/pages/AuthPage.tsx` | A new account. |

Every screen has a loading state, an error state with a "Try again" button and an empty state, and
every one of them reads its data through the client in `frontend/src/api/`.

## How to run the prototype

```bash
cd frontend
bun install
bun run dev
```

The dev server prints the local URL. Sign in with the seeded demo account of the mock:
`demo@forge.dev` with any password (the mock stores no password and only checks that the account
exists). A new account can be registered instead; it starts with an empty history.

The tests and the checks of this stage:

```bash
(cd frontend && bun run test)
(cd frontend && bun run lint)
(cd frontend && bun run typecheck)
```

## The fixture difference the seeded attempts demonstrate

US-6 stores every attempt with a snapshot of the question text, the selected answer text and the
correct answer text of the moment of the attempt, so that a later edit of a question does not change
a past attempt. The seeded attempt `att-seed-1` of `frontend/src/data/content.ts` demonstrates that
difference for the question `q-llm-1`:

| | Stored in `att-seed-1` | Current question `q-llm-1` |
|---|---|---|
| Question text | "What is the main training objective of a large language model?" | "What is the core training objective of a large language model during pre-training?" |
| Correct answer text | "Predict the next word in a sequence" | "Predict the next token in a sequence" |

The attempt detail page shows the stored wording, never the current one. The comparison is a test
of its own: `frontend/src/data/content.test.ts` fails when the stored wording and the current wording
are equal, and `frontend/tests/integration/quizHistory.test.tsx` reads the same attempt through the
mock client and checks the page against the stored texts.

The three seeded attempts of the demo account are:

| Attempt | Topic | Score | Date (UTC day) |
|---|---|---|---|
| `att-seed-3` | Retrieval-Augmented Generation | 3/5 | 2026-10-15 |
| `att-seed-2` | AI Agents | 5/5 | 2026-10-14 |
| `att-seed-1` | Large Language Models | 4/5 | 2026-10-12 |

The client returns them newest first, and the history page shows them in that order without a
sorting control.

## What the stage leaves to the next stage

- The contract of the two attempt calls does not exist yet: `openapi.yaml` is written in T12 (issue
  #12). The `ApiClient` type of `frontend/src/api/types.ts` is the only surface the screens use, and
  it was not changed in this stage.
- The backend, the database and the server-side ownership check come later (T19, T22). The mock
  already refuses the attempts of another account.
- The imported prototype store is gone: `frontend/src/store.tsx`, `frontend/src/useStore.ts` and
  `frontend/src/store/` are deleted, and no page reads the fixtures of `frontend/src/data/content.ts`.
