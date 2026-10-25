# AI workflow log

This file records how AI tools were used to build this project: what was asked, what context was given, what was produced, what was changed by hand and how it was verified. One entry per significant task. Secrets and private data are never recorded here.

Related: [`AGENTS.md`](../AGENTS.md) (instructions for AI tools), [`product-spec.md`](../product-spec.md) (specification), GitHub Issues (tasks and acceptance criteria).

## Entry template

```text
## <date> — <task name> (issue #N)
- Tool:
- Prompt / task given:
- Context given (files, issue, spec):
- Result:
- Artifacts produced (path) and where the next stage used them:
- Accepted / rejected / changed by hand, and why:
- Verification (tests run and result, manual checks):
- Links (commit, issue, tool session):
```

## Entries

Entries 0-12 are kept unchanged in [`ai-workflow-archive.md`](ai-workflow-archive.md); entries 13 and later follow here, newest last. Index of all entries:

| Entry | Date | About |
|---|---|---|
| 0 | 2026-10-17 | project planning, specification and instructions (no issue yet) (archive) |
| 1 | 2026-10-17 | roles, task template and process (no issue yet) (archive) |
| 2 | 2026-10-17 | task breakdown of the specification (issues #1-#13 for T01-T13) (archive) |
| 3 | 2026-10-18 | the orchestrator runs the whole cycle, one review at the end (no issue yet) (archive) |
| 4 | 2026-10-18 | the cycle on issue #1, then its cancellation (issue #1, cancelled) (archive) |
| 5 | 2026-10-18 | the project skills for UI and React work are installed (issue #2) (archive) |
| 6 | 2026-10-19 | pull request template, opened by the orchestrator (no issue yet) (archive) |
| 7 | 2026-10-19 | central API client module with a mock implementation (issue #3) (archive) |
| 8 | 2026-10-19 | revision of the prototype tasks T04-T09 (backlog change, issues #4-#9) (archive) |
| 9 | 2026-10-19 | topic list and topic page on the API client (issue #4) (archive) |
| 10 | 2026-10-20 | lesson page on the API client (issue #5) (archive) |
| 11 | 2026-10-20 | quiz flow on the API client (issue #6) (archive) |
| 12 | 2026-10-20 | registration, log in and log out (issue #7) (archive) |
| 13 | 2026-10-21 | lesson completion on the API client (issue #8) (this file) |
| 14 | 2026-10-21 | quiz history on the API client (issue #9) (this file) |
| 15 | 2026-10-22 | the content and quiz API contract (issue #10) (this file) |
| 16 | 2026-10-22 | authentication and session API contract (issue #11) (this file) |
| 17 | 2026-10-24 | progress and quiz history API contract (issue #12) (this file) |
| 18 | 2026-10-25 | limit context size and log entry length (chore-context-limits) (this file) |

## 2026-10-21 — Entry 13: lesson completion on the API client (issue #8)

- **Tool:** OpenCode (software engineer role) implementing groomed issue #8 on branch `issue-8`.
- **Prompt / task given:** implement the groomed issue #8 (T08): let a logged-in visitor mark a lesson complete and undo it through `frontend/src/api/` only, show the same completion state on the lesson page and on the topic page that lists the lesson, show no control and no count to an anonymous visitor, and delete the progress path of the imported store (`forge.progress`, `completedLessons`, `toggleLessonComplete`, `isLessonComplete`) without touching the quiz-history path that belongs to T09.
- **Context given:** issue #8 with its fourteen acceptance criteria, its out-of-scope list and its constraints; `AGENTS.md` (Conventions, Gotchas, Skills, Smallest complete change, Commands); `product-spec.md` section 4 US-5 and section 5 (Access, Progress); `docs/artifacts/backlog/tasks.md` (T08); the code the issue touches: `frontend/src/api/types.ts` (the `ApiClient` type, unchanged), `frontend/src/api/mock.ts`, `frontend/src/pages/LessonPage.tsx` and `frontend/src/pages/TopicPage.tsx` with their tests, `frontend/src/store/{context,StoreProvider,useStore,index}.ts(x)`, `frontend/src/session/` (from issue #7), `frontend/src/pages/HistoryPages.tsx` (the remaining consumer of the store), `frontend/src/components/ui.tsx`; the four project skills, loaded with the `skill` tool before writing the UI/React code (their vendored files were not edited).
- **Result:**
  - `frontend/src/hooks/useProgress.ts` (new): the progress of the logged-in user, read only through the client — `api.getProgress()` once per user, `api.markLessonComplete(lessonId)` / `api.undoLessonComplete(lessonId)` for the toggle, the returned list kept in the hook state so the two pages share one source. `ids` is `null` until the list arrives and stays `null` when the call fails, and `failed` tells the two apart.
  - `frontend/src/pages/LessonPage.tsx`: a completion card with one `<button>` ("Mark as complete", `aria-pressed` = the completion state) for a logged-in visitor whose progress could be read; the button is disabled and shows the shared `Spinner` while the call is pending and a second activation is ignored, a rejected call shows a `role="alert"` message, leaves `aria-pressed` unchanged and enables the button again, and a rejected `getProgress()` leaves the lesson content on screen with no control and a short status line.
  - `frontend/src/pages/TopicPage.tsx`: a progress card for a logged-in visitor with the count of that topic as `2 of 3` and the shared `ProgressBar`, plus one `Badge` with the accessible name `Completed` per completed lesson (the completed rows also show the check icon in place of the number). No count and no badge for an anonymous visitor, whose `api.getProgress` is never called.
  - The progress path of the imported store is gone: `frontend/src/store/context.ts` no longer exposes `completedLessons`, `toggleLessonComplete` and `isLessonComplete`, `frontend/src/store/StoreProvider.tsx` no longer reads or writes the `forge.progress` key, and the quiz-history path (`quizHistory`, `addQuizAttempt`, `getAttempt`, `forge.history`) stays untouched for `HistoryPages.tsx` (T09). No `forge.progress` key is written to `localStorage` or `sessionStorage` any more; the only browser storage the app writes is the quiz history.
  - `frontend/src/components/ui.tsx`: the shared `Badge` forwards the rest of the span attributes, so the completion badge can carry `aria-label="Completed"`.
  - `frontend/src/api/mock.ts`: the in-memory progress helper was renamed `progressOf` (no behaviour change) — the mock already required a session for the three progress calls.
  - Tests: `frontend/src/pages/LessonPage.test.tsx` (seven new cases: the control for a logged-in visitor, the completed state, the pending call, the rejected call, the keyboard activation with Enter and Space, the anonymous visitor and the rejected `getProgress()`), `frontend/src/pages/TopicPage.test.tsx` (four new cases: `2 of 3` with two `Completed` badges, `0 of 3`, the anonymous visitor and the rejected progress) and `frontend/tests/integration/lessonCompletion.test.tsx` (the mock as the backend: a lesson marked complete on the lesson page is counted as completed on the topic page, and the three progress calls reject without a session).
  - `frontend/tsconfig.app.json` includes the `tests` folder, so the new integration test is type-checked with the sources.
- **Accepted / rejected / changed by hand, and why:** left for the maintainer. The `ApiClient` type, `product-spec.md` and the API contract were not touched; `openapi.yaml` does not exist yet, so no contract lint was needed. The `format`/`format:check` scripts do not exist in `frontend/package.json` (recorded in Entry 9), so those checks were skipped deliberately. The `Badge` was given a rest-attribute spread instead of a new `label` prop, to keep the shared component the way the rest of the code uses it.
- **Verification:**
  - `(cd frontend && bun run test)`: 9 files, 70 tests pass, exit 0 (was 8 files / 57 tests before this change).
  - `(cd frontend && bun run lint)`: exit 0.
  - `(cd frontend && bun run typecheck)`: exit 0.
  - `rg -n "forge\.progress|completedLessons|toggleLessonComplete|isLessonComplete" frontend/src` returns nothing; `rg -n "localStorage|sessionStorage" frontend/src` returns only the quiz-history read/write in `frontend/src/store/StoreProvider.tsx` and the assertion in the quiz test.
  - The pages, the hook and the tests were read back after writing; the diff touches only the two pages, the new hook, the two test files, the integration test, the shared `Badge`, the store's two files, the mock's helper name and the typecheck include.
- **Artifacts:** `frontend/src/hooks/useProgress.ts`, `frontend/src/pages/LessonPage.tsx`, `frontend/src/pages/TopicPage.tsx` and their tests — issues #9 and #12 build on these.
- **Links:** issue #8, branch `issue-8`; the commits are referenced in the issue comment; QA verdict [PASS](https://github.com/mrrobot911/learn-ai-engineering/issues/8#issuecomment-6317278988); pull request #21, opened by the orchestrator after that verdict.

## 2026-10-21 — Entry 14: quiz history on the API client (issue #9)

- **Tool:** OpenCode (software engineer role) implementing groomed issue #9 on branch `issue-9`.
- **Prompt / task given:** implement the groomed issue #9 (T09): the quiz history list and the attempt detail read their data through the API client only (`api.listAttempts()`, `api.getAttempt(id)`), the list shows the topic title, the score as `score/total` and the UTC day of the attempt in the order the client returns, the detail shows the stored question text, the selected answer text and the correct answer text of every answer, the empty state, the anonymous visitor, the loading states and the shared error state are in place, the seeded attempt `att-seed-1` demonstrates that a stored attempt keeps the wording of its moment, the imported store is deleted and the attempt types of the imported prototype are gone.
- **Context given:** issue #9 with its nineteen acceptance criteria, its out-of-scope list and its constraints; `AGENTS.md` (Conventions, Gotchas, Skills, Smallest complete change, Commands); `product-spec.md` section 4 US-3/US-6 and section 5 (Access, History); `docs/artifacts/backlog/tasks.md` (T09); the code the issue touches: `frontend/src/pages/HistoryPages.tsx`, `frontend/src/store/` with `frontend/src/store.tsx` and `frontend/src/useStore.ts`, `frontend/src/App.tsx`, `frontend/src/api/types.ts` (the `ApiClient` type, unchanged), `frontend/src/api/mock.ts`, `frontend/src/data/content.ts`, `frontend/src/types.ts`, `frontend/src/components/ui.tsx`, `frontend/src/router.ts`, `frontend/src/session/`; the tests of the other pages as the style of the project; the four project skills, loaded with the `skill` tool before writing the UI/React code (their vendored files were not edited).
- **Result:**
  - `frontend/src/pages/HistoryPages.tsx` reads only through `frontend/src/api/`: the list calls `api.listAttempts()` once for the logged-in visitor and shows one row per summary in the order the client returned them, with the topic title, the score as `score/total` and the date as the UTC day (`YYYY-MM-DD`) of `createdAt`; every row is a link to `#/history/<id>`. The detail calls `api.getAttempt(attemptId)` and shows, per answer, the stored question text, the selected answer text (the marker "No answer selected" when it is empty) and the stored correct answer text — also for the questions that were answered correctly, which the imported page hid. The statistics of the imported page (attempts, average, best score) are gone, and so is its read of the fixtures (`getTopicById`).
  - The states: "Loading history…" and "Loading attempt…" while the call is pending, the shared `ErrorState` ("Something went wrong" and a "Try again" button that repeats the call) when the call rejects, the empty state "No quiz attempts yet" and, for an anonymous visitor, a link with the accessible name "Log in" (`#/login`); the client is never called without a session.
  - The imported store is deleted: `frontend/src/store.tsx`, `frontend/src/useStore.ts` and the directory `frontend/src/store/` are gone, `frontend/src/App.tsx` no longer renders `StoreProvider`, and `frontend/src/types.ts` exports neither `QuizAttempt` nor `QuizResultAnswer` any more (`AnswerOption.isCorrect` stays, the mock grades with it). The seeded attempts of `frontend/src/data/content.ts` use the attempt types of `frontend/src/api/types.ts` and carry ISO-8601 timestamps.
  - The fixture difference: the stored answer of `att-seed-1` for `q-llm-1` keeps the wording of its moment ("What is the main training objective of a large language model?" and "Predict the next word in a sequence"), while the current question reads "What is the core training objective of a large language model during pre-training?" and "Predict the next token in a sequence".
  - Tests: `frontend/src/pages/HistoryPages.test.tsx` (11 jsdom tests: the list order and what a row shows, the empty state, the anonymous visitor on both pages, the loading and the error state of both pages, the stored wording, the unanswered question and the correct answer of every question), `frontend/src/data/content.test.ts` (the stored wording differs from the current wording of `q-llm-1`), `frontend/tests/integration/quizHistory.test.tsx` (the mock as the backend: the seeded attempts newest first, the stored attempt read through the client and the attempts of another account) and two cases more in `frontend/src/api/mock.test.ts` (an ISO-8601 `createdAt` for every summary, and `getAttempt` rejecting an attempt of another account).
- **Accepted / rejected / changed by hand, and why:** left for the maintainer. The `ApiClient` type, `product-spec.md` and the acceptance criteria were not touched; `openapi.yaml` does not exist yet, so no contract lint was needed. The per-question explanation of a stored attempt is not shown: the issue leaves that decision to T12 (issue #12), and US-6 names the three snapshot texts. The `format`/`format:check` scripts do not exist in `frontend/package.json` (recorded in Entry 9), so those checks were skipped deliberately. The statistics of the imported page were dropped as the issue asks. `frontend/src/api/mock.ts` needed no change: it already returns the seeded attempts newest first and refuses the attempts of another account.
- **Verification:**
  - `(cd frontend && bun run test)`: 12 files, 85 tests pass, exit 0 (was 9 files / 70 tests before this change).
  - `(cd frontend && bun run lint)`: exit 0, no error and no warning.
  - `(cd frontend && bun run typecheck)`: exit 0.
  - `(cd frontend && bun run build)`: succeeds; the build output is not committed.
  - `rg -n "forge\.history|quizHistory|addQuizAttempt|useStore" frontend/src --glob '!frontend/src/api/**'` returns nothing; `rg -n "from '@/data/content'|from '@/useStore'|from '@/store'" frontend/src/pages` returns nothing; `rg -n "QuizAttempt|QuizResultAnswer" frontend/src/types.ts` returns nothing.
  - The pages and the tests were read back after writing; the diff touches only the two history pages, the fixture, the shared types, `App.tsx`, the three new test files, the mock test, the deleted store and this log.
- **Artifacts:** `docs/artifacts/frontend-prototype/prototype-stage.md` — the screens of the prototype with their routes, how to run the prototype and the fixture difference the seeded attempts demonstrate; issue #12 (the contract) and the stages after it use it as the description of the prototype. The skills artifact of the same stage (`docs/artifacts/frontend-prototype/skills.md`) stays as it is.
- **Links:** issue #9, branch `issue-9`; the commit is referenced in the issue comment; QA verdict [PASS](https://github.com/mrrobot911/learn-ai-engineering/issues/9#issuecomment-6320341195); pull request #22, opened by the orchestrator after that verdict.

## 2026-10-22 — Entry 15: the content and quiz API contract (issue #10)

- **Tool:** OpenCode (software engineer role) implementing groomed issue #10 (T10) on branch `issue-10`.
- **Prompt / task given:** implement the groomed issue #10 (T10): write `openapi.yaml` in the repository root as the public interface of version 1 for the content and quiz endpoints — exactly six operations (`GET /topics`, `GET /topics/{topicSlug}`, `GET /lessons/{lessonId}`, `POST /topics/{topicSlug}/quiz`, `POST /topics/{topicSlug}/quiz/submissions`, `GET /health`), topics addressed by slug and lessons by id, snake_case property names, one shared error object, no security scheme and an empty `security` list, an extendable component layout — and make `bunx @redocly/cli lint openapi.yaml` exit 0 without a problem of severity `error`.
- **Context given:** issue #10 with its seventeen acceptance criteria, its out-of-scope list and its constraints; `AGENTS.md` (Conventions, Gotchas, Smallest complete change, Commands); `product-spec.md` section 4 (US-1, US-2, US-3), section 5 (Content, Questions, Quiz), section 7 (API contract) and section 9 (data model); `docs/artifacts/backlog/tasks.md` (T10 and the Assumptions 1-3); `docs/task-template.md`; `docs/process.md`; `docs/artifacts/frontend-prototype/prototype-stage.md`; `frontend/src/api/types.ts` (the prototype client, read only); the rules of the built-in `recommended` configuration of the Redocly CLI 2.67.1.
- **Result:**
  - `openapi.yaml` (OpenAPI 3.1.0, 500 lines) in the repository root:
    - `info` carries the title, the version `1.0.0` and a description that names the scope (the content and quiz endpoints of version 1); `servers` holds one entry, `http://localhost:8080`; the root `security` list is empty and there is no `securitySchemes`, so no operation requires authentication.
    - The six operations above, each with a `summary`, a unique `operationId` (`listTopics`, `getTopic`, `getLesson`, `startQuiz`, `submitQuiz`, `getHealth`) and one of the three tags `content`, `quiz` and `health`, each of which has a description.
    - The topic is addressed by its slug (`topicSlug` in the path) and the lesson by its id (`lessonId`); no operation looks a topic up by id, and both topic responses carry the `id` and the `slug`.
    - `components/schemas`: `Error` (the one error object: required `code` and `message`, optional `details`; the description documents `validation_error` for the status 400, `not_found` for 404, `conflict` for 409 and `internal_error` for 500), `TopicSummary` (`id`, `slug`, `title`, `description`, `position`), `TopicList`, `TopicDetail` (the same five fields plus `lessons`), `Lesson` (`id`, `topic_id`, `title`, `body`, `position`; the `body` is documented as Markdown), `QuizOption`, `QuizQuestion`, `QuizQuestionList`, `QuizSubmissionAnswer`, `QuizSubmissionRequest`, `QuizSubmissionResultAnswer`, `QuizSubmissionResult` and `HealthStatus`. Every schema is defined once and referenced with `$ref`; there is no second error schema.
    - The start-quiz response holds `id`, `text` and `options` per question and `id` and `text` per option, and nothing else: no `is_correct`, no `correct_option` and no `explanation` reaches the client before the submission. The submission request carries only `answers` (the question id and the chosen option id, which may be `null`) and no user id and no session token; the submission response carries `score`, `total` and, per question, the question id, `is_correct`, the text of the correct option and the `explanation`.
    - The descriptions state the rules the specification asks for: published content only (a draft topic, lesson or question is never returned), the topics and the lessons of a topic in `position` order, the fixed quiz size of 5 (or all published questions of the topic when it has fewer; the client cannot choose it), the rejection of a submission that contains a question which does not belong to the topic (400), the anonymous submission (a submission without a session is answered normally) and the use of `GET /health` by the container health check and by the deployment check.
    - The document holds no authentication, progress, history or maintainer path and no security scheme, so #11, #12 and #13 add to it without changing what is here.
  - Nothing else changed: no file under `frontend/`, `backend/` or `.github/`, no test file, no dependency and no configuration file. The branch also carries the orchestrator's companion commit `af2f464`, which records the contract follow-ups in T27 and T28.
  - `README.md` (commit `3ab0bea`, by the orchestrator): the status table said `OpenAPI contract | planned` and section 6 ended with _Planned._; both now say that the contract exists and that the content and quiz endpoints are described. The QA run flagged them as stale. The two rows that are still stale for the same reason and belong to other stages (`Task tracker` and `Frontend prototype with mock API client`) were left alone.
- **Accepted / rejected / changed by hand, and why:** left for the maintainer. The engineer noted the four lint warnings below, that the server URL of the deployment is still an open question of the specification (the local URL stands for now) and that the prototype client of `frontend/src/api/types.ts` addresses topics by id, which the real HTTP client of #27 and #28 will reconcile with the slug this contract decides.
- **Verification:**
  - `bunx @redocly/cli lint openapi.yaml`: exit 0, "Woohoo! Your API description is valid. 🎉", "You have 4 warnings." and no problem of severity `error`. The four warnings are `info-license` (the Info object has no `license` field; the repository has no licence file, so none was invented), `no-server-example.com` (the server URL is `http://localhost:8080`, the local stack of the repository) and `operation-4xx-response` twice (`GET /topics` and `GET /health` have no 4xx response, because a list of published topics and a health check cannot fail with a client error). The criterion of the issue accepts warnings.
  - `bunx @redocly/cli bundle openapi.yaml` succeeds, so every `$ref` of the document resolves.
  - Manual check of the criteria against the bundled document: exactly six operations, each with a summary and a unique operation id; the start-quiz response resolves to the properties `id`, `text` and `options` only; every error response of every operation resolves to the one `Error` schema and there is no other error schema; the request body has the property `answers` only and its items carry `question_id` and `option_id` (type `["string", "null"]`); the submission result has `score`, `total` and `answers`, and its items carry `question_id`, `is_correct`, `correct_option_text` and `explanation`; the health response is an object with the required `status` string; `components` holds only `schemas`.
  - No test was run and none was added: this change adds no runtime behaviour (criterion 17 of the issue).
- **Artifacts:** `openapi.yaml` — the artifact of the contract stage. Issues #11 (authentication and session), #12 (progress and quiz history) and #13 (maintainer draft and publish operations) add to it; the backend of #14-#20 and the HTTP client of #27-#30 are written against it.
- **Links:** issue #10, branch `issue-10`; the commit is referenced in the issue comment.

## 2026-10-22 — Entry 16: authentication and session API contract (issue #11)

- **Tool:** OpenCode (software engineer role) implementing groomed issue #11 (T11) on branch `issue-11`.
- **Prompt / task given:** implement the groomed issue #11 (T11): extend `openapi.yaml` to define registration, log in, log out and the current session endpoints, including session cookie and CSRF protection; session endpoint returns 401 when no valid session; session cookie documented as `HttpOnly`, `SameSite`, `Secure` in production, path `/`, fixed expiry 7 days; CSRF token issued at registration and log in must be sent in header on every state-changing request that carries a session cookie (`GET` does not require it); quiz submission carrying a session cookie is state-changing and needs the token, anonymous submission without a session cookie accepted without it; register returns validation error for invalid input and conflict for duplicate email; log in returns authentication error for wrong credentials; registration rejects password shorter than 8 or longer than 128 characters with validation error, limits part of request schema; no response schema contains `password_hash` or any session token; `bunx @redocly/cli lint openapi.yaml` passes.
- **Context given:** issue #11; `AGENTS.md` (Conventions, Gotchas); `product-spec.md` section 4 (US-4), section 7 (security), section 9 (data model); `openapi.yaml` (as of issue #10); `docs/process.md`, `docs/task-template.md`.
- **Result:**
  - Extended `openapi.yaml`:
    - Updated `info.description` to include authentication and session and CSRF rules; removed the note saying they are added later.
    - Added tag `auth` and added operations `POST /auth/register`, `POST /auth/login`, `POST /auth/logout`, `GET /auth/session`.
    - Added security schemes `sessionCookie` (apiKey in cookie named `session`) and `csrfToken` (apiKey in header named `X-CSRF-Token`).
    - `/auth/logout` requires both `sessionCookie` and `csrfToken` (state-changing with session cookie) and declares a 403 (CSRF failure) response. `/auth/session` requires `sessionCookie` only and returns 401 when no valid session (`GET` does not require CSRF). Register/login are unauthenticated but set the session cookie and return `csrf_token`.
    - Added schemas `RegisterRequest` (email, password with `minLength: 8`, `maxLength: 128`), `LoginRequest`, `UserSummary`, `AuthResponse` (user + `csrf_token`), `SessionResponse` (user + `csrf_token`), `LogoutResponse` (`ok`). No schema contains `password_hash` or any session token other than the documented CSRF token.
    - The shared `Error` schema documents the error codes for the status 401 (`authentication_error`) and the status 403 (`forbidden`); `/auth/session`, `/auth/logout` and `POST /topics/{topicSlug}/quiz/submissions` declare the matching 401/403 responses, the quiz submission describing the conditions (session cookie present but invalid; CSRF token missing or invalid).
    - Updated `POST /topics/{topicSlug}/quiz/submissions` description to document CSRF requirement when carrying a session cookie and that anonymous submissions without a session cookie are accepted without it; the request/response schemas unchanged.
  - No runtime code, tests, dependencies, or other files changed.
  - Rework after the first QA PASS (2026-10-22), requested by the maintainer: (1) `GET /auth/session` now returns the CSRF token, so a client that keeps the session cookie across a page reload can send state-changing requests without re-authenticating; (2) the 401/403 error codes are now documented in the `Error` schema and the operations that can fail with them declare the matching responses. The T11 acceptance criteria and issue body were updated by the maintainer to match.
  - Second rework (2026-10-23), requested by the maintainer: `POST /topics/{topicSlug}/quiz/submissions` no longer declares a 401 and now states that a missing, invalid or expired session cookie is treated as anonymous — the submission is graded and answered normally and nothing is stored. The only 403 on that operation is a CSRF failure with a valid session. The shared `Error` schema now documents `authorization_error` for a valid session that lacks the required permission (for example a user without `is_maintainer`) next to `forbidden` for a CSRF failure, and records the retry decision: a client refreshes the CSRF token and retries a `forbidden` response, and does not retry an `authorization_error` response; T13 and T20 use the same `authorization_error` code. The CSRF token is documented as issued at registration, log in and on every read of the current session. The T11 acceptance criteria and issue body were updated by the maintainer to match.
- **Accepted / rejected / changed by hand, and why:** left for the maintainer. Contract-first rule followed.
- **Verification:**
  - `bunx @redocly/cli lint openapi.yaml`: exit 0, valid with warnings only (no errors), also after the second rework.
- **Artifacts:** `openapi.yaml` — updated for issue #11; issues #12 and #13 will extend it further.
- **Links:** issue #11, branch `issue-11`, pull request #24.

## 2026-10-24 — Entry 17: progress and quiz history API contract (issue #12)

- **Tool:** OpenCode (software engineer role) implementing groomed issue #12 on branch `issue-12`.
- **Prompt / task given:** implement the groomed issue #12 (T12): add contract for `GET /me/progress`, `PUT /me/progress/lessons/{lessonId}`, `DELETE /me/progress/lessons/{lessonId}` (require session, 401 without session; no user id in request; ownership from session; idempotent mark complete, undo unmarked not error) and `GET /me/attempts`, `GET /me/attempts/{attemptId}` (require session, 401 without session; `GET /me/attempts` returns attempts newest first; no user id in request). Schemas: progress response (array of lesson ids), attempt list item and attempt detail with answers. `QuizAttemptAnswer` must have `question_id` nullable, `selected_option_text` allowed empty string, the three snapshot text fields `question_text`, `selected_option_text`, `correct_option_text` (and `is_correct`); use snake_case, `$ref`, shared `Error` schema, `authentication_error` for 401. Use `sessionCookie` security; state-changing need CSRF as documented (`GET` no CSRF). Consistent with #11. Add to `openapi.yaml` only, run `bunx @redocly/cli lint openapi.yaml`, add log entry to `docs/ai-workflow.md`, commit with `#12` in message.
- **Context given:** issue #12 (groomed), `product-spec.md` US-5/US-6, section 9 data model, current `openapi.yaml` (paths/security/schemas), `AGENTS.md`.
- **Result:**
  - Extended `openapi.yaml`:
    - Added tags `progress` and `history`.
    - Added operations: `GET /me/progress` (sessionCookie only, 401 with `authentication_error`), `PUT /me/progress/lessons/{lessonId}` and `DELETE /me/progress/lessons/{lessonId}` (sessionCookie + csrfToken, 401 with `authentication_error`, 403 CSRF as documented; idempotent mark complete, undo unmarked not an error; no user id in request).
    - Added operations: `GET /me/attempts` (sessionCookie only, 401 with `authentication_error`, newest first) and `GET /me/attempts/{attemptId}` (sessionCookie only, 401 with `authentication_error`, 404 if not found/not owned).
    - Added schemas: `ProgressResponse` (array of lesson id strings), `QuizAttemptAnswer` (required `question_id` as `string|null`, `question_text`, `selected_option_text` allowed empty string, `correct_option_text`, `is_correct`), `QuizAttemptListItem`, `QuizAttemptList`, `QuizAttemptDetail` (with `answers` array of `QuizAttemptAnswer`).
- **Accepted / rejected / changed by hand, and why:** left for the maintainer. Contract-first; no code/migrations/tests/frontend/backend changes.
- **Verification:**
  - `bunx @redocly/cli lint openapi.yaml`: exit 0, valid with warnings only (no errors).
- **Artifacts:** `openapi.yaml` updated for issue #12; issues #13+ extend it further.
- **Links:** issue #12, branch `issue-12`.

## 2026-10-25 — Entry 18: limit context size and log entry length (chore-context-limits)

- **Tool:** OpenCode (main session) implementing chore-context-limits branch.
- **Prompt / task given:** limit context size and log entry length: (1) docs/process.md needs the new "Context size" section added before "Notes"; (2) AGENTS.md: add to AI workflow log convention that entries stay short (task, context files by name, result in 5-8 lines, verification commands with exit codes, maintainer decision, links; no file lists or diffs) and add (cd frontend && bun run typecheck) to run with bun run test and bun run lint before frontend changes are reported as done; (3) .opencode/agents/software-engineer.md: add that log entries keep the length set in AGENTS.md and hold no file lists or diffs.
- **Context given:** docs/process.md (partial), AGENTS.md, .opencode/agents/software-engineer.md, docs/ai-workflow.md (last entry).
- **Result:** Added "Context size" section to docs/process.md before "Notes"; updated AGENTS.md to specify short log entries and added typecheck command; updated software-engineer.md to specify log entry length; added short entry to docs/ai-workflow.md.
- **Accepted / rejected / changed by hand, and why:** left for the maintainer.
- **Verification:**
  - `git diff`: exit 0, changes are to docs/process.md, AGENTS.md and .opencode/agents/software-engineer.md
  - `grep -n "Context size" docs/process.md`: exit 0, section found at line 49
  - `git diff AGENTS.md`: exit 0, added log entry length specification and typecheck command
  - `git diff .opencode/agents/software-engineer.md`: exit 0, added log entry length specification
- **Artifacts:** docs/process.md updated with "Context size" section; AGENTS.md updated with context size and log entry length guidelines; .opencode/agents/software-engineer.md updated with log entry length specification.
- **Links:** branch chore-context-limits.

## 2026-10-25 — Entry 19: maintainer draft and publish contract (issue #13)

- **Tool:** OpenCode (orchestrator with the product manager, software engineer and QA roles) for the first version; the rework below was prepared outside OpenCode and applied by the maintainer, because the roles ran out of context on the 1700-line `openapi.yaml`.
- **Task / context:** issue #13 (T13) with its criteria; `product-spec.md` US-7; the sections of `openapi.yaml` it touches; PR #29.
- **Result:** eight maintainer operations in `openapi.yaml` (create and update a draft lesson or question, publish and unpublish each), 401 and 403 `authorization_error`, no draft operation changes the status.
- **Review finding:** the first version let a draft request set `status: published`, added an invented `maintainerAuth` header scheme and did not express "exactly one correct option" in the schema, although the role QA gave PASS. The rework removes `status`, `id` and `topic_id` from the request bodies (new `DraftLessonRequest` and `DraftQuestionRequest`), removes `maintainerAuth` and adds `contains` with `minContains` and `maxContains` 1 on `is_correct`.
- **Accepted / rejected / changed by hand, and why:** left for the maintainer.
- **Verification:** `bunx @redocly/cli lint openapi.yaml`: exit 0, 4 warnings, no error; the bundled contract was checked for the request properties, the security of the eight operations and the `contains` rule. The QA role was not run again on the rework: it passed the first version with the defects above and its context ran out on this file, so the maintainer checked the criteria with the lint and a review of the bundled contract.
- **README:** the status table and section 6 now say that the prototype is done and that the contract covers authentication, progress, history and the maintainer operations.
- **Links:** issue #13, PR #29.
