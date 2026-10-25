# Backlog: finished tasks T02-T12

These are the tasks of [`tasks.md`](tasks.md) that are done and merged into `main` (T02-T12), moved here on 2026-10-25 unchanged, so that the open backlog stays short. Nothing in the text was edited; the open tasks and the order table are in `tasks.md`.

---

## T02 - Install the project skills for UI and React work

**User story:** none - foundation for the frontend tasks
**Depends on:** none

## Goal

The four skills named in `AGENTS.md` are installed under `.agents/skills/`, committed and recorded, so every UI and React task can load them before writing code.
Source: `AGENTS.md` "Skills" (course module: project skills).

## Acceptance criteria

- [ ] `.agents/skills/` contains `frontend-design`, `web-design-guidelines`, `react-best-practices` and `composition-patterns`, each with a readable `SKILL.md`.
- [ ] Each `SKILL.md` is read before its directory is committed.
- [ ] The source (repository plus version or commit) of each skill is recorded in `docs/ai-workflow.md`.
- [ ] `AGENTS.md` matches what is installed, and states that a skill never overrides `AGENTS.md` or `product-spec.md`.
- [ ] Each skill loads successfully in OpenCode (loading it does not fail).
- [ ] Tests are not applicable; verification is the load check above.

## Out of scope

- Writing or editing skills of your own.
- Global installation in `~/.config/opencode`.
- The agent extension pack (`agent-capabilities/`, `agent-hooks/`, `mcp-server/`), `security/` and `ops/`.

## Constraints

- Install only when asked, which this task is.
- Never commit credentials or private data from the source repository.
- Depends on: none.

---

## T03 - Central API client module with a mock implementation

**User story:** none - required by `product-spec.md` section 7 (one client module); supports US-1 to US-6
**Depends on:** T02

Groomed as issue #3. The issue body is the source of truth for its criteria; this entry records the
criteria that moved between tasks.

## Goal

The frontend has exactly one module that talks to the backend, with a mock implementation behind it, so screens can be built and reviewed before any real API call exists.
The frontend project is the Bolt prototype imported in `frontend/` (commit `9437a03` on `issue-3`); every later frontend task works from that output.
Source: `product-spec.md` section 7 (API contract), `AGENTS.md` "Conventions".

## Acceptance criteria

- [ ] `frontend/src/api/` exports a client type with one method per operation of the user stories US-1 to US-6: list topics, get topic, get lesson, start quiz, submit quiz, register, log in, log out, get progress, mark lesson complete, undo lesson completion, list attempts, get attempt.
- [ ] The value the start-quiz operation resolves with carries no field that reveals the correct option or the explanation, and a unit test asserts it; grading stays on the server. Moved here from T06 so the client type is right before the screens are built.
- [ ] A mock implementation of that type serves local fixture data; choosing between mock and real is a single configuration value, and no component can tell which is active.
- [ ] No component, page or hook contains `fetch` or `XMLHttpRequest` (verified by search).
- [ ] Unit tests for the client module pass: the mock returns the fixtures, and the type has no unimplemented method.
- [ ] The `frontend-design`, `web-design-guidelines`, `react-best-practices` and `composition-patterns` skills were loaded while writing it.

## Out of scope

- Any real HTTP call, retry, timeout or status-to-error mapping (T27-T30).
- Any screen.

## Constraints

- No dependency is added for something the platform or the installed tools already provide.
- All backend calls go through `frontend/src/api/`.
- Depends on: T02.

---

## T04 - Prototype: topic list and topic page

**User story:** US-1 (browse topics and lessons)
**Depends on:** T03

## Goal

The topic list and topic pages of the Bolt import (`frontend/src/pages/TopicsPage.tsx`, `frontend/src/pages/TopicPage.tsx`) read their data through the API client of T03 instead of the fixture and browser-storage modules, and carry the loading, empty, error and not-found states a real client needs.
Source: `product-spec.md` section 4 US-1; the Bolt prototype imported with issue #3 as the starting point.

## Acceptance criteria

- [ ] Both pages get their data only through `frontend/src/api/`: the home page lists every topic `listTopics` returns, in the order it returns them, with title and description, each linking to its topic page; the topic page resolves its slug through the client and lists that topic's lessons in the given order, each linking to its lesson page.
- [ ] Loading, empty and error states exist on both pages; a failing client call shows the error state with a retry action, and an unknown topic slug shows a "not found" state instead of an empty list.
- [ ] Neither page imports `frontend/src/data/content.ts`, `frontend/src/store.tsx` or reads browser storage; the artificial `setTimeout` latency of the import is gone and the loading state follows the client's pending promise.
- [ ] The lesson-completed and quiz-attempt counters the imported home page showed are removed, not ported: they are not in the specification; the per-topic completion view returns with T08.
- [ ] The five lint problems of the two files are gone: the unused `Clock` in `TopicsPage.tsx`, and the unused `FileText`, `CircleDot`, `ErrorState` and `topicSlug` in `TopicPage.tsx`.
- [ ] The `react-refresh/only-export-components` warning in `frontend/src/components/ui.tsx` is fixed by moving the `toast` helper out of the component file; every existing importer keeps working.
- [ ] `@supabase/supabase-js` is removed from `frontend/package.json` and `frontend/bun.lock`; nothing imports it.
- [ ] A component-testing harness exists and is used: the testing libraries and the DOM environment the page tests need are added as pinned devDependencies, and tests pass for list rendering, order, the three states and the unknown slug.

## Out of scope

- The lesson body (T05).
- Completion state (T08).
- Attempt counters and every history view (T09).
- Any real API call.

## Constraints

- Build only the screens in the specification; no extra screens or features for visual reasons.
- Keyboard access and labels for every interactive element; headings in order.
- The slug from the route is resolved through the client, with no second lookup mechanism.
- Depends on: T03.

---

## T05 - Prototype: lesson page

**User story:** US-1
**Depends on:** T04

## Goal

The lesson page of the Bolt import (`frontend/src/pages/LessonPage.tsx`) reads the lesson and its topic through the API client and shows the Markdown body, for anonymous and logged-in visitors alike.
Source: `product-spec.md` section 4 US-1, section 5 (lesson body in Markdown).

## Acceptance criteria

- [ ] The page loads the lesson through the client and its topic through the client, and shows the topic, the lesson title and the body rendered from Markdown as HTML.
- [ ] Raw HTML inside the Markdown body is not rendered as HTML.
- [ ] Loading, error and not-found states exist; a failing client call shows the error state with a retry action, and an unknown lesson id shows a "not found" state.
- [ ] The completion toggle, the completion badges and every read of `frontend/src/store.tsx` or the fixtures are gone from the page; they return with T08, backed by the client.
- [ ] Tests for the new behaviour exist and pass: a heading, a list and a code block in the Markdown body are rendered as HTML; raw HTML is shown as text; the title and topic are shown; the not-found state appears for an unknown id.
- [ ] The page is reachable and readable with the keyboard only, and the body is not exposed as raw Markdown.

## Out of scope

- Completion toggle and completion state (T08).

## Constraints

- `react-best-practices` and `composition-patterns` apply to the component structure.
- Depends on: T04.

---

## T06 - Prototype: quiz flow

**User story:** US-2 (take a quiz), US-3 (anonymous access)
**Depends on:** T03

## Goal

The quiz pages of the Bolt import (`frontend/src/pages/QuizPages.tsx`) start and submit the quiz through the API client, so the browser neither picks the questions nor grades the answers.
Source: `product-spec.md` section 4 US-2, US-3.

## Acceptance criteria

Groomed by issue #6 on 2026-10-20; the issue carries the full given/when/then wording.

- [ ] Starting a quiz through the client shows exactly the list returned by `api.startQuiz(topicId)`: 5 questions, or all when the topic has fewer, options in the order the client returned them, and no UI to change the quiz size.
- [ ] Before the submission response arrives, neither the DOM nor the page state contains a correct option, a marked option or an explanation: the rendered question data is limited to `QuizQuestion` from `frontend/src/api/types.ts`.
- [ ] An option can be selected and cleared again; submission works with some or all questions unanswered, and unanswered questions travel with `optionId: null`.
- [ ] The result view shows the score out of the total and, per question, correctness, the correct answer and the explanation, all from the submission response; the page computes none of it.
- [ ] No browser grading left: `frontend/src/lib/quiz.ts` is deleted, `pickQuizQuestions`, `makeSeed`, `gradeAnswer` and the page-local `QUIZ_SIZE` are gone from `QuizPages.tsx`, and the quiz result types the pages use come from `frontend/src/api/types.ts`.
- [ ] Nothing is written to `sessionStorage` or `localStorage`; `QuizPages.tsx` imports neither `@/store`, `@/data/content` nor `@/lib/quiz`; a reload with no submission in memory shows the empty state, never stale data.
- [ ] The imported start screen's question-bank size, attempts count and last-attempt card are removed, not ported; the signed-out banner that read `user` from the store goes with the store dependency.
- [ ] Loading indicators while the client calls are pending, an error state with retry when they reject (the `ErrorState` pattern of `LessonPage.tsx`), and no artificial latency timers around client calls.
- [ ] The play screen is keyboard operable: options focusable and selectable with Enter or Space, buttons and question dots reachable with Tab, the selected option visually distinguishable.
- [ ] Tests exist next to the pages and pass for the question list, the absence of correct-answer data before submission, an unanswered submission, the result view, the loading state and the error state with retry.

## Out of scope

- Server-side grading (T16) and the real HTTP client (T27-T30).
- Storing attempts (T19) and every history view (T09).
- Any login (T07).
- Deleting `AnswerOption.isCorrect` or the `QuizResultAnswer`/`QuizAttempt` types from `frontend/src/types.ts`: the fixture field is what the mock grades with, and the attempt types are still used by `store.tsx` and `HistoryPages.tsx` until T09 removes them (moved to T09 by the grooming of issue #6).
- Changes to `openapi.yaml`: the contract tasks come after the prototype stage.

## Constraints

- All quiz data flows through the client in `frontend/src/api/`; no direct `fetch`.
- Correct answers must never be present in the client data before submission.
- Depends on: T03.

---

## T07 - Prototype: registration, log in and log out

**User story:** US-4 (register and log in)
**Depends on:** T03

## Goal

The auth page and the shell of the Bolt import (`frontend/src/pages/AuthPage.tsx`, `frontend/src/components/AppShell.tsx`, `frontend/src/App.tsx`) manage the session through the API client, and the browser-storage account simulation of `frontend/src/store.tsx` is replaced.
Source: `product-spec.md` section 4 US-4.

## Acceptance criteria

- [ ] Registration and log-in call the client and validate the fields the client's `Credentials` carry: email and password, with the password limits of the specification (8 to 128 characters); the error returned for invalid or duplicate input is shown next to the form.
- [ ] The name field of the imported registration form and the display name in the header are gone: the client types carry no name, and the header shows the email of the logged-in user.
- [ ] Registering or logging in updates the session state in the app; the header shows the logged-in state, and logging out clears it. The session is held in memory only: nothing writes it to `localStorage` or `sessionStorage`, and a reload starts the prototype anonymous (the real session arrives with T29).
- [ ] The seeded demo account, the plaintext passwords and every `forge.*` key of the imported store are gone; the session state is only read and written through the API client, never from the store's browser storage.
- [ ] The `react-refresh/only-export-components` warning in `frontend/src/store.tsx` is gone: the session state this task introduces lives in a file that exports no plain functions next to components.
- [ ] Password inputs are not echoed into logs or error messages.
- [ ] Tests for the new behaviour exist and pass for the validation error, the duplicate-email error, the logged-in header state and the logged-out state.

## Out of scope

- Cookie handling, CSRF and password hashing (T11, T17).
- Password rules beyond what the contract defines (see Decisions).
- Persisting the session across a reload (T29).
- Progress and quiz history, which stay on the old store until T08 and T09.

## Constraints

- Labels on every input, errors announced to assistive technology, keyboard operable.
- Depends on: T03.

---

## T08 - Prototype: lesson completion

**User story:** US-5 (track lesson progress)
**Depends on:** T06, T07

## Goal

A logged-in visitor can mark a lesson complete and undo it through the API client, and sees the completion state on the lesson and topic pages.
Source: `product-spec.md` section 4 US-5.

## Acceptance criteria

- [ ] The lesson page shows a control that reflects the current completion state for the logged-in user, read through the client, and toggles it both ways through the client.
- [ ] The topic page shows how many lessons of that topic are completed and which ones, from the client's progress.
- [ ] An anonymous visitor sees no completion control.
- [ ] While the request is in flight the control is disabled; a failed request shows an error and leaves the shown state unchanged.
- [ ] No completion state is read from `frontend/src/store.tsx` or browser storage anymore: the progress path of the imported store is removed in this task.
- [ ] Tests for the new behaviour exist and pass for completed, not completed, anonymous and failed-request cases.

## Out of scope

- Real persistence and ownership checks (T18, T22).
- Progress percentages or streaks: not in the specification.

## Constraints

- All calls through `frontend/src/api/`.
- Depends on: T06, T07.

---

## T09 - Prototype: quiz history

**User story:** US-6 (quiz history)
**Depends on:** T07

## Goal

A logged-in user can list their past quiz attempts and open one to see its detail, all through the API client, and the frontend prototype stage ends with every check green.
Source: `product-spec.md` section 4 US-6.

## Acceptance criteria

- [ ] A history page lists the user's attempts with topic, date and score, newest first, and links to an attempt page; the list and the detail come through the client.
- [ ] The attempt page shows, per question, the question text, the selected answer text (empty when unanswered) and the correct answer text.
- [ ] The mock data demonstrates the snapshot: an attempt shows the question and answer wording stored with the attempt, which differs from the current content; the fixtures may be adjusted for this in `frontend/src/data/content.ts`.
- [ ] An empty state is shown when the user has no attempts.
- [ ] An anonymous visitor is directed to log in instead of being shown an empty history.
- [ ] The history path of the imported store is removed; `frontend/src/store.tsx` is deleted when nothing of it remains, and no page imports the store or the Bolt fixtures directly.
- [ ] The quiz result types only this path needed are deleted with it: `QuizResultAnswer`, `QuizAttempt` and the other now-unused exports of `frontend/src/types.ts`, and `AnswerOption.isCorrect` stays only as long as `frontend/src/api/mock.ts` grades with it. Moved here from T06 by the grooming of issue #6.
- [ ] Tests for the new behaviour exist and pass for the list order, the detail view and the empty state.
- [ ] `(cd frontend && bun run lint)`, `(cd frontend && bun run test)` and `(cd frontend && bun run typecheck)` all exit 0, with no lint problem left from the imported files.

## Out of scope

- Filtering, sorting and pagination: not in the specification.
- Real storage (T19, T22).

## Constraints

- Depends on: T07.

---

## T10 - Contract: content and quiz endpoints

**User story:** US-1, US-2, US-3
**Depends on:** T03, T04, T05, T06

## Goal

`openapi.yaml` defines the public content and quiz interface, so the backend can be written against a contract and the frontend knows what to call.
Source: `product-spec.md` section 4 US-1, US-2, US-3, section 7 (API contract).

## Acceptance criteria

- [ ] `openapi.yaml` defines: list topics, get one topic with its lessons, get one lesson, start a quiz on a topic, submit a quiz for a topic, and a `GET /health` used by the containers.
- [ ] The start-quiz response contains the questions with their options and **no** field that reveals the correct option or the explanation; the submit request lists the question and the chosen option per question, where the option may be absent.
- [ ] The submit response contains the score, the total and per question the correctness, the correct option text and the explanation.
- [ ] No content or quiz endpoint declares a security requirement; anonymous submission is allowed.
- [ ] One error object is defined once and reused by every endpoint, with a machine-readable `code`, a user-safe `message` and optional field `details`; the codes 400, 404, 409 and 500 are documented.
- [ ] `bunx @redocly/cli lint openapi.yaml` passes.
- [ ] Topics are addressed by slug; lessons are addressed by id, because the data model has no lesson slug (`product-spec.md` section 9).

## Out of scope

- Authentication, progress, history and maintainer endpoints (T11-T13).
- Anything outside version 1.

## Constraints

- Contract first: no implementation and no real API call before this exists.
- Add no field or behaviour that the specification does not describe.
- Depends on: T03, T04, T05, T06.

---

## T11 - Contract: authentication and session

**User story:** US-4
**Depends on:** T10

## Goal

`openapi.yaml` defines registration, log in, log out and the current session, including how the session cookie and CSRF protection work.
Source: `product-spec.md` section 4 US-4, section 7 (security).

## Acceptance criteria

- [ ] Register, log in, log out and "current session" endpoints are defined; the session endpoint answers 401 when there is no valid session.
- [ ] The current-session endpoint returns the CSRF token in its response, so a client that keeps the session cookie across a page reload can send state-changing requests without re-authenticating.
- [ ] The session is documented as a cookie that is `HttpOnly` and `SameSite`, `Secure` in production, path `/`, with a fixed expiry of 7 days.
- [ ] Cross-site request forgery protection is documented: the token is issued at registration, log in and on every read of the current session, and must be sent back in a header on every state-changing request that carries a session cookie. `GET` does not require it. A quiz submission that carries a session cookie is state-changing and needs the token; an anonymous submission without a session cookie is accepted without it.
- [ ] In `openapi.yaml`, the `POST /topics/{topicSlug}/quiz/submissions` operation declares no `401` response and states that a submission with a missing, invalid or expired session cookie is treated as anonymous: it is graded and answered normally and nothing is stored. The operation declares `403` only for a CSRF failure when the session is valid.
- [ ] The shared `Error` schema documents the error codes for status `401` (`authentication_error`: no valid session or authentication failed), status `403` (`forbidden`: CSRF protection failed with a valid session) and status `403` (`authorization_error`: a valid session lacks the required permission, for example a user without `is_maintainer`); each operation that returns these statuses declares the matching response.
- [ ] The 403 decision is recorded in the contract: a client retries a `forbidden` response after refreshing the CSRF token and does not retry an `authorization_error` response. T13 and T20 use the same `authorization_error` code for a valid session that lacks the required permission.
- [ ] Register returns a validation error for invalid input and a conflict for a duplicate email; log in returns an authentication error for wrong credentials.
- [ ] Registration rejects a password shorter than 8 or longer than 128 characters with a validation error; the limits are part of the request schema.
- [ ] No response schema contains `password_hash` or any session token other than the documented CSRF token.
- [ ] `bunx @redocly/cli lint openapi.yaml` passes.

## Out of scope

- Password reset by email and email verification: non-goals.

## Constraints

- The password hashing algorithm is not part of the contract; it is chosen in T17.
- Depends on: T10.

---

## T12 - Contract: progress and quiz history

**User story:** US-5, US-6
**Depends on:** T10

## Goal

`openapi.yaml` defines how a logged-in user marks lessons complete and reads their own quiz attempts.
Source: `product-spec.md` section 4 US-5, US-6.

## Acceptance criteria

- [ ] Reading the current user's progress, marking a lesson complete and undoing it are defined; all three require a session and answer 401 without one.
- [ ] Marking complete is idempotent and undoing an unmarked lesson is not an error.
- [ ] Listing the current user's attempts and reading one attempt are defined, both requiring a session.
- [ ] The attempt answer schema has `question_id` nullable, `selected_option_text` allowed to be an empty string, and the three snapshot text fields.
- [ ] No endpoint takes a user id in the request, so one user cannot address another user's data.
- [ ] `bunx @redocly/cli lint openapi.yaml` passes.

## Out of scope

- Pagination, filtering and sorting of attempts.
- Any statistic or leaderboard.

## Constraints

- Ownership is enforced by the server from the session, never from the request body.
- Depends on: T10.

---
