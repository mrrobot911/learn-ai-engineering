# Product Specification: AI Engineering Learning Platform

Status: draft v0.2

## 1. Problem

Developers who want to move into AI engineering have to piece together scattered articles, videos and docs, and have no simple way to check what they actually understood. This application gives them structured learning materials organised by topic and short multiple-choice quizzes to test their knowledge. Visitors can use everything without registering; registered users can additionally keep their progress and quiz history.

## 2. Target users

- **Visitor (anonymous):** a developer who wants to read lessons and take quizzes immediately, without an account.
- **Registered user:** a developer who wants to come back later and see which lessons they finished and how their quiz results changed over time.
- **Content maintainer:** the project author, who prepares and publishes lessons and questions (see section 6).

## 3. Scope

### In scope (version 1)

1. Learning content organised as topics, each containing ordered lessons.
2. Multiple-choice quizzes per topic, drawn at random from a larger question bank.
3. Anonymous access to all lessons and quizzes.
4. User accounts: registration, login, logout.
5. For registered users: lesson completion tracking and quiz attempt history.
6. Initial topics: LLMs, AI agents, MCP, RAG, embeddings, evaluation, deployment.
7. Initial published content (lessons and questions) available on first start (see US-8).

### Non-goals (version 1)

- AI-generated curriculum management in the product UI.
- Social features (comments, leaderboards, sharing).
- Interactive exercises beyond multiple-choice quizzes.
- Paid plans, payments, email verification flows, password reset by email.
- Admin web interface; content is loaded by the maintainer through the API or seed data.
- Mobile apps.

## 4. User stories and acceptance criteria

### US-1 Browse topics and lessons
As a visitor, I want to see a list of topics and read their lessons so that I can learn in a clear order.

- The home page lists all published topics.
- A topic page lists its lessons in order.
- A lesson page shows the lesson content.
- Unpublished content is never returned or displayed.

### US-2 Take a quiz
As a visitor, I want to take a quiz on a topic so that I can check my understanding.

- Starting a quiz on a topic returns exactly 5 questions, chosen randomly from that topic's published questions. The user cannot change the quiz size in version 1.
- If the topic has fewer than 5 published questions, all of them are returned.
- Each question has several answer options shown in a stable order, and exactly one correct answer.
- Correct answers are not sent to the client before the attempt is submitted.
- A submission lists the quiz questions and the chosen option for each; the chosen option may be empty.
- The server grades the submission. A question with no chosen option, or with an option that does not belong to that question, counts as incorrect. A submission that contains a question that does not belong to the topic is rejected with a validation error.
- The result returns the score and, for each question, whether it was correct and the right answer with a short explanation.
- Two attempts on the same topic can receive different question sets.

### US-3 Anonymous access
As a visitor, I want to use lessons and quizzes without registering.

- No lesson or quiz endpoint requires authentication, including quiz submission.
- An anonymous quiz result is shown after submission but is not stored under any account.

### US-4 Register and log in
As a visitor, I want to create an account so that my progress is saved.

- Registration requires a unique email and a password; invalid or duplicate input returns a clear validation error.
- Passwords are stored only as salted hashes.
- Login returns an authenticated session; logout ends it.
- The session is kept between requests with a cookie and expires after a fixed period (7 days); an expired or missing session is treated as not logged in.
- Protected endpoints return an authentication error when no valid session is present.

### US-5 Track lesson progress
As a registered user, I want to mark lessons as completed so that I can see how far I have got.

- A logged-in user can mark a lesson as completed and undo it.
- Topic and lesson pages show completion state for the logged-in user.
- A user can only see and change their own progress.

### US-6 Quiz history
As a registered user, I want my quiz attempts saved so that I can see how my results change.

- Submitting a quiz while logged in stores the attempt: topic, time, score and per-question results.
- Every submission creates a new attempt; earlier attempts are never overwritten.
- Each stored result keeps a snapshot of the question text, the selected answer text and the correct answer text at the time of the attempt, so later edits or removal of a question do not change past attempts.
- The user can view a list of their past attempts and open one to see its details.
- A user can only see their own attempts.

### US-7 Content preparation (maintainer)
As the maintainer, I want to prepare draft lessons and questions with AI assistance and review them before publication.

- Content has a status: draft or published; only published content is visible to users.
- The maintainer is a registered user with the `is_maintainer` flag, set directly in the database or seed data (no UI for it).
- Creating and editing draft lessons and questions requires a logged-in maintainer; other users receive an authorisation error.
- AI tooling acts only through these draft operations; it can create and edit drafts but cannot publish.
- Publishing is a separate explicit action by the maintainer after reviewing the draft; there is no automatic publication.
- Publishing makes the content visible to all users; unpublishing returns it to draft.

### US-8 Initial content
As a visitor, I want to find real learning material when the app starts so that it is useful immediately.

- On first start the system loads seed content so that each of the seven initial topics has at least 1 published lesson and 6 published questions.
- Loading seed content is repeatable and does not duplicate content that is already present.

## 5. Functional requirements summary

| Area | Requirement |
|---|---|
| Content | Topics with ordered lessons; each lesson has title, body (Markdown), status |
| Questions | Per topic; text, 2 to 6 options, one correct option, explanation, status |
| Quiz | Random subset of exactly 5 published questions per attempt (fewer if the topic has fewer) |
| Auth | Email and password; cookie-based session with a fixed expiry; `is_maintainer` flag for content operations |
| Progress | Per user and lesson: completed or not |
| History | Per user: quiz attempts with score and per-question results |
| Access | Content and quizzes are public; progress and history require login |

## 6. AI-assisted content preparation

Content moves through four separate stages:

1. **Preparation:** an AI agent creates or edits draft lessons and questions through the maintainer-only draft operations.
2. **Review:** the maintainer reads and corrects the draft.
3. **Publication:** only the maintainer changes the status to published, as an explicit action.
4. **Display:** users see published content only.

The interface for the agent is part of the agent extension pack and is documented there; it exposes only draft operations. Which exact operations it provides and how the agent authenticates are decided when the extension pack is designed and are not part of the user-facing product.

## 7. Non-functional requirements

- **API contract:** all frontend-backend communication is defined in `openapi.yaml`; the backend follows it and the frontend calls the API through one client module.
- **Persistence:** data survives restarts; separate configuration for development, test and production.
- **Security:** hashed passwords, input validation, users isolated from each other's data, secrets only via environment variables. The session cookie is `HttpOnly` and `SameSite`, and state-changing requests are protected against cross-site request forgery.
- **Testing:** unit tests for frontend and backend core logic; integration tests kept separate and documented.
- **Operability:** the whole system starts with one documented command (Docker Compose); CI runs tests and deploys only when they pass.
- **Reproducibility:** README explains setup, run, test and deploy from a clean machine.

## 8. Technology decisions

React with TypeScript (frontend, Bun as package manager and script runner), Go (backend), PostgreSQL (database), Docker Compose (local and deployment packaging), GitHub Actions (CI/CD).

## 9. Data model (high level)

- `topics`: id, slug, title, description, position, status
- `lessons`: id, topic_id, title, body, position, status
- `questions`: id, topic_id, text, explanation, status
- `answer_options`: id, question_id, text, position, is_correct
- `users`: id, email, password_hash, is_maintainer, created_at
- `sessions`: id, user_id, created_at, expires_at
- `lesson_progress`: user_id, lesson_id, completed_at
- `quiz_attempts`: id, user_id, topic_id, score, total, created_at
- `quiz_attempt_answers`: attempt_id, question_id (nullable reference), question_text, selected_option_text (empty if unanswered), correct_option_text, is_correct

## 10. Open questions

- Hosting target for the deployment.
- Which operations the agent interface exposes and how the agent authenticates (decided with the extension pack).
- Exact number of lessons and questions in the seed content beyond the minimum in US-8.

Decided: quiz size is always 5; session is cookie-based with a 7-day expiry; the maintainer is a user with an `is_maintainer` flag; AI-assisted curriculum generation stays a non-goal for version 1.
