# Backlog: tasks from the product specification

Derived from [`product-spec.md`](../../../product-spec.md) (draft v0.2) on 2026-10-17, following [`docs/task-template.md`](../../task-template.md) and the order in [`AGENTS.md`](../../../AGENTS.md).

- `T02`-`T13` became GitHub issues on 2026-10-17, where the issue number matches the task number (`T02` is #2, `T13` is #13); in those issues `Depends on: Txx` is written as `Depends on: #<issue number>`.
- `T01` was issue #1. It was cancelled by the maintainer on 2026-10-18 and removed from this file: the scaffold it asked for contradicts the project order of `AGENTS.md`, the frontend project comes from Bolt and the Go project belongs to `T14`, after the contract. Nothing of it was carried over to another task.
- `T14`-`T37` are still placeholders for issue numbers. When the maintainer creates them, every `Depends on: Txx` becomes `Depends on: #<issue number>`.
- `T02`-`T12` are done and merged into `main`; their text is kept unchanged in [`done.md`](done.md). The issues hold the same text, so open tasks are read here and finished ones there.
- Each task fits one session and can be read on its own.
- This file creates no issues and no code. The maintainer merges, splits and removes tasks before creating them (see [`docs/process.md`](../../process.md), "Backlog"). The issues for T02-T13 were created from this file on 2026-10-17.
- `T04`-`T09` were revised on 2026-10-19, after the Bolt prototype was imported and the API client of issue #3 was merged: each of them now adapts and verifies the pages that already exist in `frontend/` instead of building them, carries the lint problems of the imported files it touches, and `T09` closes the prototype stage with `bun run lint` exiting 0. Issues #4-#9 were updated in the same change. Nothing was merged, split or renumbered.
- No task adds behaviour outside version 1 of the specification.
- Every task ends with an entry in [`docs/ai-workflow.md`](../../ai-workflow.md) and a committed stage artifact; the retrospective pass over that bookkeeping is T37.

## Order

| Stage | Tasks | User stories |
|---|---|---|
| Frontend prototype with mock API client | T02-T09 | US-1, US-2, US-3, US-4, US-5, US-6 |
| API contract | T10-T13 | US-1, US-2, US-3 / US-4 / US-5, US-6 / US-7 |
| Backend by user story | T14-T20 | US-1, US-2, US-3, US-4, US-5, US-6, US-7 |
| Persistence: migrations, repositories, seed | T21-T25 | US-1, US-4, US-5, US-6, US-7, US-8 |
| Backend integration tests | T26 | US-1 ... US-6 |
| Connect frontend to backend | T27-T30 | US-1, US-2, US-3, US-4, US-5, US-6 |
| Frontend integration tests | T31 | US-1 ... US-6 |
| Containers | T32 | none (supports all) |
| CI | T33 | none (supports all) |
| CD | T34-T35 | none (supports all) |
| Documentation | T36-T37 | none (supports all) |

Tasks not covered by a user story come from section 7 (non-functional requirements) or from course deliverables the specification does not describe; each one names its source.

---

## T13 - Contract: maintainer draft and publish operations

**User story:** US-7 (content preparation)
**Depends on:** T10

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
- Depends on: T10.

---

## T14 - Backend foundation: server, configuration and error mapping

**User story:** none - supports US-1 to US-7
**Depends on:** T10, T11, T12, T13

## Goal

The Go service starts, reads its configuration from the environment, and maps domain errors to the error shape of the contract, so later backend tasks only add handlers and services.
Source: `product-spec.md` section 7 (security, configuration), `AGENTS.md` "Conventions".

## Acceptance criteria

- [ ] `backend/cmd/api` starts an HTTP server on the address from the environment and answers `GET /health` with 200.
- [ ] Configuration is read from environment variables with development defaults; a missing required variable stops the server with a clear message, and every variable is documented in `.env.example`.
- [ ] One domain error type maps to 400, 401, 403, 404, 409 and 500 and always writes the error object defined in the contract.
- [ ] Requests are logged with method, path, status and duration, without request bodies, cookies, passwords or hashes.
- [ ] Handlers are thin and the repository interfaces of the layers exist next to the services, with an in-memory implementation used by unit tests.
- [ ] Tests for the new behaviour exist and pass: the health endpoint, each status mapping, and a handler that returns a service error.

## Out of scope

- Any domain endpoint other than `/health`.
- The database, migrations and the real repositories (T21-T22).

## Constraints

- SQL only in the repository layer; logic in services.
- Depends on: T10, T11, T12, T13.

---

## T15 - Backend: public content endpoints

**User story:** US-1
**Depends on:** T14

## Goal

The API serves published topics, their lessons in order, and a single lesson, so the content screens have something real to call.
Source: `product-spec.md` section 4 US-1.

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

- Handlers stay thin; responses match `openapi.yaml` exactly.
- Depends on: T14.

---

## T16 - Backend: start a quiz and grade a submission

**User story:** US-2, US-3
**Depends on:** T14

## Goal

The API picks 5 random published questions for a topic and grades a submission on the server, so the browser never decides the score.
Source: `product-spec.md` section 4 US-2, US-3.

## Acceptance criteria

- [ ] Starting a quiz returns exactly 5 published questions of that topic, or all of them when the topic has fewer; two calls can return different question sets.
- [ ] The start response contains no field that reveals the correct option or the explanation.
- [ ] Grading is done server-side: an unanswered question is incorrect, an option that does not belong to the question is incorrect, and the submission still succeeds.
- [ ] A submission containing a question that does not belong to the topic is rejected with a 400 validation error and nothing is graded.
- [ ] The result contains the score, the total and, per question, the correctness, the correct option text and the explanation.
- [ ] A submission without a session is graded and answered normally.
- [ ] Unit tests exist and pass for the count, the random selection, both incorrect cases, the rejected submission and the result shape.

## Out of scope

- Storing attempts (T19).
- Question or option selection by the client (quiz size is fixed).

## Constraints

- A quiz is stateless: the submission carries the question and chosen option ids and the server checks they belong to the topic.
- Depends on: T14.

---

## T17 - Backend: registration, log in, log out and session

**User story:** US-4
**Depends on:** T14

## Goal

A visitor can create an account, log in with a session cookie, log out, and is recognised as anonymous afterwards, so progress and history can be protected.
Source: `product-spec.md` section 4 US-4, section 7 (security).

## Acceptance criteria

- [ ] Register validates the input and rejects a duplicate email with 409; the password is stored only as a salted hash, and no plain password is kept anywhere.
- [ ] Log in verifies the hash, creates a session that expires after 7 days, and sets the session cookie with the attributes defined in the contract; a wrong password returns an authentication error.
- [ ] Log out deletes the session and clears the cookie.
- [ ] A missing or expired session is treated as not logged in: the current-session endpoint answers 401 and the state-changing endpoints do the same.
- [ ] State-changing requests that carry a session cookie are refused without the CSRF token defined in the contract; a request without a session cookie is not checked for a token, because there is no session to protect.
- [ ] Passwords, hashes and session tokens never appear in a response or in a log.
- [ ] Unit tests exist and pass for register, duplicate email, log in, wrong password, log out, an expired session and a missing CSRF token.

## Out of scope

- Password reset, email verification and roles other than `is_maintainer` (T20).
- Rate limiting and password policy beyond the contract: see Decisions.

## Constraints

- Session state comes from the cookie, never from a request body or query parameter.
- Depends on: T14.

---

## T18 - Backend: lesson progress

**User story:** US-5
**Depends on:** T17

## Goal

A logged-in user can mark a lesson complete and undo it, and read their own completion state, so the screens can show progress.
Source: `product-spec.md` section 4 US-5.

## Acceptance criteria

- [ ] Marking a lesson complete records the completion time and is idempotent; undoing removes it and is safe to repeat.
- [ ] Reading progress returns only the current user's completed lessons.
- [ ] Every one of the three endpoints answers 401 without a valid session.
- [ ] No request carries a user id, so one user cannot change or read another user's progress.
- [ ] Unit tests exist and pass for mark, repeated mark, undo, read and the unauthenticated case.

## Out of scope

- Progress statistics and streaks: not in the specification.

## Constraints

- Ownership check in the protected path, not only in the client.
- Depends on: T17.

---

## T19 - Backend: quiz attempt history

**User story:** US-6, US-3
**Depends on:** T16, T17

## Goal

Every submission from a logged-in user is stored with a snapshot of the questions and answers, so past attempts stay readable and cannot be changed by later content edits.
Source: `product-spec.md` section 4 US-6, US-3.

## Acceptance criteria

- [ ] A submission with a valid session stores one attempt with topic, time, score and total, and one answer row per question.
- [ ] Every submission creates a new attempt; an earlier attempt is never overwritten.
- [ ] Each stored answer keeps the question text, the selected option text (empty when unanswered) and the correct option text as they were at the time.
- [ ] A submission without a session stores nothing and is answered normally.
- [ ] Listing attempts returns only the current user's attempts, newest first; reading an attempt that belongs to another user answers 404.
- [ ] Unit tests exist and pass for storage per submission, the snapshot wording, the anonymous case and the ownership case.

## Out of scope

- Deleting attempts and exporting them.

## Constraints

- No attempt is ever edited after it is written.
- Depends on: T16, T17.

---

## T20 - Backend: maintainer draft and publish operations

**User story:** US-7
**Depends on:** T15, T16, T17

## Goal

The maintainer can create and edit drafts and publish them explicitly, while other users cannot, so AI-assisted content is reviewed before anybody sees it.
Source: `product-spec.md` section 4 US-7, section 6.

## Acceptance criteria

- [ ] A session without `is_maintainer` answers 403 and a missing session answers 401 on every draft and publish operation.
- [ ] Creating and editing a lesson or a question stores it as a draft and never as published.
- [ ] A question with fewer than 2 or more than 6 options, or without exactly one correct option, is rejected with a validation error.
- [ ] Publishing sets the status to published and makes the content visible through the public endpoints of T15 and T16; unpublishing hides it again.
- [ ] Unit tests exist and pass for the authorisation cases, the validation case, the draft visibility before publication and the visibility after publish and unpublish.

## Out of scope

- An admin UI, and any way for AI tooling to publish.
- Bulk operations and content import.

## Constraints

- Publication is always a separate explicit action; nothing publishes as a side effect.
- Depends on: T15, T16, T17.

---

## T21 - Persistence: content tables and repositories

**User story:** US-1 (also serves US-2 and US-7)
**Depends on:** T15, T16, T20

## Goal

Topics, lessons, questions and answer options live in PostgreSQL through numbered migrations, and the public and maintainer endpoints read and write them.
Source: `product-spec.md` section 9 (data model), section 7 (persistence).

## Acceptance criteria

- [ ] Numbered migration files create `topics`, `lessons`, `questions` and `answer_options` with the fields of section 9; the topic slug is unique, positions are unique per parent, and every table has a status.
- [ ] Migrations are applied in order and recorded, and applying them to an empty database twice leaves the same schema.
- [ ] PostgreSQL implementations replace the in-memory content repositories behind the same interfaces, and the service layer is not changed to make them fit.
- [ ] An integration test (build tag `integration`) proves the published filter and the position ordering against a test database.
- [ ] `DATABASE_URL` and the test database variable are documented in `.env.example`, with separate values for development and test.
- [ ] The existing unit tests still pass unchanged; the integration test passes against `docker compose up -d db`.

## Out of scope

- Users, sessions, progress and attempts (T22).
- Down migrations: migrations only move forward, and destructive ones need the maintainer's approval.

## Constraints

- Schema changes go through numbered migration files, never by editing the schema by hand.
- Integration tests never point at the development database.
- Depends on: T15, T16, T20.

---

## T22 - Persistence: users, sessions, progress and quiz attempt tables

**User story:** US-4, US-5, US-6
**Depends on:** T17, T18, T19, T21

## Goal

Accounts, sessions, lesson completion and quiz attempts survive a restart, so a user stays logged in, keeps their progress, and a past attempt still reads exactly as it was graded.
Source: `product-spec.md` section 9, section 4 US-4, US-5 and US-6.

## Acceptance criteria

- [ ] Migrations create `users` (unique email, password hash, `is_maintainer` default false, created at) and `sessions` (id, user id, created at, expires at, with indexes on user id and expires at).
- [ ] Migrations create `lesson_progress` with the user and lesson as key and a completion timestamp.
- [ ] Migrations create `quiz_attempts` (user, topic, score, total, created at) and `quiz_attempt_answers` (attempt, nullable question reference, question text, selected option text, correct option text, correctness).
- [ ] The migrations are numbered in that order, so `quiz_attempts` comes after `users` because it references it.
- [ ] The repositories behind registration, log in, log out, progress, attempt storage and attempt reading use PostgreSQL behind the existing interfaces.
- [ ] An integration test registers, logs in, finds the session, marks a lesson, restarts the connection and reads the state back; an expired session is refused after a restart.
- [ ] An integration test stores two attempts for one user, then changes the question text in the database and reads both attempts again: the stored snapshot text is unchanged and the new attempt uses the new text.
- [ ] Deleting a question does not break reading an old attempt.
- [ ] The password hash is never selected into a response type or a log line.

## Out of scope

- Progress statistics, statistics over attempts and password reset.

## Constraints

- No migration drops or renames an existing table without asking the maintainer first.
- The snapshot columns are written once and never updated.
- Depends on: T17, T18, T19, T21.

---

## T23 - Seed mechanism and content for the first topic

**User story:** US-8 (initial content)
**Depends on:** T22

## Goal

Seed content is loaded from versioned files on first start and is safe to load again, and the first topic ships with a real lesson and a real question bank.
Source: `product-spec.md` section 4 US-8.

## Acceptance criteria

- [ ] Seed content lives in versioned files under `backend/`, so it can be reviewed in a diff; no content is generated at runtime.
- [ ] A documented command and the backend start both load the seed; the loader inserts what is missing and reports what it skipped.
- [ ] Loading the seed twice on the same database creates no duplicate topic, lesson, question or option, and the second run reports nothing created.
- [ ] The first topic (LLMs) has at least 1 published lesson and at least 6 published questions, each with 2 to 6 options, exactly one correct option and an explanation.
- [ ] An integration test loads the seed into an empty test database, loads it again, and asserts the counts did not change.

## Out of scope

- The other six topics (T24).
- AI-generated curriculum: a non-goal.

## Constraints

- Content is authored text in the repository, never produced by a tool at runtime.
- Passwords and secrets never go into seed files.
- Depends on: T22.

---

## T24 - Seed content for the remaining six topics

**User story:** US-8
**Depends on:** T23

## Goal

Every initial topic named in the specification has real learning material on first start, so the application is useful immediately.
Source: `product-spec.md` section 4 US-8, section 3 (initial topics).

## Acceptance criteria

- [ ] AI agents, MCP, RAG, embeddings, evaluation and deployment each have at least 1 published lesson and at least 6 published questions in the seed files.
- [ ] Every question has 2 to 6 options, exactly one correct option and a short explanation.
- [ ] The loader is unchanged; a fresh database gets all seven topics, and a second run still creates nothing.
- [ ] An integration test asserts the minimum count for each of the seven topics after one load.

## Out of scope

- More content than the minimum, translations, images and videos.

## Constraints

- Only published content goes into the seed; drafts are created through the maintainer operations (T20, T25).
- Depends on: T23.

---

## T25 - Maintainer account and the publish flow end to end

**User story:** US-7
**Depends on:** T20, T24

## Goal

A maintainer account exists after seeding, and the whole draft-to-published path works against the real database.
Source: `product-spec.md` section 4 US-7, section 3 (content loaded by the maintainer through the API or seed data).

## Acceptance criteria

- [ ] The seed creates a user with `is_maintainer` true, taking its credentials from environment variables; no password is committed.
- [ ] An integration test logs in as the maintainer, creates a draft lesson and a draft question, sees neither through the public endpoints, publishes them, sees both, unpublishes them and sees neither.
- [ ] An integration test with a normal registered user gets 403 on the draft and publish operations.
- [ ] An integration test with no session gets 401 on the same operations.
- [ ] The seeded maintainer account is documented in `.env.example` with the variables it needs.

## Out of scope

- An admin interface, roles beyond the flag, and any AI tool that publishes.

## Constraints

- The maintainer flag is set in the database or the seed only.
- Depends on: T20, T24.

---

## T26 - Backend integration tests for the key workflows

**User story:** US-1, US-2, US-3, US-4, US-5, US-6
**Depends on:** T25

## Goal

The whole API is exercised end to end against a real database, so the workflows in README section 9 are proven, not assumed.
Source: `product-spec.md` section 7 (testing), `README.md` section 9.

## Acceptance criteria

- [ ] Integration tests (build tag `integration`) drive the HTTP API against the test database for: browsing published content, starting and submitting a quiz, registering, logging in and logging out, saving progress, and reading quiz history.
- [ ] They also assert that unpublished content never appears in a public response and that a submission with a foreign question is rejected.
- [ ] The suite creates and cleans its own data and can run twice in a row against the same test database.
- [ ] The documented command `(cd backend && go test -tags=integration ./...)` runs only these tests, and unit tests still run without a database.
- [ ] The results are recorded in the stage artifact `docs/artifacts/integration/backend-results.md` and linked from the issue.

## Out of scope

- Repository-level SQL tests, which already live with the repositories (T21-T22).
- Load and performance tests.

## Constraints

- Integration tests stay separate from unit tests and never point at the development database.
- Depends on: T25.

---

## T27 - Connect the frontend: content screens

**User story:** US-1
**Depends on:** T04, T05, T26

## Goal

The topics, topic and lesson screens read from the running backend instead of the mock, so the prototype becomes the real application.
Source: `product-spec.md` section 7 (one client module, contract), section 4 US-1.

## Acceptance criteria

- [ ] The API client has a real HTTP implementation selected by one environment variable; the components are not changed to make it work.
- [ ] The three screens load their data from the backend, and a run against the running stack shows the seeded topics and lessons.
- [ ] HTTP failures, a 404 and a 500 map to the error and not-found states built in T04 and T05.
- [ ] Unit tests for the client's status-to-error mapping pass, and the component tests still pass against a fake client.
- [ ] No component contains a direct `fetch`.
- [ ] The topics are addressed by slug, as the contract of T10 decides, so the client and the screens are reconciled with it: the prototype addressed the topics by id (`api.getTopic(topicId)`).
- [ ] The four fields of the prototype that the data model of `product-spec.md` section 9 does not have (`icon`, `color`, `summary`, `durationMin`) are resolved: either the screens derive them or the data model changes. The contract of T10 does not carry them.

## Out of scope

- Quiz, accounts, progress and history (T28-T30).
- Changing the visual design.

## Constraints

- All backend calls go through `frontend/src/api/`; the base URL comes from the environment.
- Depends on: T04, T05, T26.

---

## T28 - Connect the frontend: quiz

**User story:** US-2, US-3
**Depends on:** T06, T16, T26

## Goal

The quiz flow calls the backend for the questions and for the grading, so the score the user sees is the server's score.
Source: `product-spec.md` section 4 US-2, US-3.

## Acceptance criteria

- [ ] Starting a quiz calls the backend and renders exactly what it returns, including fewer than 5 questions when the topic has fewer.
- [ ] The browser sends the question and chosen option ids and renders the returned score, correctness, correct answer and explanation; it computes none of them.
- [ ] A rejected submission (a question from another topic) is shown as a validation error, and an unanswered question is submitted as unanswered and marked incorrect.
- [ ] The CSRF token is sent with the submission when a session cookie is present; an anonymous submission sends no token and is still answered.
- [ ] The correct answer is never rendered before the submission response arrives.
- [ ] The topic of the quiz is addressed by slug, as the contract of T10 decides, so `startQuiz` and `submitQuiz` of the client are reconciled with it: the prototype used the topic id.
- [ ] The result screen of T06 renders the question text and the selected answer text as well; the submit response of the contract of T10 carries the question id, the correctness, the text of the correct option and the explanation (US-2), so it is decided here whether the client maps the two missing texts or the screen drops them.
- [ ] Unit tests for the client mapping and the flow pass.

## Out of scope

- Storing attempts in the UI and any history view (T30).

## Constraints

- The quiz size is not configurable in the UI.
- Depends on: T06, T16, T26.

---

## T29 - Connect the frontend: accounts and session

**User story:** US-4
**Depends on:** T07, T17, T26

## Goal

Registration, log in and log out work against the backend with a cookie session, and the app knows on start whether a session exists.
Source: `product-spec.md` section 4 US-4, section 7 (security).

## Acceptance criteria

- [ ] Registration, log in and log out call the backend; the duplicate email and validation errors of the contract are shown in the form.
- [ ] The session cookie is sent by the browser and the CSRF token of the contract is attached to every state-changing request sent with a session cookie; no session token or password is written to `localStorage` or `sessionStorage`.
- [ ] On start the app calls the current-session endpoint and shows the logged-in or anonymous state; a 401 shows the anonymous state without an error banner.
- [ ] A session that expired while the app was open returns the UI to the anonymous state with a clear message.
- [ ] Unit tests pass for the session state machine and the error mapping.

## Out of scope

- "Remember me", multiple devices and social login.

## Constraints

- No token in browser storage; the cookie is the session.
- Depends on: T07, T17, T26.

---

## T30 - Connect the frontend: progress and quiz history

**User story:** US-5, US-6
**Depends on:** T08, T09, T18, T19, T26

## Goal

Completion state and quiz history come from the backend for the logged-in user, so the screens show the user's own data.
Source: `product-spec.md` section 4 US-5, US-6.

## Acceptance criteria

- [ ] The completion control and the completion counts on the topic page come from the backend and survive a page reload.
- [ ] An anonymous visitor sees no completion control and no history link.
- [ ] The history list and the attempt detail come from the backend, newest first, showing the stored snapshot wording.
- [ ] A 401 from a progress or history call sends the user to log in and back afterwards.
- [ ] Unit tests pass for the loaded, empty, anonymous and error cases.

## Out of scope

- Editing or deleting attempts, and any aggregation over attempts.

## Constraints

- The client sends no user id; the server derives the user from the session.
- Depends on: T08, T09, T18, T19, T26.

---

## T31 - Frontend integration tests

**User story:** US-1, US-2, US-3, US-4, US-5, US-6
**Depends on:** T26, T30

## Goal

The frontend is proven against the running backend and database, so the wired-up screens are verified and not only unit tested.
Source: `product-spec.md` section 7 (testing), `README.md` section 9.

## Acceptance criteria

- [ ] `frontend/tests/integration/` contains tests that run against a running backend and a seeded test database, separate from the unit tests with the fake client.
- [ ] They cover the workflows of README section 9: browsing content, taking and submitting a quiz, registering, logging in and out, saving progress and reading quiz history.
- [ ] `bun run test` still runs only the unit tests; a separate documented command runs the integration tests, and README section 9 lists it.
- [ ] The suite starts from a clean database and passes twice in a row.
- [ ] The results are recorded in `docs/artifacts/integration/frontend-results.md` and linked from the issue.

## Out of scope

- Visual regression tests and end-to-end browser automation.

## Constraints

- Integration tests stay in `frontend/tests/integration/`, never mixed with unit tests.
- Depends on: T26, T30.

---

## T32 - Docker Compose runs the whole system

**User story:** none - supports all stories (`product-spec.md` section 7, operability)
**Depends on:** T31

## Goal

`docker compose up --build` from a clean clone starts the database, the API and the frontend, with the seed content already loaded.
Source: `product-spec.md` section 7 (operability), `README.md` section 8.

## Acceptance criteria

- [ ] `backend/Dockerfile` and `frontend/Dockerfile` build with pinned base image versions and no `latest` tag; the Go and Bun versions match the ones pinned elsewhere.
- [ ] `docker-compose.yml` defines `db`, `backend` and `frontend`, with the PostgreSQL major version as the image tag, health checks and start order.
- [ ] `docker compose up --build` from a clean clone gives a usable application in the browser with the seeded topics, lessons and questions.
- [ ] The seed runs on first start and is safe on every later start.
- [ ] `docker compose down -v` followed by `docker compose up --build` still works.
- [ ] Every variable the containers read is documented in `.env.example`, with a working default for local use and no real secret.

## Out of scope

- Production hardening, TLS termination, replicas and backups.

## Constraints

- Pinned versions only; the lockfiles are used in the images.
- Depends on: T31.

---

## T33 - CI runs the tests and the checks

**User story:** none - supports all stories (`product-spec.md` section 7, testing and operability)
**Depends on:** T14

## Goal

Every push and pull request runs the same checks as the ones documented in `AGENTS.md`, so broken work is caught before it is merged.
Source: `product-spec.md` section 7, `README.md` section 9.

## Acceptance criteria

- [ ] A workflow runs on every push and pull request and executes: `(cd backend && gofmt -l . && go vet ./...)`, `(cd backend && go test ./...)`, the frontend install, lint, format check and unit tests. The `bunx @redocly/cli lint openapi.yaml` step is enabled after T10, when `openapi.yaml` exists; until then the workflow runs the other checks.
- [ ] Go, Bun and the PostgreSQL version come from the files that pin them; every action is referenced by a pinned version.
- [ ] A separate job runs the backend integration tests against a PostgreSQL service; it is added after T21, when the first migration exists, and it creates its own test database and service credentials from the workflow, never from a shared or developer database.
- [ ] The workflow is green on the default branch and README section 9 says which checks CI runs.
- [ ] No step deploys anything.

## Out of scope

- Deployment (T34, T35), and pull request comment bots.

## Constraints

- Do not run a deploy command or workflow.
- Depends on: T14.

---

## T34 - CD builds and publishes images

**User story:** none - supports all stories (`product-spec.md` section 7, operability)
**Depends on:** T32, T33

## Goal

When the checks pass on the default branch, the images are built once and pushed to a container registry, so deployment is a promotion of a known build.
Source: `product-spec.md` section 7 (operability), `README.md` section 10.

## Acceptance criteria

- [ ] A workflow runs only after the checks of T33 pass on the default branch; it cannot be triggered by a pull request.
- [ ] It builds the backend and frontend images and pushes them to a registry, tagged with the commit SHA and with a moving tag for the branch.
- [ ] Registry credentials come from repository secrets; no credential is in a file or in a log.
- [ ] A pushed image is the same image that gets deployed, and the workflow records the SHA it pushed.
- [ ] README section 10 describes what the workflow does.

## Out of scope

- Deploying to a host: that is T35.
- Automatic rollback.

## Constraints

- Never trigger a workflow or deploy without the maintainer asking.
- Depends on: T32, T33.

---

## T35 - Deploy to the chosen hosting target

**User story:** none - supports all stories (`product-spec.md` section 7, operability)
**Depends on:** T34

## Goal

The application runs on a public URL, with the deployment steps written down so the same deployment can be repeated.
Source: `product-spec.md` section 7 (operability, reproducibility), section 10 (open question), `README.md` section 10.

## Acceptance criteria

- [ ] The hosting target is chosen by the maintainer and recorded in README section 10.
- [ ] A documented, repeatable command or workflow deploys the backend and frontend from the images of T34 and applies the migrations; secrets come from the platform's secret store.
- [ ] The deployed database is not the development database, and the seed has run there.
- [ ] `docs/deployment.md` describes the steps, the environment variables, how to verify the deployment and how to roll back.
- [ ] The deployed health endpoint answers 200, the public URL is in README section 10, and the public URL was opened and checked by a person.
- [ ] The deployment happened once from the written steps without undocumented manual actions.

## Out of scope

- Monitoring, alerting, backups and multiple environments.

## Constraints

- Blocked until the maintainer chooses the hosting target (`product-spec.md` section 10).
- Do not run deploy commands unless explicitly asked.
- Depends on: T34.

---

## T36 - README is runnable from a clean machine

**User story:** none - supports all stories (`product-spec.md` section 7, reproducibility)
**Depends on:** T35

## Goal

A person who has never seen the repository can clone it and run, test and deploy the project by following the README alone, because every documented command was executed from a clean clone.
Source: `product-spec.md` section 7 (reproducibility), `README.md` section 11.

## Acceptance criteria

- [ ] Every command in README sections 8, 9 and 11 is executed from a clean clone in this task and works exactly as written; any correction is a change to the README, not to the command list.
- [ ] The status table matches reality: finished items no longer say _planned_, and unfinished ones say what is missing.
- [ ] The repository structure section matches the real tree, including every new environment variable in `.env.example` and the integration test command from T31.
- [ ] The links to the repository, the issue tracker and the live application are filled in.
- [ ] The AI workflow entry for this task records which commands were run from the clean clone and their result.

## Out of scope

- Translations, screenshots and videos.

## Constraints

- No command is documented that was not run.
- Depends on: T35.

---

## T37 - Complete the AI workflow log and the stage artifacts

**User story:** none - supports all stories (`AGENTS.md` "Conventions", `product-spec.md` section 7)
**Depends on:** T36

## Goal

Every delivered issue has its entry in the AI workflow log and its stage artifact committed and linked, so the AI-assisted work can be reviewed end to end.
Source: `AGENTS.md` "Conventions" (log notable AI-assisted steps), `README.md` section 4.

## Acceptance criteria

- [ ] `docs/ai-workflow.md` has one entry per closed issue, following the entry template: tool, prompt, context given, result, artifacts and where they were used, what was changed by hand, verification, links.
- [ ] Every stage artifact in `docs/artifacts/` is linked from its log entry and from its issue.
- [ ] No secret, credential or personal data appears in the log or in any artifact; the sources of the skills installed in T02 are recorded.
- [ ] `docs/process.md` and `README.md` section 4 describe the workflow as it was actually followed, including anything that changed.
- [ ] The log entries reference the commits and issues they belong to.

## Out of scope

- New documentation that is not about the work that was done.

## Constraints

- Never record secrets or private data.
- Depends on: T36.

---

## Assumptions

1. **`/health` in the contract.** The specification does not describe a health endpoint, but the containers and the deployment check in T32-T35 need one. It is added to the contract in T10.
2. **Addressing.** Topics are addressed by slug (the data model has one); lessons are addressed by id (the data model has no lesson slug).
3. **Stateless quizzes.** No quiz-set table is added: the submission carries the question and chosen option ids, and the server checks that the questions belong to the topic. This follows the acceptance criterion that a foreign question is rejected.
4. **CSRF.** Proposed for the contract in T11: double-submit token, issued at registration and log in, sent in a header on state-changing requests that carry a session cookie. The specification only requires the protection, not the mechanism.
5. **Test runner for the frontend.** `bun run test` must run a runner other than Bun's built-in one (Vitest), as stated in `AGENTS.md`.
6. **In-memory repositories in the backend stage.** T14-T20 are unit tested with in-memory implementations behind the repository interfaces; they are replaced by PostgreSQL implementations in T21-T22.
7. **Extra course deliverables are not in this list.** The agent extension pack (`agent-capabilities/`, `agent-hooks/`, `mcp-server/`), `security/` and `ops/` are course modules that `AGENTS.md` and README section 12 mention; the instruction for this breakdown listed CI/CD, deployment, README and docs, and the AI workflow log only. They are not broken into tasks here.

## Decisions

Answered by the maintainer on 2026-10-17 and binding for the tasks above.

1. **Passwords.** Registration accepts at least 8 and at most 128 characters; passwords are hashed with Argon2id. The length rule goes into the contract in T11; the algorithm is chosen in T17.
2. **Session cookie `SameSite`.** `Lax`, so normal navigation back to the site keeps working.
3. **Seed content.** Drafted with AI assistance, reviewed by hand and committed as seed files. One lesson and six questions per topic is the minimum, so T24 stays one session.
4. **Frontend hosting.** The frontend is served by its own static server image; T32 assumes a static server image plus an API base URL from the environment.
5. **Hosting target.** The maintainer chooses it before the deployment task starts; the image task (T34) is not blocked by that choice.
6. **Task count.** The list stays as it is, with one exception: T22 and T23 were merged into a single persistence task, so the list ends at T37.
