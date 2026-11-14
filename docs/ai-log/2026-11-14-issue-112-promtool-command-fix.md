## 2026-11-14 — Fix promtool command in alert checks README (issue #112)
- Tool: opencode (iceman-free)
- Task given: Fix one wrong promtool command in `observability/README.md` (alert checks table, rule logic row); run required checks, add log, commit locally.
- Context given (files by name): `observability/README.md`, `observability/alert-integration-check.sh`, `docs/ai-workflow.md`, `docs` (excluding `docs/ai-log/**` for search).
- Result (5-8 lines): Searched for `promtool test rules` outside `docs/ai-log/**` and found the README command and a non-command mention in `observability/alert-integration-check.sh` (left unchanged). Replaced README row command to override the Prometheus image entrypoint: `docker run --rm --entrypoint promtool -v "$PWD/observability/prometheus:/work" -w /work prom/prometheus:v2.55.1 test rules alerts_test.yml`. No other README lines, scripts, alert configs, or log files changed (except this entry).
- Artifacts (path) and where the next stage used them: `observability/README.md`; this log.
- Accepted / rejected / changed by hand, and why:
- Verification (commands and exit codes, manual checks): `rg -n "promtool test rules" observability README.md docs --glob '!docs/ai-log/**'` (0). `bash -n`/edits as requested. `docker run --rm --entrypoint promtool -v "$PWD/observability/prometheus:/work" -w /work prom/prometheus:v2.55.1 test rules alerts_test.yml` (0, `SUCCESS`). `git diff --stat` after README edit showed `observability/README.md | 2 +-`.
- Links (commit, issue, pull request, tool session): issue #112 (local commit only; no push/PR/merge/close).
