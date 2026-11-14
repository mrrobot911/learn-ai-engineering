#!/usr/bin/env bash
# Alertmanager -> webhook delivery in isolation (AC3). About 1 minute.
#
# Starts only the real `alertmanager` service from docker-compose.yml, so the committed
# entrypoint renders observability/alertmanager/alertmanager.yml.tmpl from ALERT_WEBHOOK_URL.
# A synthetic alert is posted to Alertmanager and a throwaway listener must receive it as JSON.
# Prometheus, the backend and the rule are not involved: see alerts_test.yml (rule logic)
# and alert-integration-check.sh (real metrics).
set -euo pipefail
# shellcheck source=observability/alert-lib.sh
source "$(dirname "${BASH_SOURCE[0]}")/alert-lib.sh"

main() {
  ac_init
  ac_up alertmanager
  ac_wait "Alertmanager healthy" 60 2 ac_am_ready

  ac_log "posting a synthetic $AC_ALERT alert"
  ac_curl -fsS -X POST -H 'Content-Type: application/json' http://alertmanager:9093/api/v2/alerts \
    -d "[{\"labels\":{\"alertname\":\"$AC_ALERT\",\"severity\":\"warning\",\"journey\":\"quiz-submission\"},\"annotations\":{\"summary\":\"$AC_SUMMARY\"}}]" \
    >/dev/null || ac_fail "Alertmanager rejected the synthetic alert"

  ac_wait "webhook received the firing notification" 60 2 ac_webhook_has firing
  ac_webhook_check firing
  ac_log "Alertmanager webhook check passed."
}

# Sourcing the file only defines main; running it executes it.
if [ "${BASH_SOURCE[0]}" = "$0" ]; then
  main
fi
