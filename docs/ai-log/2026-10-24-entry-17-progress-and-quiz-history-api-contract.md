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
