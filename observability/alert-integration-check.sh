#!/usr/bin/env bash
# Real pipeline check for the quiz submission 5xx alert (AC4):
#   backend -> OTel collector -> Prometheus rule -> Alertmanager -> webhook
#
#   alert-integration-check.sh [fast|full]
#
#   fast (default, about 2-3 minutes): steps 1-4. Proves that real requests produce metrics the rule
#        selects, that 2xx/4xx traffic leaves the rule inactive and sends nothing, and that
#        5xx traffic activates the rule (state "pending", so its expression is true on live data).
#   full (about 5-6 minutes): adds step 5. Keeps the 5xx traffic running through the unchanged
#        `for: 2m` window and waits for the webhook notification (the AC4 "after the window" part).
#
# Rule logic (thresholds, minimum traffic, 4xx, other routes, resolve) is covered by
# `promtool test rules`, and delivery by alertmanager-webhook-check.sh; this script only proves
# that the real pieces are wired together. The 5xx comes from the test-only fault hook
# (ENV=test + FAULT_INJECT_QUIZ_SUBMISSION_5XX=true); nothing is stopped.
set -euo pipefail
# shellcheck source=observability/alert-lib.sh
source "$(dirname "${BASH_SOURCE[0]}")/alert-lib.sh"

MODE="${1:-fast}"
case "$MODE" in fast | full) ;; *) echo "usage: $0 [fast|full]" >&2; exit 2 ;; esac

step_start_stack() {
  ac_log "step 1: start db, backend, otel-collector, prometheus, alertmanager (fault hook off)"
  ac_up db backend otel-collector prometheus alertmanager
  ac_wait "backend healthy" 120 2 ac_backend_ready
  ac_wait "Alertmanager healthy" 60 2 ac_am_ready
  ac_wait "Prometheus scrapes the collector" 90 3 ac_collector_up
  ac_wait "rule $AC_ALERT loaded" 60 2 ac_rule_loaded
  ac_prepare_requests
}

step_no_5xx_sends_nothing() {
  ac_log "step 2: only 2xx/4xx submissions must leave the rule inactive and send nothing"
  local _
  for _ in $(seq 1 20); do
    ac_expect_code 2 "$AC_VALID_BODY" "valid submission"
    ac_expect_code 4 '{}' "invalid submission"
    sleep 0.5
  done
  ac_wait "Prometheus sees 2xx on the submission route" 90 3 ac_has_series 200
  ac_wait "Prometheus sees 4xx on the submission route" 90 3 ac_has_series 400
  ac_log "letting two scrape/evaluation cycles pass"
  sleep 40
  [ "$(ac_rule_state)" = "inactive" ] || ac_fail "rule is $(ac_rule_state) after 2xx/4xx-only traffic"
  [ "$(ac_webhook_posts)" = "0" ] || ac_fail "webhook received a notification after 2xx/4xx-only traffic"
  ac_log "ok: rule inactive, webhook silent"
}

step_enable_fault() {
  ac_log "step 3: enable the test-only fault hook and generate 5xx traffic"
  ac_set_fault true
  ac_expect_code 5 "$AC_VALID_BODY" "submission with fault hook enabled"
  ac_traffic_start
}

step_rule_activates() {
  ac_log "step 4: live 5xx metrics must activate the rule"
  ac_wait "Prometheus sees 5xx on the submission route" 90 3 ac_has_series 500
  ac_wait "rule $AC_ALERT is pending or firing" 120 3 ac_rule_active
}

step_webhook_after_window() {
  ac_log "step 5: keep traffic running through the for: 2m window and wait for the webhook"
  ac_wait "webhook received the firing notification" 360 5 ac_webhook_has firing
  ac_webhook_check firing
}

main() {
  ac_init
  step_start_stack
  step_no_5xx_sends_nothing
  step_enable_fault
  step_rule_activates
  if [ "$MODE" = "full" ]; then
    step_webhook_after_window
  else
    ac_log "fast mode: stopping before the alert window; run '$0 full' for the webhook delivery"
  fi
  ac_log "Alert integration check ($MODE) passed."
}

# Sourcing the file only defines the steps; running it executes them.
if [ "${BASH_SOURCE[0]}" = "$0" ]; then
  main
fi
