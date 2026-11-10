# Backlog: tasks from the product specification

Derived from [`product-spec.md`](../../../product-spec.md) (draft v0.2) on 2026-10-17, following [`docs/task-template.md`](../../task-template.md) and the order in [`AGENTS.md`](../../../AGENTS.md).

- `T02`-`T13` became GitHub issues on 2026-10-17, where the issue number matches the task number (`T02` is #2, `T13` is #13); in those issues `Depends on: Txx` is written as `Depends on: #<issue number>`.
- `T01` was issue #1. It was cancelled by the maintainer on 2026-10-18 and removed from this file: the scaffold it asked for contradicts the project order of `AGENTS.md`, the frontend project comes from Bolt and the Go project belongs to `T14`, after the contract. Nothing of it was carried over to another task.
- `T34a` (#89) and `T35b` (#92) were cancelled by the maintainer on 2026-11-10 and removed from this file: Render builds the backend and frontend images from the repository Dockerfiles, so there is no registry, image push, deploy hook or `render.yaml`. Nothing was renumbered. Deployment happens only after the CI checks of `T33a` and `T33b` pass, through Render's *After CI Checks Pass* Auto-Deploy setting.
- `T27`-`T37` were reworked on 2026-11-07 against the merged backend (#34-#52) and split into a/b tasks; every task has a GitHub issue (#75-#94) and every `Depends on:` line uses issue numbers. The issues were created from this file in dependency order on the same day.
- `T02`-`T12` are done and merged into `main`; their text is kept unchanged in [`done.md`](done.md). The issues hold the same text, so open tasks are read here and finished ones there.
- Each task fits one session and can be read on its own.
- This file creates no code. The maintainer split T27-T37 before the issues were created (see [`docs/process.md`](../../process.md), "Backlog"); the text here and in the issues is the same.
- `T04`-`T09` were revised on 2026-10-19, after the Bolt prototype was imported and the API client of issue #3 was merged: each of them now adapts and verifies the pages that already exist in `frontend/` instead of building them, carries the lint problems of the imported files it touches, and `T09` closes the prototype stage with `bun run lint` exiting 0. Issues #4-#9 were updated in the same change. Nothing was merged, split or renumbered.
- `T32` precedes `T31` because the frontend integration suite starts the database with the compose file. `T35a` precedes `T34b` because the Render services and the Auto-Deploy setting must exist before the deploy gate can be verified.
- No task adds behaviour outside version 1 of the specification.
- Every task ends with an entry file in [`docs/ai-log/`](../../ai-log/) (rules in [`docs/ai-workflow.md`](../../ai-workflow.md)) and a committed stage artifact; the retrospective pass over that bookkeeping is T37.
- `T39a`-`T41b` were added on 2026-11-11 with issues #109-#114, after #84 and #87: the observability, alerting and on-call work is local only and is not deployed. `T36` and `T37` now depend on these tasks.

## Order

| Stage | Tasks | User stories |
|---|---|---|
| Frontend prototype with mock API client | T02-T09 | US-1, US-2, US-3, US-4, US-5, US-6 |
| API contract | T10-T13 | US-1, US-2, US-3 / US-4 / US-5, US-6 / US-7 |
| Backend by user story | T14-T20 | US-1, US-2, US-3, US-4, US-5, US-6, US-7 |
| Persistence: migrations, repositories, seed | T21-T25 | US-1, US-4, US-5, US-6, US-7, US-8 |
| Backend integration tests | T26 | US-1 ... US-6 |
| Connect frontend to backend | T27a-T27b, T28a-T28b, T29a-T29b, T30a-T30b | US-1, US-2, US-3, US-4, US-5, US-6 |
| Containers | T32a-T32b | none (supports all) |
| Frontend integration tests | T31a-T31b | US-1 ... US-6 |
| CI | T33a-T33b | none (supports all) |
| CD and deployment | T35a, T34b | none (supports all) |
| Observability, alerting and on-call (local) | T39a-T41b | none (supports all) |
| Documentation | T36-T37 | none (supports all) |

Tasks not covered by a user story come from section 7 (non-functional requirements) or from course deliverables the specification does not describe; each one names its source.

---

## T13 - Contract: maintainer draft and publish operations

**User story:** US-7 (content preparation)
**Depends on:** #10

## Goal

`openapi.yaml` defines the operations an AI tool and the maintainer use to prepare drafts, and the separate action that publishes them.
Source: `product-spec.md` section 4 US-7, section 6.

## Acceptance criteria

- [ ] Creating and editing draft lessons and draft questions is defined for the topics and their lessons and questions.
- [ ] Publishing and unpublishing are separate explicit endpoints that only change the status; no draft operation can publish.
- [ ] A question schema requires 2 to 6 options and exactly one correct option, and the request is rejected otherwise.
- [ ] Every operation here requires a session with `is_maintainer`; the responses for a missing session (401) and for a user without the flag (403) are documented.
- [ ] None of these operations appears on a public or user endpoint.
- [ ] `bunx @redocly/cli lint openapi.yaml` passes.

## Out of scope

- An admin web interface: a non-goal.
- How the AI agent authenticates and which exact operations it exposes: decided with the extension pack.

## Constraints

- Agents may create and edit drafts but must never publish.
- Depends on: #10.

---


**Note:** Contract issue #32 (404 for progress operations) created. T18 depends on #32.

## T14a - Backend foundation: server, configuration and /health (#34)

**User story:** none - supports US-1 to US-7  
**Depends on:** #10, #11, #12, #13

## Goal

The Go service starts, reads its configuration from the environment, and answers `GET /health` with 200, so later backend tasks have a running server to build on.

## Acceptance criteria

- [ ] `backend/cmd/api` starts an HTTP server on the address from the environment and answers `GET /health` with 200.
- [ ] Configuration is read from environment variables with development defaults; a missing required variable stops the server with a clear message, and every variable is documented in `.env.example`.
- [ ] Requests are logged with method, path, status and duration, without request bodies, cookies, passwords or hashes.
- [ ] Handlers are thin and repository interfaces exist next to services, with an in-memory implementation used by unit tests.
- [ ] Tests for the new behaviour exist and pass.

## Out of scope

- Any domain endpoint other than `/health`.
- Error mapping details (T14b).
- The database, migrations and the real repositories (T21-T22).

## Constraints

- SQL only in the repository layer; logic in services.
- Depends on: #10, #11, #12, #13.

---

## T14b - Backend foundation: error type and mapping (#35)

**User story:** none - supports US-1 to US-7  
**Depends on:** #34

## Goal

One domain error type maps to the contract error object with correct codes, so later backend tasks only add handlers and services.

## Acceptance criteria

- [ ] One domain error type maps to 400 (validation_error), 401 (authentication_error), 403 (forbidden for CSRF failure with valid session; authorization_error for valid session without is_maintainer), 404 (not_found), 409 (conflict), and 500 (internal_error) and always writes the error object defined in the contract.
- [ ] A handler that returns a service error produces the correct status and error shape.
- [ ] Tests for each status mapping exist and pass.

## Out of scope

- Any domain endpoint logic beyond error mapping.
- Database, migrations and real repositories (T21-T22).

## Constraints

- Errors must follow the contract shape exactly.
- Depends on: #34.

---
## T15 - Backend: public content endpoints (#36)

**User story:** US-1  
**Depends on:** #35

## Goal

The API serves published topics, their lessons in order, and a single lesson, so the content screens have something real to call. Topics are addressed by slug, lessons by id.

## Acceptance criteria

- [ ] List topics returns only published topics, ordered by position, with the fields of the contract.
- [ ] Get topic returns the published topic with its published lessons in position order.
- [ ] Get lesson returns the published lesson with its Markdown body.
- [ ] An unknown slug and a draft topic, lesson or question each answer 404; no draft content is ever returned.
- [ ] The published-only rule lives in the service, and the repository filters on status.
- [ ] Unit tests exist and pass for the list, the ordering, draft exclusion and both 404 cases.

## Out of scope

- Progress state on the topic page (T18).
- Markdown validation or content versioning.

## Constraints

- Handlers stay thin; responses match `openapi.yaml` exactly. JSON is snake_case.
- Depends on: #35.

---
## T16a - Backend: start quiz (#37)

**User story:** US-2  
**Depends on:** #35

## Goal

The API picks 5 random published questions for a topic, so quiz start is available without revealing answers.

## Acceptance criteria

- [ ] Starting a quiz returns exactly 5 published questions of that topic, or all of them when the topic has fewer; two calls can return different question sets.
- [ ] The start response contains no field that reveals the correct option or the explanation.
- [ ] Topics are addressed by slug. The contract has no quiz id.
- [ ] Unit tests exist and pass for the count and random selection.

## Out of scope

- Grading and submission (T16b).
- Storing attempts (T19).

## Constraints

- Quiz is stateless. Depends on: #35.

---

## T16b - Backend: grade and submit quiz (#38)

**User story:** US-2, US-3  
**Depends on:** #37

## Goal

The server grades submissions on the server, so the browser never decides the score. Quiz submission works without a session; anonymous is preserved.

## Acceptance criteria

- [ ] Grading is done server-side: an unanswered question is incorrect, an option that does not belong to the question is incorrect, and the submission still succeeds.
- [ ] A submission containing a question that does not belong to the topic is rejected with a 400 validation_error and nothing is graded.
- [ ] The result contains the score, the total and, per question, the correctness, the correct option text and the explanation.
- [ ] A submission without a session is graded and answered normally (no 401). An invalid or expired cookie means anonymous.
- [ ] The first occurrence of a question id counts; later duplicates are ignored. An empty answer list is a validation_error.
- [ ] Unit tests exist and pass for incorrect cases, rejected submission, duplicates handling, empty list, and anonymous submission.

## Out of scope

- Storing attempts (T19).
- Session-required behavior for submission.

## Constraints

- The submission carries question and chosen option ids and the server checks they belong to the topic. Depends on: #37.

---
## T17a - Backend: registration, login, logout, current session (#39)

**User story:** US-4  
**Depends on:** #35

## Goal

A visitor can create an account, log in with a session cookie, log out, and be recognised as anonymous afterwards. Passwords stored as Argon2id.

## Acceptance criteria

- [ ] Register validates the input and rejects a duplicate email with 409 conflict; the password is stored only as a salted Argon2id hash (8-128 characters), no plain password kept anywhere.
- [ ] Log in verifies the hash, creates a session that expires after 7 days, and sets the session cookie with attributes: HttpOnly, SameSite=Lax, Secure in production, path /, 7 days; a wrong password returns an authentication_error (401).
- [ ] Log out deletes the session and clears the cookie.
- [ ] A missing or expired session is treated as not logged in: the current-session endpoint answers 401 (authentication_error).
- [ ] Passwords, hashes and session tokens never appear in a response or in a log.
- [ ] Unit tests exist and pass for register, duplicate email, log in, wrong password, log out, expired session.

## Out of scope

- CSRF token issuing/checking (T17b).
- Roles other than is_maintainer (T20).

## Constraints

- Session state comes from the cookie, never from request body or query parameter. Depends on: #35.

---

## T17b - Backend: CSRF token issuing and checking (#40)

**User story:** US-4  
**Depends on:** #39

## Goal

CSRF protection is implemented as specified: csrf_token issued by register, login and every GET /auth/session; header X-CSRF-Token required only for state-changing requests that carry a session cookie.

## Acceptance criteria

- [ ] csrf_token is issued by register, login and every GET /auth/session.
- [ ] X-CSRF-Token header is required only for state-changing requests that carry a session cookie.
- [ ] A request without a session cookie is not checked for a token (anonymous requests remain unaffected). Quiz submission without session cookie needs no token.
- [ ] State-changing requests with a session cookie but missing/invalid CSRF token return 403 with code forbidden (CSRF failure).
- [ ] Unit tests exist and pass for CSRF token issuing and checking.

## Out of scope

- Other auth flows.
- Depends on: #39.

---
## T18 - Backend: lesson progress (#41)

**User story:** US-5  
**Depends on:** #40, #32

## Goal

A logged-in user can mark a lesson complete and undo it, and read their own completion state. PUT/DELETE /me/progress/lessons/{lessonId} return 404 for unknown or unpublished lesson.

## Acceptance criteria

- [ ] Marking a lesson complete records the completion time and is idempotent; undoing removes it and is safe to repeat.
- [ ] Reading progress returns only the current user's completed lessons via GET /me/progress.
- [ ] Every one of the three endpoints answers 401 without a valid session (authentication_error).
- [ ] PUT and DELETE return 404 (not_found) for an unknown or unpublished lesson.
- [ ] No request carries a user id, so one user cannot change or read another user's progress.
- [ ] Unit tests exist and pass for mark, repeated mark, undo, read, unauthenticated case, and 404 cases.

## Out of scope

- Progress statistics and streaks.
- Depends on: #40, #32.

---
## T19 - Backend: quiz attempt history (#42)

**User story:** US-6, US-3  
**Depends on:** #38, #40

## Goal

Every submission from a logged-in user is stored with a snapshot of the questions and answers. The server stores the question, selected option and correct option texts in the attempt snapshot. Attempt list has no topic_title; client maps from topics.

## Acceptance criteria

- [ ] A submission with a valid session stores one attempt with topic, time, score and total, and one answer row per question.
- [ ] Every submission creates a new attempt; an earlier attempt is never overwritten.
- [ ] Each stored answer keeps the question text, the selected option text (empty when unanswered) and the correct option text as they were at the time.
- [ ] A submission without a session stores nothing and is answered normally.
- [ ] Listing attempts returns only the current user's attempts, newest first (no topic_title in list); reading an attempt that belongs to another user answers 404.
- [ ] Unit tests exist and pass for storage per submission, snapshot wording, anonymous case and ownership case.

## Out of scope

- Deleting attempts and exporting them.
- Depends on: #38, #40.

---
## T20a - Backend: maintainer create and edit drafts (#43)

**User story:** US-7  
**Depends on:** #36, #38, #40

## Goal

Maintainer can create and edit draft lessons and questions; no endpoint to list drafts (maintainer gets id from create response).

## Acceptance criteria

- [ ] A session without `is_maintainer` answers 403 (authorization_error) and a missing session answers 401 (authentication_error) on every draft operation.
- [ ] A valid session without is_maintainer gets 403 with code authorization_error.
- [ ] Creating and editing a lesson or a question stores it as a draft and never as published.
- [ ] A question with fewer than 2 or more than 6 options, or without exactly one correct option, is rejected with a validation_error.
- [ ] Topics/lessons addressed by slug/id as per contract; JSON snake_case.
- [ ] Unit tests exist and pass for authorisation cases and validation cases.

## Out of scope

- Publish and unpublish (T20b).
- Admin UI.

## Constraints

- Depends on: #36, #38, #40.

---

## T20b - Backend: maintainer publish and unpublish (#44)

**User story:** US-7  
**Depends on:** #43

## Goal

Publishing and unpublishing are separate explicit endpoints that only change status; nothing publishes as a side effect.

## Acceptance criteria

- [ ] A session without `is_maintainer` answers 403 (authorization_error) and a missing session answers 401 (authentication_error) on publish/unpublish operations.
- [ ] Publishing sets status to published and makes content visible through public endpoints of T15 and T16; unpublishing hides it again.
- [ ] No draft operation can publish.
- [ ] Unit tests exist and pass for publish/unpublish and visibility changes.

## Out of scope

- Create/edit drafts (T20a).
- Depends on: #43.

---
## T21a - Persistence: content tables and migrations (#45)

**User story:** US-1 (also serves US-2 and US-7)  
**Depends on:** #36, #38, #44

## Goal

Topics, lessons, questions and answer options live in PostgreSQL through numbered migrations.

## Acceptance criteria

- [ ] Numbered migration files create `topics`, `lessons`, `questions` and `answer_options` with fields of section 9; topic slug unique, positions unique per parent, every table has status.
- [ ] Migrations applied in order and recorded; applying twice leaves same schema.
- [ ] `DATABASE_URL` and test database variable documented in `.env.example`.
- [ ] Migration tests/checks pass.

## Out of scope

- Repository implementations (T21b).
- Users/sessions/progress/attempts (T22a).

## Constraints

- Schema changes via numbered migration files only. Depends on: #36, #38, #44.

---

## T21b - Persistence: content repositories (#46)

**User story:** US-1 (also serves US-2 and US-7)  
**Depends on:** #45

## Goal

PostgreSQL implementations replace in-memory content repositories behind same interfaces.

## Acceptance criteria

- [ ] PostgreSQL implementations replace in-memory content repositories behind the same interfaces; service layer unchanged.
- [ ] An integration test (build tag `integration`) proves published filter and position ordering against test database.
- [ ] Existing unit tests still pass unchanged; integration test passes against `docker compose up -d db`.
- [ ] Integration tests never point at development database.

## Out of scope

- Other domain tables (T22a).
- Depends on: #45.

---
## T22a - Persistence: users, sessions, progress and attempt tables (#47)

**User story:** US-4, US-5, US-6  
**Depends on:** #40, #41, #42, #46

## Goal

Migrations for users, sessions, lesson_progress, quiz_attempts and quiz_attempt_answers.

## Acceptance criteria

- [ ] Migrations create `users` (unique email, password hash, `is_maintainer` default false, created at) and `sessions` (id, user id, created at, expires at, indexes on user id and expires at).
- [ ] Migrations create `lesson_progress` with user and lesson as key and completion timestamp.
- [ ] Migrations create `quiz_attempts` (user, topic, score, total, created at) and `quiz_attempt_answers` (attempt, nullable question reference, question text, selected option text, correct option text, correctness).
- [ ] Migrations numbered in order so quiz_attempts comes after users.
- [ ] `DATABASE_URL`/test vars documented.

## Out of scope

- Repository implementations (T22b).
- Depends on: #40, #41, #42, #46.

---

## T22b - Persistence: users, sessions, progress and attempt repositories (#48)

**User story:** US-4, US-5, US-6  
**Depends on:** #47

## Goal

Postgres repos for auth, progress, attempts with snapshot behavior.

## Acceptance criteria

- [ ] Repositories behind registration, login, logout, progress, attempt storage and reading use PostgreSQL behind existing interfaces.
- [ ] Integration test registers, logs in, finds session, marks lesson, restarts connection and reads state back; expired session refused after restart.
- [ ] Integration test stores two attempts for one user, changes question text in DB, reads both: stored snapshot unchanged, new attempt uses new text.
- [ ] Deleting a question does not break reading an old attempt.
- [ ] Password hash never selected into response type or log line.
- [ ] Integration tests never point at development database.

## Out of scope

- Seed (T23).
- Depends on: #47.

---
## T23 - Seed mechanism and content for the first topic (#49)

**User story:** US-8  
**Depends on:** #48

## Goal

Seed content loaded from versioned files on first start, repeatable, no duplicates. First topic (LLMs) has >=1 published lesson and >=6 published questions.

## Acceptance criteria

- [ ] Seed content lives in versioned files under `backend/`, reviewed in diff; no runtime generation.
- [ ] Documented command and backend start both load seed; loader inserts missing and reports skipped.
- [ ] Loading twice creates no duplicates; second run reports nothing created.
- [ ] First topic (LLMs) has at least 1 published lesson and at least 6 published questions, each 2-6 options, exactly one correct, explanation.
- [ ] Integration test loads seed into empty test DB, loads again, asserts counts unchanged.

## Out of scope

- Other six topics (T24).
- Depends on: #48.

---
## T24 - Seed content for the remaining six topics (#50)

**User story:** US-8  
**Depends on:** #49

## Goal

All seven topics have real learning material.

## Acceptance criteria

- [ ] AI agents, MCP, RAG, embeddings, evaluation and deployment each have at least 1 published lesson and at least 6 published questions in seed files.
- [ ] Every question has 2-6 options, exactly one correct option, short explanation.
- [ ] Loader unchanged; fresh DB gets all seven topics, second run creates nothing.
- [ ] Integration test asserts minimum counts for each of the seven topics after one load.

## Out of scope

- More than minimum, translations, images/videos.
- Depends on: #49.

---
## T25 - Maintainer account and the publish flow end to end (#51)

**User story:** US-7  
**Depends on:** #44, #50

## Goal

Maintainer account exists after seeding; draft-to-published path works against real DB.

## Acceptance criteria

- [ ] Seed creates user with `is_maintainer` true, credentials from env vars; no password committed.
- [ ] Integration test logs in as maintainer, creates draft lesson and question, sees neither through public endpoints, publishes them, sees both, unpublishes them, sees neither.
- [ ] Integration test with normal user gets 403 on draft/publish operations.
- [ ] Integration test with no session gets 401 on same operations.
- [ ] Seeded maintainer account documented in `.env.example` with needed variables.

## Out of scope

- Admin UI, roles beyond flag, AI tool that publishes.
- Depends on: #44, #50.

---
## T26 - Backend integration tests for the key workflows (#52)

**User story:** US-1, US-2, US-3, US-4, US-5, US-6  
**Depends on:** #51

## Goal

Whole API exercised end to end against real database.

## Acceptance criteria

- [ ] Integration tests (build tag `integration`) drive HTTP API against test database for: browsing published content, starting and submitting quiz, registering, logging in/out, saving progress, reading quiz history.
- [ ] Also assert unpublished content never appears in public response and submission with foreign question is rejected.
- [ ] Suite creates and cleans its own data and can run twice in a row against same test database.
- [ ] Documented command `(cd backend && go test -tags=integration ./...)` runs only these tests; unit tests still run without database.
- [ ] Results recorded in stage artifact `docs/artifacts/integration/backend-results.md` and linked from issue.

## Out of scope

- Repository-level SQL tests (already in T21b/T22b).
- Load/performance tests.
- Depends on: #51.

---
## T27a - Frontend: real HTTP API client (#75)

**User story:** US-1, US-2, US-3, US-4, US-5, US-6  
**Depends on:** #3, #52

## Goal

The `ApiClient` has a real HTTP implementation selected by one environment variable, so the screens can call the finished backend of `openapi.yaml` without knowing which implementation is active.
Source: `product-spec.md` section 7 (one client module, contract), `openapi.yaml`.

## Acceptance criteria

- [ ] `VITE_API_MODE` selects `mock` or `http`, and `VITE_API_BASE_URL` is the only base-URL source; a production build never selects the mock, and no component imports the HTTP implementation.
- [ ] The client implements the 13 existing operations plus `GET /auth/session`, using the methods, paths, request bodies and status codes of `openapi.yaml`; JSON snake_case is mapped to the camelCase screen types and back.
- [ ] Contract responses are mapped into the screen types, including the four fields the contract does not carry: `icon` and `color` from the topic slug, `summary` from the first body paragraph, and `durationMin` from body word count at one documented rate; an empty body falls back without an error.
- [ ] Topic operations take a slug (`getTopic(slug)`, `startQuiz(slug)`, `submitQuiz({ topicSlug, ... })`); the mock and the existing call sites are updated so the app and typecheck stay green.
- [ ] Requests send cookies (`credentials: 'include'`); the client keeps the CSRF token from auth responses and `GET /auth/session` in memory and sends `X-CSRF-Token` on state-changing requests when a token is present.
- [ ] HTTP and network failures raise one typed error carrying the HTTP status, the contract `code` and `message`; 401, 403 `forbidden`, 404, 409 and 500 are distinguishable.
- [ ] Unit tests cover path/method/body mapping, snake_case conversion, the presentation-field mapper, cookies, the CSRF header and status/code mapping; `bun run lint`, `bun run typecheck` and `bun run test` pass.

## Out of scope

- Wiring the screens to the real client: T27b-T30b.
- Integration tests: T31a-T31b.
- Maintainer operations: no screen uses them.

## Constraints

- All backend calls go through `frontend/src/api/`; components never call `fetch`.
- No session token, CSRF token or password in `localStorage` or `sessionStorage`.
- Development uses the Vite proxy `/api` to the backend, so the browser stays same-origin and the backend needs no CORS change.
- Depends on: #3, #52.

---

## T27b - Frontend: content screens on the backend (#76)

**User story:** US-1  
**Depends on:** #75, #4, #5, #52

## Goal

The topics, topic and lesson screens read the seeded backend instead of the mock.
Source: `product-spec.md` section 4 US-1, section 7 (contract).

## Acceptance criteria

- [ ] The topics page renders `GET /topics` in `position` order; the topic page calls `getTopic(slug)` directly; the lesson page calls `getLesson(id)`; no screen resolves a topic id through `listTopics`.
- [ ] A 404 from an unknown topic or lesson shows the existing not-found state; a 500 or network failure shows the retry state; loading and empty states remain.
- [ ] A manual run against the running seeded stack shows all seven topics, their lessons in order and the lesson Markdown; no draft content is visible.
- [ ] Component tests cover success, 404, 500 and empty against the fake client; unit tests cover the slug calls; the existing tests pass.

## Out of scope

- Quiz, accounts, progress and history: T28a-T30b.
- Changing the visual design.

## Constraints

- All backend calls go through `frontend/src/api/`.
- Topics are addressed by slug, lessons by id, as `openapi.yaml` defines.
- Depends on: #75, #4, #5, #52.

---

## T28a - Frontend: quiz play on the backend (#77)

**User story:** US-2, US-3  
**Depends on:** #75, #76, #6, #38

## Goal

Starting and submitting a quiz call the backend, so the score the user sees is the server's score.
Source: `product-spec.md` section 4 US-2, US-3.

## Acceptance criteria

- [ ] The start and play screens load the topic by slug and call `startQuiz(slug)`; no topic id lookup remains.
- [ ] Questions render in the server's order exactly as returned, including fewer than five when the topic has fewer; no correct option or explanation is shown before the submission response.
- [ ] The submission sends one `{ questionId, optionId | null }` per question under the topic slug; it attaches the CSRF token when a session token is present, and an anonymous submission sends none and is still graded.
- [ ] The screen renders the server's `score`, `total` and per-answer `isCorrect` and computes none of them; a foreign-question 400 is shown as a validation error; an unanswered question is sent as `null`.
- [ ] Unit tests cover five and fewer questions, unanswered, the CSRF header, the 400 case and retry; the existing tests pass.

## Out of scope

- The result review mapping: T28b.
- Storing attempts in the UI and history: T30b.

## Constraints

- The quiz size is not configurable in the UI.
- Depends on: #75, #76, #6, #38.

---

## T28b - Frontend: quiz result review (#78)

**User story:** US-2  
**Depends on:** #77

## Goal

The result screen reviews the server's grading and shows the question and selected answer without regrading in the browser.
Source: `product-spec.md` section 4 US-2.

## Acceptance criteria

- [ ] The result screen shows the server's score, per-question correctness, correct answer text and explanation; it computes no score or correctness.
- [ ] Question text and selected answer text are filled from the in-memory quiz questions and answers, because `QuizSubmissionResultAnswer` does not carry them; an unanswered question shows "No answer selected".
- [ ] The result is not written to browser storage; reloading the result route keeps the existing not-found state.
- [ ] Unit tests cover correct, incorrect, unanswered, explanation and reload.

## Out of scope

- The history detail view: T30b.

## Constraints

- The submission response stays the source for score, correctness, correct answer and explanation.
- Depends on: #77.

---

## T29a - Frontend: registration, login and logout (#79)

**User story:** US-4  
**Depends on:** #75, #7, #39

## Goal

Registration, login and logout call the backend with the cookie session and the CSRF token.
Source: `product-spec.md` section 4 US-4, section 7 (security).

## Acceptance criteria

- [ ] The auth page calls register, login and logout through the client; 400 validation and 409 duplicate-email messages from `Error.message` are shown in the form; success updates the session state.
- [ ] The session cookie is `HttpOnly` and browser-managed; no session token, CSRF token or password is written to `localStorage` or `sessionStorage`.
- [ ] The client keeps the token from register and login in memory and attaches it to logout and other state-changing requests; a 403 `forbidden` refreshes the token with `GET /auth/session` and retries once, while `authorization_error` is not retried.
- [ ] The user type mirrors `UserSummary` (`id`, `email`); the prototype-only maintainer badge and `createdAt` are removed because the contract does not expose them.
- [ ] Unit tests cover validation and duplicate email, CSRF attach and retry, and assert that browser storage stays empty.

## Out of scope

- Session bootstrap and expiry handling: T29b.
- Progress and history: T30a-T30b.

## Constraints

- No token in browser storage; the cookie is the session.
- Depends on: #75, #7, #39.

---

## T29b - Frontend: session bootstrap and expiry (#80)

**User story:** US-4  
**Depends on:** #79

## Goal

The app knows on start whether a session exists and returns to the anonymous state when a session expires.
Source: `product-spec.md` section 4 US-4.

## Acceptance criteria

- [ ] On mount `SessionProvider` calls `GET /auth/session`; a 200 sets the user and CSRF token, a 401 leaves the user anonymous without an error banner, and `authReady` stays false until the call settles.
- [ ] A 401 from a protected call clears the session and shows a clear message; there is no automatic retry loop.
- [ ] The app shell and routes render the logged-in or anonymous state from the provider only; the mock supports the same state machine for unit tests.
- [ ] Unit tests cover the 200, 401, pending and expiry cases.

## Out of scope

- Register, login and logout forms: T29a.
- Protecting progress and history routes: T30a-T30b.

## Constraints

- A missing or expired session is not an error state.
- Depends on: #79.

---

## T30a - Frontend: lesson progress (#81)

**User story:** US-5  
**Depends on:** #80, #8, #41

## Goal

Completion state comes from the backend for the logged-in user and survives a reload.
Source: `product-spec.md` section 4 US-5.

## Acceptance criteria

- [ ] Logged-in topic and lesson pages read `GET /me/progress`; mark and unmark call `PUT` and `DELETE` with the CSRF token and update the counts; the state survives a page reload.
- [ ] An anonymous visitor sees no completion control and no history link.
- [ ] A 401 from a progress call sends the user to login and back to the same topic or lesson; the client handles a 403 refresh and retry; a 500 shows progress as unavailable without hiding the content.
- [ ] Unit tests cover loaded, empty, anonymous, 401 and 500 cases.

## Out of scope

- Quiz history: T30b.
- Editing progress outside the lesson and topic pages.

## Constraints

- The client sends no user id; the server derives the user from the session.
- Depends on: #80, #8, #41.

---

## T30b - Frontend: quiz history (#82)

**User story:** US-6  
**Depends on:** #80, #81, #9, #42

## Goal

The history list and attempt detail come from the backend for the logged-in user.
Source: `product-spec.md` section 4 US-6.

## Acceptance criteria

- [ ] The list uses `GET /me/attempts` newest-first from the server; the detail uses `GET /me/attempts/{attemptId}`; the stored snapshot wording is rendered.
- [ ] `topic_slug` is resolved to a title through `GET /topics`, with the slug as fallback when the topic is no longer published; the contract stays unchanged.
- [ ] A nullable `question_id` is handled (the list key uses position) and an empty `selected_option_text` shows "No answer selected".
- [ ] A 401 from a history call sends the user to login and back; 404, 500 and empty states use the existing components.
- [ ] Unit tests cover list, detail, empty, anonymous, 401, 404 and a null question id.

## Out of scope

- Editing or deleting attempts, and aggregation over attempts.
- Progress: T30a.

## Constraints

- The client sends no user id; the server derives the user from the session.
- Depends on: #80, #81, #9, #42.

---

## T32a - Containers: backend image (#83)

**User story:** none - supports all stories (`product-spec.md` section 7, operability)  
**Depends on:** #49, #50, #52

## Goal

The backend runs as a pinned container image that applies migrations and loads the seed on start.
Source: `product-spec.md` section 7 (operability), `README.md` section 8.

## Acceptance criteria

- [ ] `backend/Dockerfile` is multi-stage, pins the Go version from `go.mod`, builds with `go.sum` and uses no `latest` tag.
- [ ] The runtime image contains the API binary, the `migrations/` directory and the embedded seed; the working directory makes `os.DirFS("migrations")` resolve.
- [ ] The container reads `HOST`, `PORT`, `ENV`, `DATABASE_URL` and `SEED_*`; with a database it applies migrations, loads the seed and answers `GET /health` with 200.
- [ ] `docker build` succeeds from a clean checkout and the image runs as a non-root user.

## Out of scope

- The frontend image and the compose stack: T32b.
- Production hardening, TLS, replicas and backups.

## Constraints

- Pinned versions only; no secret in the image.
- Depends on: #49, #50, #52.

---

## T32b - Containers: frontend image and Compose stack (#84)

**User story:** none - supports all stories (`product-spec.md` section 7, operability)  
**Depends on:** #83, #76, #78, #80, #82, #49, #50

## Goal

`docker compose up --build` from a clean clone starts the database, the API and the frontend, with the seed content loaded.
Source: `product-spec.md` section 7 (operability), `README.md` section 8.

## Acceptance criteria

- [ ] `frontend/Dockerfile` pins the Bun and static-server versions, builds with `VITE_API_MODE=http` and `VITE_API_BASE_URL=/api`, serves the SPA and falls back to `index.html`.
- [ ] The static server proxies `/api/` to `http://$API_UPSTREAM` (`host:port`, default `backend:8080`), so the browser is same-origin; the session cookie keeps `SameSite=Lax` and no CORS is added.
- [ ] `docker-compose.yml` defines `db`, `backend` and `frontend`, with the pinned PostgreSQL image, health checks and start order; the backend waits for the database and the frontend for the backend.
- [ ] `docker compose up --build` from a clean clone gives a usable application with the seeded topics, lessons and questions; `docker compose down -v` followed by `up --build` works; the seed is safe on every start.
- [ ] `.env.example` documents every variable the containers read (`DATABASE_URL`, `POSTGRES_*`, `SEED_*`, `VITE_API_MODE`, `VITE_API_BASE_URL`, `API_UPSTREAM`) with local defaults and no real secret; the Bun version is pinned in `.bun-version`.

## Out of scope

- Production hardening, TLS termination, replicas and backups.
- Deployment: T35a and T34b.

## Constraints

- Pinned versions only; the lockfiles are used in the images.
- Depends on: #83, #76, #78, #80, #82, #49, #50.

---

## T31a - Frontend integration tests: harness, content and quiz (#85)

**User story:** US-1, US-2, US-3  
**Depends on:** #82, #84, #52

## Goal

The content and quiz screens are proven against the running backend and database, not only against the mock.
Source: `product-spec.md` section 7 (testing), `README.md` section 9.

## Acceptance criteria

- [ ] `bun run test` runs only the unit tests; `frontend/tests/integration/` is excluded, and `bun run test:integration` runs only that folder.
- [ ] The existing mock-based cross-screen tests are moved to the unit suite or rewritten, so the integration folder runs only against a real backend and a seeded test database.
- [ ] A documented command starts an isolated test database and the backend (the compose `db` service plus `go run ./cmd/api`, or the full compose stack), applies migrations and seed, and never points at the development database; the harness keeps a cookie jar for `HttpOnly` cookies and uses unique emails per run.
- [ ] Tests cover US-1 and US-2/US-3: topics, topic and lesson; five and fewer questions; no correctness leak before submission; server score; an anonymous result is not stored; a foreign question is rejected with 400.
- [ ] The suite starts from a clean database and passes twice in a row.

## Out of scope

- Accounts, progress and history tests: T31b.
- Browser automation and visual regression tests.

## Constraints

- Integration tests stay in `frontend/tests/integration/`, never mixed with unit tests.
- Depends on: #82, #84, #52.

---

## T31b - Frontend integration tests: accounts, progress and history (#86)

**User story:** US-4, US-5, US-6  
**Depends on:** #85, #81, #82

## Goal

Registration, the session, progress and quiz history are proven against the running backend and database.
Source: `product-spec.md` section 7 (testing), `README.md` section 9.

## Acceptance criteria

- [ ] Tests cover register, login, current session and logout, including a duplicate email and a CSRF refresh and retry.
- [ ] Progress mark and unmark survive a reload and are isolated between two users.
- [ ] A logged-in quiz submission stores an attempt; the history list and detail show the stored snapshot; another user receives 404.
- [ ] The suite starts from a clean database and passes twice in a row; the results are recorded in `docs/artifacts/integration/frontend-results.md` and linked from the issue.
- [ ] The integration command in README section 9 runs both integration tasks.

## Out of scope

- Browser automation and visual regression tests.
- Backend integration tests: already #52.

## Constraints

- Integration tests stay in `frontend/tests/integration/`, never mixed with unit tests.
- Depends on: #85, #81, #82.

---

## T33a - CI: unit tests and checks (#87)

**User story:** none - supports all stories (`product-spec.md` section 7, testing and operability)  
**Depends on:** #86, #84

## Goal

Every push and pull request runs the same checks and unit tests as `AGENTS.md` through a Makefile, so broken work is caught before it is merged and the same commands run locally and in CI.
Source: `product-spec.md` section 7, `README.md` section 9.

## Acceptance criteria

- [ ] A Makefile provides `make test` (all unit tests), `make test-integration` (backend and frontend integration tests on the test database; backend runs with `-p 1` because packages share one database), `make check` (gofmt, go vet, typecheck, lint, format:check, redocly lint), `make up`, and `make down`.
- [ ] A workflow runs on every push and pull request and calls these Makefile targets for the non-database checks and unit tests (at minimum `make check` and `make test`), so CI does not duplicate the command list.
- [ ] The Go version comes from `go.mod`, the Bun version from `.bun-version`, and the PostgreSQL major from the compose image; every action is referenced by a pinned version.
- [ ] A failing check fails the workflow, no step deploys anything, and the workflow is green on the default branch.
- [ ] README section 9 lists the Makefile targets CI runs.

## Out of scope

- The integration job: T33b (it calls `make test-integration`).
- Deployment: T35a and T34b.

## Constraints

- Do not run a deploy command or workflow.
- `-p 1` for backend integration tests is documented by this issue.
- Depends on: #86, #84.

---

## T33b - CI: integration tests against PostgreSQL (#88)

**User story:** none - supports all stories (`product-spec.md` section 7, testing and operability)  
**Depends on:** #87, #86, #84

## Goal

Every push and pull request runs the backend and frontend integration suites against a clean PostgreSQL database.
Source: `product-spec.md` section 7, `README.md` section 9.

## Acceptance criteria

- [ ] A separate job starts `postgres:16-alpine`, the same major as the compose file, creates `learn_ai_engineering_test`, and supplies `TEST_DATABASE_URL` and `DATABASE_URL` from the workflow only.
- [ ] The job calls `make test-integration` (defined by #87) against that test database; `-p 1` for the backend integration tests is documented by #87 because the packages share one database.
- [ ] The job starts from an empty database each run, applies migrations and seed, and never uses a shared or developer database.
- [ ] A failing test fails the job, no step deploys anything, and README section 9 documents `make test-integration`.

## Out of scope

- Deployment: T35a and T34b.
- Load and performance tests.

## Constraints

- Never point integration tests at the development database.
- Depends on: #87, #86, #84.

---

## T35a - Render: service setup and environment (#90)

**User story:** none - supports all stories (`product-spec.md` section 7, operability)  
**Depends on:** #88, #84

## Goal

The Render Postgres database, the backend and frontend services and their environment are set up in the Render dashboard and written down in `docs/deployment.md`, so deployment is repeatable and every later deploy is built by Render from the repository.
Source: `product-spec.md` section 7 (operability, reproducibility), section 10 (hosting target), `README.md` section 10.

## Acceptance criteria

- [ ] The maintainer created the Render Postgres database and the backend and frontend web services in the Render dashboard, linked both services to the `main` branch, and `docs/deployment.md` records the same setup: Docker runtime, `backend/` and `frontend/` as the build context, and the same region.
- [ ] The backend reads `DATABASE_URL` from the Render database, `ENV=production` and the `SEED_*` secrets from Render environment variables; the frontend reads `API_UPSTREAM` (backend host and port) and the `VITE_*` build values; no secret value is in the repository.
- [ ] The frontend is the public origin and proxies `/api/` to the backend, so the browser stays same-origin, the session cookie keeps `SameSite=Lax`, and the backend is reached over the private network.
- [ ] The backend health check path is `/health`, and both services have Auto-Deploy set to *After CI Checks Pass* (the exact Render option; see `https://render.com/docs/deploys#configuring-auto-deploys`).
- [ ] `docs/deployment.md` documents the rollback: redeploy an earlier commit from the Render dashboard.
- [ ] There is no `render.yaml`, no registry, no image push and no deploy hook; secrets live only in Render environment variables.
- [ ] This task does not verify a deployment; the CD gate and the first verification are T34b.

## Out of scope

- The CD gate and the first deployment verification: T34b.
- Monitoring, alerting, backups and multiple environments.

## Constraints

- The hosting target is Render, decided on 2026-11-07.
- Render builds both images from the repository Dockerfiles; no credential or real secret in the repository.
- Depends on: #88, #84.

---

## T34b - CD: deploy gate and first deployment verification (#91)

**User story:** none - supports all stories (`product-spec.md` section 7, operability)  
**Depends on:** #90, #88, #84

## Goal

A commit whose CI checks pass is deployed by Render and the deployed build is verified on the public URL, while a commit with failing CI is not deployed.
Source: `product-spec.md` section 7 (operability), `README.md` section 10.

## Acceptance criteria

- [ ] Both Render services keep Auto-Deploy set to *After CI Checks Pass*; a pull request never deploys, and Render does not deploy a commit when any check fails or when zero checks are detected (see `https://render.com/docs/deploys#integrating-with-ci`).
- [ ] Verified once on `main`: a commit with a failing CI check produced no Render deploy (the commit is reverted afterwards); the next commit with all CI checks passing was deployed.
- [ ] The deployed build is verified on the public URL: the page opens, topics and a quiz work, registration, login and history work, and the seed data is present (seven topics with at least one published lesson and six published questions each).
- [ ] `/health` answers 200 through the frontend origin, and the session cookie works over HTTPS.
- [ ] The public URL, the deployed commit and the result are in README section 10 and `docs/deployment.md`; `docs/deployment.md` keeps the rollback from T35a (redeploy an earlier commit).
- [ ] A failed check or a failed deploy is visible and stops the rollout; no secret or deploy credential is logged.
- [ ] The AI log records the checks, the commits and the results.

## Out of scope

- Monitoring, alerting, backups and multiple environments.
- Custom domain and TLS beyond the Render default.
- Automatic rollback.

## Constraints

- Do not run deploy commands unless explicitly asked; the deploy is Render's Auto-Deploy after the CI checks pass.
- No `render.yaml`, no registry, no image push and no deploy hook; secrets live only in Render environment variables.
- Depends on: #90, #88, #84.

---

## T39a - Observability stack: local Grafana profile (#109)

**User story:** none - supports all stories (`product-spec.md` section 7, operability)  
**Depends on:** #84

## Goal

A developer can start a local observability stack (collector, Prometheus, Loki, Tempo, Grafana) beside the application with one command, without changing the default stack and without deploying it.
Source: `product-spec.md` section 7 (operability), `README.md` section 8.

## Acceptance criteria

- [ ] `docker-compose.yml` defines an `observability` profile with pinned images for an OpenTelemetry Collector, Prometheus, Loki, Tempo and Grafana; `docker compose up --build` starts only `db`, `backend` and `frontend`, and `docker compose --profile observability up --build` adds the five observability services.
- [ ] The collector accepts OTLP over gRPC `4317` and HTTP `4318`; it exports metrics to Prometheus, logs to Loki and traces to Tempo, and the pipeline configuration is committed under `observability/`.
- [ ] Grafana starts provisioned with the three datasources (Prometheus, Loki, Tempo) from files, needs no manual setup, and its UI is reachable on a localhost-bound port; Prometheus, Loki, Tempo and the collector likewise bind only to localhost where they expose a port.
- [ ] A documented smoke check starts the profile and fails unless all five services are healthy (Grafana and the other HTTP APIs answer their health endpoints, and the collector's OTLP endpoints accept a connection).
- [ ] The profile uses local volumes only, contains no secret, and is not referenced by the Render setup or any CI/CD workflow; the stack README records the ports and the start and stop commands.
- [ ] `.env.example` documents every new variable and no real credential.

## Out of scope

- Backend instrumentation and the first journey: T39b.
- The dashboard: T39c.
- Alerts and the on-call receiver: T40 and T41a.
- Remote or hosted observability and retention beyond the local volumes.

## Constraints

- Pinned versions only; no `latest`.
- Local only: no deploy command, workflow or Render setting is changed.
- No secret, personal data or customer data in configuration or fixtures.
- Depends on: #84.

---

## T39b - Backend observability for the quiz journey (#110)

**User story:** US-2, US-3  
**Depends on:** #109

## Goal

Submitting a quiz produces metrics, one trace and structured logs that share the same `trace_id`, so the quiz journey can be followed in Grafana without recording what the visitor typed.
Source: `product-spec.md` US-2 (take a quiz), section 7 (operability).

## Acceptance criteria

- [ ] The Go backend uses the OpenTelemetry SDK and OTLP exporter, configured only by environment variables (`OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_SERVICE_NAME` and the standard `OTEL_*` names); with no endpoint the telemetry is a no-op and the API starts and behaves the same.
- [ ] The key journey is US-2: `GET /topics/{topicSlug}/quiz` and `POST /topics/{topicSlug}/quiz/submissions`; each request creates a server span with the route template and status, and the submission trace has child spans for loading the questions, grading and storing the attempt when a session is present.
- [ ] Metrics cover request count and duration by method, route template and status, plus quiz submissions; the request duration histogram records the journey `trace_id` as an exemplar, and the metric names and labels contain no user, session or content identifiers.
- [ ] Every request log is structured with `slog` (JSON to stdout) and carries `trace_id` and `span_id`; the quiz journey log lines carry the same `trace_id` as its trace, and the collector sends them to Loki.
- [ ] Telemetry contains no personal or customer data and no credentials: no request or response body, email, password, cookie, CSRF token, session id, answer text or free text is recorded; only route template, method, status, duration, counts and the trace/span ids.
- [ ] A test sends requests with unique email, password, cookie and body marker values and asserts that none of them appears in the exported spans, logs or metrics; another test asserts the quiz trace and its logs share one `trace_id`.
- [ ] The new dependency is pinned in `go.mod` and `go.sum`, and the journey is visible in Grafana (trace in Tempo, logs in Loki, metrics in Prometheus) when the `observability` profile runs.

## Out of scope

- The dashboard: T39c.
- Alerting and the on-call receiver: T40 and T41a.
- Frontend telemetry, database and host metrics, and instrumenting other journeys.
- Production telemetry or a hosted backend.

## Constraints

- Never log request bodies, cookies, passwords, emails, CSRF tokens or session ids.
- Telemetry is disabled by default and never blocks startup or the request when the collector is absent.
- Pinned versions only; no secret in the repository.
- Depends on: #109.

---

## T39c - Grafana dashboard for the quiz journey (#111)

**User story:** US-2  
**Depends on:** #110

## Goal

The quiz journey has one provisioned Grafana dashboard that shows its traffic, errors, latency, traces and logs, so a failure can be followed from the user impact to the trace.
Source: `product-spec.md` US-2, section 7 (operability).

## Acceptance criteria

- [ ] The dashboard JSON is committed under `observability/grafana/dashboards/` and provisioned by Grafana on start; no panel is created or edited by hand.
- [ ] It shows, for `GET /topics/{topicSlug}/quiz` and `POST /topics/{topicSlug}/quiz/submissions`: request rate, 5xx rate and p95 latency from Prometheus, plus quiz submissions and a trace search from Tempo.
- [ ] A logs panel from Loki lists the journey logs and shows the `trace_id`; a trace id from the logs or the trace panel opens the matching trace in Tempo, and a trace links back to its logs (the correlation is in the provisioned datasource files).
- [ ] The dashboard can be filtered by journey and time range; panels show data after the smoke traffic and show "no data" instead of an error when the profile is empty.
- [ ] A check queries the Grafana API after start and fails unless the dashboard UID exists, contains the required panels and has no provisioning error.
- [ ] The stack README records the dashboard URL and UID.

## Out of scope

- Alerts: T40.
- Other journeys and SLO reporting.
- Screenshots committed as evidence (the JSON is the artifact).

## Constraints

- Provisioned from files only; no manual dashboard state.
- No personal data or credentials in panel queries, annotations or fixtures.
- Depends on: #110.

---

## T40 - Alert on user impact: quiz submission 5xx rate (#112)

**User story:** US-2  
**Depends on:** #111

## Goal

A sustained 5xx rate on the quiz submission endpoint raises one alert that is delivered to a webhook, so user-visible failure is signalled without paging on internal host metrics.
Source: `product-spec.md` US-2, section 7 (operability).

## Acceptance criteria

- [ ] One alert rule fires when the 5xx rate of `POST /topics/{topicSlug}/quiz/submissions` is above 5% for 2 minutes with a minimum of requests; it resolves when the rate falls back, and it does not fire on 4xx responses, on other endpoints or when there is no traffic.
- [ ] The rule is provisioned from a committed file and its name and summary state the user impact (visitors cannot submit a quiz); the threshold, window and minimum traffic are documented.
- [ ] The webhook URL comes from an environment variable; the notification is sent as JSON and a test listener or test stub (not committed) receives it; no secret is committed.
- [ ] A check produces 5xx responses on the endpoint and observes the webhook delivery after the window; a run with only 2xx/4xx responses sends nothing.
- [ ] The alert is local to the `observability` profile and is not deployed.

## Out of scope

- Storing incident data and waking the on-call agent: T41a.
- Multiple alerts, paging providers, silences and escalation policies.
- Alerts on host, container or database internals.

## Constraints

- User impact only: the 5xx rate of the key endpoint, not CPU, memory or uptime.
- Local only; no deploy and no secret in the repository.
- Depends on: #111.

---

## T41a - Alert receiver and on-call agent (#113)

**User story:** US-2  
**Depends on:** #112, #87

## Goal

When the quiz alert fires, a local receiver stores the incident's logs and traces and starts an on-call agent through `opencode run`, using a model provider and key that come only from environment variables.
Source: `product-spec.md` US-2, section 7 (operability), `AGENTS.md` (on-call role).

## Acceptance criteria

- [ ] A receiver service is part of the `observability` profile, with a pinned image, a health endpoint and a webhook endpoint; the profile points the alert's webhook URL to it by default.
- [ ] On an alert it resolves the incident window from the payload, fetches the matching logs from Loki and traces from Tempo through their APIs, and writes an incident bundle (alert payload, metadata, logs, traces) under a local volume recorded in the bundle; it never writes credentials or personal data.
- [ ] It starts `opencode run` in the repository checkout as the on-call role, passing the bundle path and a prompt built from the alert and the stored evidence; the model provider, model and API key come from environment variables (`OPENCODE_PROVIDER`, `OPENCODE_MODEL`, `OPENCODE_API_KEY` or the provider's documented equivalents) and are documented in `.env.example`; when the key is missing it fails fast with a clear message and starts no agent.
- [ ] The provider, model and key are read from environment variables; the key never appears in the repository, the incident bundle, logs or the agent prompt, and only the provider and model names may be recorded as metadata.
- [ ] `AGENTS.md` gains an **On-call agent** role and rule: it runs only in the local observability drill, works on a temporary branch, never deploys, never prints or commits secrets or personal data, and leaves its diagnosis and fix as a commit or pull request on that branch.
- [ ] A test posts a sample alert with stub Loki and Tempo responses and a fake `opencode` executable; it asserts that the bundle contains the expected logs and traces, that the stub is called with the incident context, and that a sentinel API key passed in the environment is absent from the bundle, the logs and the generated prompt.

## Out of scope

- The real-bug drill and its evidence: T41b.
- Production paging, on-call rotation, automatic merge, deploy or rollback.
- Multiple alerts and richer correlation.

## Constraints

- No deploy command or workflow; local profile only.
- The provider and key come from environment variables only; never from the repository, prompt, logs or fixtures.
- No personal or customer data and no request bodies in the bundle.
- Depends on: #112, #87.

---

## T41b - On-call drill with a real bug (#114)

**User story:** US-2  
**Depends on:** #113

## Goal

A real bug on a temporary branch produces the alert, the receiver stores the incident and the on-call agent fixes the bug on that branch, with the whole run recorded as evidence.
Source: `product-spec.md` US-2, section 7 (operability), `AGENTS.md` (on-call role).

## Acceptance criteria

- [ ] A temporary branch (for example `drill/on-call-<date>`) carries a real bug that makes `POST /topics/{topicSlug}/quiz/submissions` return 5xx for a normal quiz submission; nothing in the alert, receiver or agent path is mocked.
- [ ] With the `observability` profile running and the journey exercised, the alert fires, the receiver stores the incident bundle, and it starts `opencode run`; the agent diagnoses the bug and fixes it with a commit on that branch, or a pull request from that branch.
- [ ] After the fix and a restart of the affected service, the drill submission succeeds, the alert resolves, and the existing tests pass on the branch.
- [ ] `docs/artifacts/incidents/<incident-id>/` contains the alert payload, the stored logs and traces, the agent's diagnosis and fix output with no secrets or personal data, and a `README.md` that links the commit or pull request and summarises the timeline.
- [ ] The branch is not merged; it is deleted or left for review after the evidence is committed, and nothing is deployed.
- [ ] The AI log entry records the drill, the commands and the result.

## Out of scope

- Production incidents, real paging and on-call rotation.
- Automatic merge, deploy or rollback.
- Improving the agent or the alerts beyond what the drill proves.

## Constraints

- Never deploy and never point the drill at production or real user data.
- No secret, credential or personal data in the branch, the evidence, the logs or the prompt.
- Depends on: #113.

---

## T36 - README is runnable from a clean machine (#93)

**User story:** none - supports all stories (`product-spec.md` section 7, reproducibility)  
**Depends on:** #91, #86, #84, #88, #109, #110, #111, #112, #113, #114

## Goal

A person who has never seen the repository can clone it and run, test and deploy the project by following the README alone.
Source: `product-spec.md` section 7 (reproducibility), `README.md` section 11.

## Acceptance criteria

- [ ] Every command in README sections 8, 9, 10 and 11 is executed from a clean clone in this task and works exactly as written; any correction is a change to the README. README uses the Makefile targets from #87 (`make test`, `make test-integration`, `make check`, `make up`, `make down`) where they apply.
- [ ] The status table matches reality: finished items no longer say _planned_, and unfinished ones say what is missing.
- [ ] The repository structure matches the real tree, including the Dockerfiles, `docker-compose.yml`, every variable in `.env.example`, the Makefile, and the integration test command.
- [ ] README documents the local observability stack: the `observability` Compose profile, its ports, the Grafana dashboard, the alert and receiver, the on-call drill and the `OTEL_*` and `OPENCODE_*` variables, with links to `observability/README.md` and `docs/artifacts/incidents/`.
- [ ] The status table marks observability as local only and not deployed, and the repository structure includes `observability/` and `docs/artifacts/incidents/`.
- [ ] The links to the repository, the issue tracker and the live application are filled in; the deployment section names Render and the public URL.
- [ ] The AI workflow entry records which commands were run from the clean clone and their result.

## Out of scope

- Translations, screenshots and videos.
- New features.

## Constraints

- No command is documented that was not run.
- Depends on: #91, #86, #84, #88, #109, #110, #111, #112, #113, #114.

---

## T37 - Complete the AI workflow log and the stage artifacts (#94)

**User story:** none - supports all stories (`AGENTS.md` "Conventions", `product-spec.md` section 7)  
**Depends on:** #93, #109, #110, #111, #112, #113, #114

## Goal

Every delivered issue has its entry in the AI workflow log and its stage artifact committed and linked, so the AI-assisted work can be reviewed end to end.
Source: `AGENTS.md` "Conventions" (log notable AI-assisted steps), `README.md` section 4.

## Acceptance criteria

- [ ] `docs/ai-log/` has one entry file per closed issue from the backend stage onward, following the template in `docs/ai-workflow.md`: tool, prompt, context given, result, artifacts and where they were used, what was changed by hand, verification, links.
- [ ] Every stage artifact in `docs/artifacts/` is linked from its log entry and from its issue.
- [ ] The entries for T39a-T41b link the dashboard JSON, the alert rule, the receiver, the drill evidence in `docs/artifacts/incidents/` and the on-call role in `AGENTS.md`; no entry contains a secret, credential or personal data.
- [ ] No secret, credential or personal data appears in the log or in any artifact; the sources of the skills installed in T02 are recorded.
- [ ] `docs/process.md` and README section 4 describe the workflow as it was actually followed, including anything that changed.
- [ ] The log entries reference the commits and issues they belong to.

## Out of scope

- New documentation that is not about the work that was done.

## Constraints

- Never record secrets or private data.
- Depends on: #93, #109, #110, #111, #112, #113, #114.

## Assumptions

Status on 2026-11-07: assumptions 1-6 are implemented by the merged backend (#34-#52); they stay here as the record the backend followed.

1. **`/health` in the contract.** The specification does not describe a health endpoint, but the containers and the deployment tasks (T32, T35a and T34b) need one. It is added to the contract in T10.
2. **Addressing.** Topics are addressed by slug (the data model has one); lessons are addressed by id (the data model has no lesson slug).
3. **Stateless quizzes.** No quiz-set table is added: the submission carries the question and chosen option ids, and the server checks that the questions belong to the topic. This follows the acceptance criterion that a foreign question is rejected.
4. **CSRF.** Proposed for the contract in T11: double-submit token, issued at registration and log in, sent in a header on state-changing requests that carry a session cookie. The specification only requires the protection, not the mechanism.
5. **Test runner for the frontend.** `bun run test` must run a runner other than Bun's built-in one (Vitest), as stated in `AGENTS.md`.
6. **In-memory repositories in the backend stage.** T14-T20 are unit tested with in-memory implementations behind the repository interfaces; they are replaced by PostgreSQL implementations in T21-T22.
7. **Extra course deliverables are not in this list.** The agent extension pack (`agent-capabilities/`, `agent-hooks/`, `mcp-server/`), `security/` and `ops/` are course modules that `AGENTS.md` and README section 12 mention; the instruction for this breakdown listed CI/CD, deployment, README and docs, and the AI workflow log only. They are not broken into tasks here.

## Decisions

Answered by the maintainer on 2026-10-17; items 4-5 and 7-8 were updated on 2026-11-07 and item 5 on 2026-11-10; they are binding for the tasks above.

1. **Passwords.** Registration accepts at least 8 and at most 128 characters; passwords are hashed with Argon2id. The length rule goes into the contract in T11; the algorithm is chosen in T17.
2. **Session cookie `SameSite`.** `Lax`, so normal navigation back to the site keeps working.
3. **Seed content.** Drafted with AI assistance, reviewed by hand and committed as seed files. One lesson and six questions per topic is the minimum, so T24 stays one session.
4. **Frontend hosting.** The frontend is served by its own static-server image, which also proxies `/api/` to the backend so the browser stays same-origin. The API base URL is `/api`; no CORS is added.
5. **Hosting target.** Render, decided on 2026-11-07. Render builds the backend and frontend images from the repository Dockerfiles. The maintainer creates the Render Postgres database and the two web services in the Render dashboard and links the `main` branch; T35a records the setup. T34b verifies that Render deploys a commit only after the CI checks pass. There is no `render.yaml`, no image registry or push, and no deploy hook.
6. **Task count.** The list stays as it is, with T27-T35 split into a/b tasks except the cancelled T34a and T35b, and T36-T37 unchanged, so the list still ends at T37.
7. **Integration commands.** Backend integration tests run as `(cd backend && go test -p 1 -tags=integration ./...)`; `-p 1` is required because the packages share one test database. Frontend integration tests run as `(cd frontend && bun run test:integration)`.
8. **Same-origin deployment.** The frontend image proxies `/api/` to the backend in Compose and on Render, so session cookies keep `SameSite=Lax` and the backend needs no CORS configuration.
