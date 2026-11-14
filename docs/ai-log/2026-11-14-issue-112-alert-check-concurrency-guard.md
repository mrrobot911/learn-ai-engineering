## 2026-11-14 — Guard concurrent alert checks (issue #112)
- Tool: opencode (iceman-free)
- Task given: Work on issue 112; add a fail-fast guard in `observability/alert-lib.sh` so alert checks cannot run concurrently on the same `alert-check` project/resources, and document cleanup in `observability/README.md`. Verify by running the two alert checks one at a time.
- Context given (files by name): `gh issue view 112`, `AGENTS.md`, `observability/alert-lib.sh`, `observability/alertmanager-webhook-check.sh`, `observability/alert-integration-check.sh`, `observability/README.md`
- Result (5-8 lines): Added a fail-fast check at the start of `ac_init` (before trap/temp setup) that detects existing `alert-check` Compose containers or the `alert-check-listener`/`alert-check-curl` containers. When present, the script exits with a message instructing to run checks one at a time and how to clean up: `docker compose -p alert-check down -v; docker rm -f alert-check-listener alert-check-curl`. Added one sentence to `observability/README.md` documenting the one-at-a-time requirement and cleanup command. No alert rule, thresholds, timeouts, or check logic were changed.
- Artifacts (path) and where the next stage used them: `observability/alert-lib.sh`, `observability/README.md`; used by `alertmanager-webhook-check.sh` and `alert-integration-check.sh`.
- Accepted / rejected / changed by hand, and why: (maintainer)
- Verification (commands and exit codes, manual checks): `./observability/alertmanager-webhook-check.sh` (exit 0); `./observability/alert-integration-check.sh` (exit 0, fast mode) run sequentially.
- Links (commit, issue, pull request, tool session): issue #112; commit TBD
