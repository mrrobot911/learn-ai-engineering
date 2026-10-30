## 2026-10-26 — Entry 21: Rework backend backlog and create issues

- **Tool:** OpenCode (main session, acting as PM per docs/process.md) — direct work as instructed.
- **Task / context:** Rework T14-T26 in docs/artifacts/backlog/tasks.md against finished contract (openapi.yaml); create contract issue for progress 404 first; apply decisions (progress 404, no draft list, quiz submission rules); split heavy tasks with letter suffixes; create issues in dependency order.
- **Result:** Created contract issue #32, added 404 for progress ops to openapi.yaml. Reworked tasks.md splitting T14->T14a/T14b, T16->T16a/T16b, T17->T17a/T17b, T20->T20a/T20b, T21->T21a/T21b, T22->T22a/T22b. Created issues #34(T14a), #35(T14b), #36(T15), #37(T16a), #38(T16b), #39(T17a), #40(T17b), #41(T18), #42(T19), #43(T20a), #44(T20b), #45(T21a), #46(T21b), #47(T22a), #48(T22b), #49(T23), #50(T24), #51(T25), #52(T26). All dependencies updated to issue numbers; tasks.md updated with issue numbers.
- **Accepted / rejected / changed by hand:** Applied all decisions directly as requested.
- **Verification:** `bunx @redocly/cli lint openapi.yaml` exit 0 (4 warnings). Created issues via `gh issue create`.
- **Artifacts:** docs/artifacts/backlog/tasks.md (reworked), openapi.yaml (progress 404s).
- **Links:** issues #32,#33,#34-#52, PR #33.
