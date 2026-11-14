#!/usr/bin/env bash
# Shared helpers for the alert checks. Source it, do not run it:
#   observability/alertmanager-webhook-check.sh
#   observability/alert-integration-check.sh
#
# Everything runs in an isolated Compose project ("alert-check"): its containers, network
# and volumes are separate from a normal `docker compose --profile observability up`, so
# cleanup here (down -v) never touches the dev stack or its data.
#
# Nothing is published on the host. HTTP calls go through a small curl container on the
# Compose network, and the webhook test listener is a throwaway container whose code is
# written to a temp directory at run time (it is not committed).
#
# Environment:
#   AC_KEEP=1   keep the stack and temp files after the run, for debugging.

AC_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
AC_PROJECT="alert-check"
AC_NETWORK="${AC_PROJECT}_default"
AC_LISTENER="${AC_PROJECT}-listener"
AC_CURL="${AC_PROJECT}-curl"
AC_ROUTE="/topics/{topicSlug}/quiz/submissions"
AC_ALERT="QuizSubmission5xxRateHigh"
AC_SUMMARY="Visitors cannot submit a quiz"
AC_METRIC='{__name__=~"http_server_request_count(_total)?"'

export COMPOSE_PROJECT_NAME="$AC_PROJECT"
# The webhook URL is passed to the stack through the same environment variable as in real use.
export ALERT_WEBHOOK_URL="http://${AC_LISTENER}:9999/alerts"

AC_COMPOSE=(docker compose
  -f "$AC_ROOT/docker-compose.yml"
  -f "$AC_ROOT/observability/alert-test.compose.yml"
  --profile observability)

AC_TMP=""
AC_TRAFFIC_PID=""
AC_TOPIC=""
AC_VALID_BODY=""

ac_log() { printf '[alert-check] %s\n' "$*"; }

ac_fail() {
  printf '[alert-check] FAIL: %s\n' "$*" >&2
  ac_diagnose >&2 || true
  exit 1
}

ac_diagnose() {
  echo "--- diagnostics ---"
  if [ -n "$AC_TMP" ] && [ -s "$AC_TMP/webhook.log" ]; then
    echo "webhook log:"; cat "$AC_TMP/webhook.log"
  else
    echo "webhook log: empty"
  fi
  if docker inspect "$AC_CURL" >/dev/null 2>&1; then
    echo "rule state: $(ac_rule_state 2>/dev/null || echo unavailable)"
    echo "alertmanager alerts: $(ac_curl -fsS http://alertmanager:9093/api/v2/alerts 2>/dev/null || echo unavailable)"
  fi
  "${AC_COMPOSE[@]}" ps 2>/dev/null || true
  "${AC_COMPOSE[@]}" logs --tail=30 backend otel-collector prometheus alertmanager 2>/dev/null || true
}

# ---- lifecycle -------------------------------------------------------------------------

ac_init() {
  local existing
  existing="$(${AC_COMPOSE[@]} ps -aq 2>/dev/null || true)"
  if [ -n "$existing" ] || docker container inspect "$AC_LISTENER" "$AC_CURL" >/dev/null 2>&1; then
    printf '[alert-check] FAIL: alert-check resources already exist. Run the checks one at a time and clean up with: docker compose -p alert-check down -v; docker rm -f alert-check-listener alert-check-curl\n' >&2
    exit 1
  fi
  AC_TMP="$(mktemp -d)"
  : > "$AC_TMP/webhook.log"
  chmod 777 "$AC_TMP"
  chmod 666 "$AC_TMP/webhook.log"
  trap ac_cleanup EXIT
  ac_write_listener
}

ac_cleanup() {
  set +e
  ac_traffic_stop
  if [ "${AC_KEEP:-}" = "1" ]; then
    ac_log "AC_KEEP=1: stack left running (project $AC_PROJECT); remove with: docker compose -p $AC_PROJECT down -v"
    return
  fi
  docker rm -f "$AC_LISTENER" "$AC_CURL" >/dev/null 2>&1
  "${AC_COMPOSE[@]}" down -v --remove-orphans >/dev/null 2>&1
  [ -n "$AC_TMP" ] && rm -rf "$AC_TMP"
}

# ac_up SERVICE... : start only the given services, then the listener and the curl container.
ac_up() {
  "${AC_COMPOSE[@]}" up -d --build "$@" || ac_fail "docker compose up failed"
  docker rm -f "$AC_LISTENER" "$AC_CURL" >/dev/null 2>&1 || true
  docker run -d --name "$AC_LISTENER" --network "$AC_NETWORK" -v "$AC_TMP:/w" \
    python:3.12-alpine python -u /w/listener.py /w/webhook.log 9999 >/dev/null \
    || ac_fail "could not start the webhook listener"
  docker run -d --name "$AC_CURL" --network "$AC_NETWORK" --entrypoint sleep \
    curlimages/curl:8.10.1 86400 >/dev/null || ac_fail "could not start the curl container"
}

# ac_wait DESCRIPTION TIMEOUT_SECONDS INTERVAL_SECONDS COMMAND... : poll until COMMAND succeeds.
ac_wait() {
  local desc="$1" timeout="$2" interval="$3"
  shift 3
  local end=$((SECONDS + timeout))
  until "$@" >/dev/null 2>&1; do
    [ "$SECONDS" -lt "$end" ] || ac_fail "timed out after ${timeout}s waiting for: $desc"
    sleep "$interval"
  done
  ac_log "ok: $desc"
}

# ---- HTTP inside the Compose network ---------------------------------------------------

ac_curl() { docker exec "$AC_CURL" curl -sS "$@"; }

ac_am_ready() { ac_curl -fsS http://alertmanager:9093/-/healthy; }
ac_backend_ready() { ac_curl -fsS http://backend:8080/health; }

# ---- Prometheus ------------------------------------------------------------------------

# ac_series_count PROMQL : number of series returned by an instant query.
ac_series_count() {
  ac_curl -fsSG --data-urlencode "query=$1" http://prometheus:9090/api/v1/query \
    | python3 -c 'import json,sys; print(len(json.load(sys.stdin)["data"]["result"]))'
}

# ac_has_series STATUS : the rule selector sees a live series for the submission route and STATUS.
ac_has_series() {
  [ "$(ac_series_count "${AC_METRIC},route=\"${AC_ROUTE}\",status=\"$1\"}")" -gt 0 ]
}

ac_collector_up() { [ "$(ac_series_count 'up{job="otel-collector"} == 1')" -gt 0 ]; }

# ac_rule_state : inactive | pending | firing | missing
ac_rule_state() {
  ac_curl -fsS http://prometheus:9090/api/v1/rules | python3 -c '
import json, sys
for group in json.load(sys.stdin)["data"]["groups"]:
    for rule in group["rules"]:
        if rule.get("name") == "'"$AC_ALERT"'":
            print(rule.get("state", "unknown"))
            sys.exit(0)
print("missing")'
}

ac_rule_loaded() { [ "$(ac_rule_state)" != "missing" ]; }
ac_rule_active() { case "$(ac_rule_state)" in pending | firing) return 0 ;; *) return 1 ;; esac; }

# ---- webhook listener ------------------------------------------------------------------

# Writes the listener to the temp directory: one JSON line per POST (path, content type, body).
ac_write_listener() {
  cat > "$AC_TMP/listener.py" <<'PY'
import http.server
import json
import sys

log_path, port = sys.argv[1], int(sys.argv[2])


class Handler(http.server.BaseHTTPRequestHandler):
    def do_POST(self):
        length = int(self.headers.get("Content-Length", "0"))
        body = self.rfile.read(length).decode("utf-8", "replace") if length else ""
        record = {"path": self.path, "content_type": self.headers.get("Content-Type", ""), "body": body}
        with open(log_path, "a", encoding="utf-8") as handle:
            handle.write(json.dumps(record) + "\n")
        self.send_response(200)
        self.end_headers()

    def log_message(self, *args):
        pass


http.server.ThreadingHTTPServer(("0.0.0.0", port), Handler).serve_forever()
PY
}

ac_webhook_posts() { wc -l < "$AC_TMP/webhook.log" | tr -d ' '; }

# ac_webhook_has STATUS : a notification with this status (firing|resolved) for the alert arrived.
ac_webhook_has() { ac_webhook_check "$1" >/dev/null 2>&1; }

# ac_webhook_check STATUS : exits 0 and prints details when a JSON notification with the given
# status contains the alert with the expected name and the user-impact summary.
ac_webhook_check() {
  python3 - "$AC_TMP/webhook.log" "$1" "$AC_ALERT" "$AC_SUMMARY" <<'PY'
import json
import sys

log_path, want_status, alertname, summary = sys.argv[1:5]
for line in open(log_path, encoding="utf-8"):
    record = json.loads(line)
    if not record["content_type"].startswith("application/json"):
        continue
    try:
        payload = json.loads(record["body"])
    except ValueError:
        continue
    if payload.get("status") != want_status:
        continue
    for alert in payload.get("alerts", []):
        if alert.get("labels", {}).get("alertname") == alertname and \
                alert.get("annotations", {}).get("summary") == summary:
            print("webhook JSON ok: status=%s alertname=%s summary=%r" % (want_status, alertname, summary))
            sys.exit(0)
sys.exit(1)
PY
}

# ---- test traffic ----------------------------------------------------------------------

# Picks a topic that has quiz questions and builds a valid submission body for it.
ac_prepare_requests() {
  local topics slug quiz body
  topics="$(ac_curl -fsS http://backend:8080/topics)" || ac_fail "GET /topics failed"
  for slug in $(printf '%s' "$topics" | python3 -c 'import json,sys; print(" ".join(t["slug"] for t in json.load(sys.stdin)))'); do
    quiz="$(ac_curl -fsS -X POST "http://backend:8080/topics/${slug}/quiz" 2>/dev/null || true)"
    body="$(printf '%s' "$quiz" | python3 -c '
import json, sys
try:
    questions = json.load(sys.stdin)
except ValueError:
    sys.exit(0)
answers = [{"question_id": q["id"], "option_id": q["options"][0]["id"]} for q in questions if q.get("options")]
if answers:
    print(json.dumps({"answers": answers}))')"
    if [ -n "$body" ]; then
      AC_TOPIC="$slug"
      AC_VALID_BODY="$body"
      ac_log "ok: using topic '$AC_TOPIC' for submissions"
      return 0
    fi
  done
  ac_fail "no topic with quiz questions found"
}

# ac_submit BODY : POST a submission and print the HTTP status code.
ac_submit() {
  ac_curl -o /dev/null -w '%{http_code}' -H 'Content-Type: application/json' -d "$1" \
    "http://backend:8080/topics/${AC_TOPIC}/quiz/submissions"
}

# ac_expect_code EXPECTED_FIRST_DIGIT BODY DESCRIPTION : one request, strict about the status class.
ac_expect_code() {
  local code
  code="$(ac_submit "$2")" || ac_fail "request failed: $3"
  [ "${code:0:1}" = "$1" ] || ac_fail "$3: expected ${1}xx, got $code"
}

ac_traffic_start() {
  ( while :; do ac_submit "$AC_VALID_BODY" >/dev/null 2>&1 || true; sleep 0.5; done ) &
  AC_TRAFFIC_PID=$!
}

ac_traffic_stop() {
  if [ -n "$AC_TRAFFIC_PID" ]; then
    kill "$AC_TRAFFIC_PID" >/dev/null 2>&1 || true
    wait "$AC_TRAFFIC_PID" >/dev/null 2>&1 || true
    AC_TRAFFIC_PID=""
  fi
}

# ac_set_fault true|false : restart only the backend with the test-only fault hook on or off.
ac_set_fault() {
  FAULT_INJECT_QUIZ_SUBMISSION_5XX="$1" "${AC_COMPOSE[@]}" up -d --no-deps --force-recreate backend \
    || ac_fail "could not restart the backend"
  ac_wait "backend healthy (fault injection: $1)" 90 2 ac_backend_ready
}
