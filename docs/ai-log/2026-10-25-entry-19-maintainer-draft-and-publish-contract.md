## 2026-10-25 — Entry 19: maintainer draft and publish contract (issue #13)

- **Tool:** OpenCode (orchestrator with the product manager, software engineer and QA roles) for the first version; the rework below was prepared outside OpenCode and applied by the maintainer, because the roles ran out of context on the 1700-line `openapi.yaml`.
- **Task / context:** issue #13 (T13) with its criteria; `product-spec.md` US-7; the sections of `openapi.yaml` it touches; PR #29.
- **Result:** eight maintainer operations in `openapi.yaml` (create and update a draft lesson or question, publish and unpublish each), 401 and 403 `authorization_error`, no draft operation changes the status.
- **Review finding:** the first version let a draft request set `status: published`, added an invented `maintainerAuth` header scheme and did not express "exactly one correct option" in the schema, although the role QA gave PASS. The rework removes `status`, `id` and `topic_id` from the request bodies (new `DraftLessonRequest` and `DraftQuestionRequest`), removes `maintainerAuth` and adds `contains` with `minContains` and `maxContains` 1 on `is_correct`.
- **Accepted / rejected / changed by hand, and why:** left for the maintainer.
- **Verification:** `bunx @redocly/cli lint openapi.yaml`: exit 0, 4 warnings, no error; the bundled contract was checked for the request properties, the security of the eight operations and the `contains` rule. The QA role was not run again on the rework: it passed the first version with the defects above and its context ran out on this file, so the maintainer checked the criteria with the lint and a review of the bundled contract.
- **README:** the status table and section 6 now say that the prototype is done and that the contract covers authentication, progress, history and the maintainer operations.
- **Links:** issue #13, PR #29.
