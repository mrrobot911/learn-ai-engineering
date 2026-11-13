#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

WEBHOOK_PORT="${WEBHOOK_PORT:-9999}"
WEBHOOK_LOG="$(mktemp)"
LISTENER_PID=""

cleanup() {
  set +e
  if [ -n "$LISTENER_PID" ]; then
    kill "$LISTENER_PID" >/dev/null 2>&1 || true
    wait "$LISTENER_PID" >/dev/null 2>&1 || true
  fi
  docker compose --profile observability start db >/dev/null 2>&1 || true
  docker compose --profile observability down >/dev/null 2>&1 || true
  rm -f "$WEBHOOK_LOG" "${WEBHOOK_LOG}.py" >/dev/null 2>&1 || true
}
trap cleanup EXIT

cat > "${WEBHOOK_LOG}.py" <<'PY'
import http.server, sys
path = sys.argv[1]
class H(http.server.BaseHTTPRequestHandler):
    def do_POST(self):
        n = int(self.headers.get('Content-Length', '0'))
        body = self.rfile.read(n).decode('utf-8', 'replace') if n else ''
        with open(path, 'a', encoding='utf-8') as f:
            f.write(body + "\n---\n")
        self.send_response(200)
        self.end_headers()
    def log_message(self, fmt, *args):
        pass
http.server.HTTPServer(('0.0.0.0', int(sys.argv[2])), H).serve_forever()
PY
: > "$WEBHOOK_LOG"
python3 "${WEBHOOK_LOG}.py" "$WEBHOOK_LOG" "$WEBHOOK_PORT" &
LISTENER_PID=$!
sleep 1

ALERT_WEBHOOK_URL="http://host.docker.internal:${WEBHOOK_PORT}/alerts" \
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318 \
OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf \
docker compose --profile observability up -d --build

wait_http() {
  local name="$1" url="$2"
  for i in $(seq 1 90); do
    if curl -fsS "$url" >/dev/null 2>&1; then
      echo "$name: ok"
      return 0
    fi
    sleep 2
  done
  echo "$name: failed health check ($url)" >&2
  return 1
}

wait_http "Grafana" "http://127.0.0.1:3001/api/health"
wait_http "Prometheus" "http://127.0.0.1:9090/-/healthy"
wait_http "Alertmanager" "http://127.0.0.1:9093/-/healthy"
wait_http "App" "http://127.0.0.1:3000/api/health"

# Wait for Prometheus to have loaded the alert rule.
for i in $(seq 1 60); do
  rules=$(curl -fsS "http://127.0.0.1:9090/api/v1/rules" || true)
  if printf '%s' "$rules" | grep -q 'QuizSubmission5xxRateHigh'; then
    echo "Prometheus rule QuizSubmission5xxRateHigh: ok"
    break
  fi
  sleep 2
  if [ "$i" -eq 60 ]; then
    echo "alert rule not loaded in Prometheus" >&2
    exit 1
  fi
done

topic=""
for i in $(seq 1 60); do
  topics_json=$(curl -fsS "http://127.0.0.1:3000/api/topics" 2>/dev/null || true)
  topic=$(printf '%s' "$topics_json" | python3 -c 'import json,sys
try:
    d=json.load(sys.stdin)
    print((d[0]["slug"] if isinstance(d,list) and d else ""))
except Exception:
    print("")' 2>/dev/null || true)
  if [ -n "$topic" ]; then
    break
  fi
  sleep 2
  if [ "$i" -eq 60 ]; then
    echo "no topic available from /api/topics" >&2
    printf '%s\n' "$topics_json" >&2
    exit 1
  fi
done

quiz_json=""
for i in $(seq 1 30); do
  quiz_json=$(curl -fsS -X POST "http://127.0.0.1:3000/api/topics/${topic}/quiz" 2>/dev/null || true)
  if [ -n "$quiz_json" ]; then
    break
  fi
  sleep 2
  if [ "$i" -eq 30 ]; then
    echo "failed to start quiz for topic ${topic}" >&2
    exit 1
  fi
done
submission_body=$(printf '%s' "$quiz_json" | python3 -c 'import json,sys
qs = json.load(sys.stdin)
answers = []
for q in qs:
    opts = q.get("options") or []
    answers.append({"question_id": q["id"], "option_id": opts[0]["id"] if opts else None})
print(json.dumps({"answers": answers}))
')

echo "Phase 1: only 2xx/4xx responses (must not trigger alert)"
for i in $(seq 1 20); do
  code=$(curl -sS -o /dev/null -w '%{http_code}' -X POST -H 'Content-Type: application/json' \
    -d "$submission_body" "http://127.0.0.1:3000/api/topics/${topic}/quiz/submissions")
  if [ "${code:0:1}" = "5" ]; then
    echo "unexpected 5xx in phase 1: $code" >&2
    exit 1
  fi
  # 4xx on same endpoint (invalid body) - must not count as 5xx/firing cause.
  code4=$(curl -sS -o /dev/null -w '%{http_code}' -X POST -H 'Content-Type: application/json' \
    -d '{}' "http://127.0.0.1:3000/api/topics/${topic}/quiz/submissions")
  if [ "${code4:0:1}" = "5" ]; then
    echo "unexpected 5xx on 4xx request in phase 1: $code4" >&2
    exit 1
  fi
  sleep 1
done
sleep 30
alerts_phase1=$(curl -fsS "http://127.0.0.1:9093/api/v2/alerts" || true)
if printf '%s' "$alerts_phase1" | grep -q 'QuizSubmission5xxRateHigh'; then
  echo "alert fired during 2xx/4xx-only phase" >&2
  printf '%s\n' "$alerts_phase1" >&2
  exit 1
fi
if grep -q 'QuizSubmission5xxRateHigh' "$WEBHOOK_LOG"; then
  echo "webhook received alert during 2xx/4xx-only phase" >&2
  exit 1
fi
echo "2xx/4xx-only phase: no alert"

echo "Phase 2: induce 5xx responses and wait for webhook delivery after window"
docker compose --profile observability stop db >/dev/null

# Generate sustained 5xx traffic on POST /topics/{topicSlug}/quiz/submissions.
# Alert: >5% 5xx for 2m, minimum 5 requests in 2m. Run for ~5 minutes.
end=$((SECONDS + 300))
while [ $SECONDS -lt $end ]; do
  curl -sS -o /dev/null -X POST -H 'Content-Type: application/json' \
    -d "$submission_body" "http://127.0.0.1:3000/api/topics/${topic}/quiz/submissions" || true
  sleep 2
done

# Wait for firing alert and webhook delivery.
for i in $(seq 1 120); do
  alerts=$(curl -fsS "http://127.0.0.1:9093/api/v2/alerts" || true)
  if printf '%s' "$alerts" | grep -q 'QuizSubmission5xxRateHigh' && grep -q 'QuizSubmission5xxRateHigh' "$WEBHOOK_LOG"; then
    echo "Alertmanager/webhook observed firing QuizSubmission5xxRateHigh"
    break
  fi
  sleep 5
  if [ "$i" -eq 120 ]; then
    echo "alert or webhook delivery not observed after 5xx phase" >&2
    echo "--- alerts ---" >&2
    printf '%s\n' "$alerts" >&2
    echo "--- webhook ---" >&2
    cat "$WEBHOOK_LOG" >&2 || true
    exit 1
  fi
done

python3 - "$WEBHOOK_LOG" <<'PY'
import json, sys
path = sys.argv[1]
text = open(path, encoding='utf-8').read()
chunks = [c.strip() for c in text.split('\n---\n') if c.strip()]
found = False
for c in chunks:
    try:
        obj = json.loads(c)
    except Exception:
        continue
    alerts = obj.get('alerts') or []
    for a in alerts:
        labels = a.get('labels') or {}
        ann = a.get('annotations') or {}
        if labels.get('alertname') == 'QuizSubmission5xxRateHigh' and 'Visitors cannot submit a quiz' in ann.get('summary', ''):
            found = True
if not found:
    raise SystemExit('webhook JSON missing QuizSubmission5xxRateHigh with summary "Visitors cannot submit a quiz"')
print('Webhook JSON delivery: ok')
PY

# Restore DB and verify alert can resolve (best-effort; not required to block pass if slow).
docker compose --profile observability start db >/dev/null
wait_http "App" "http://127.0.0.1:3000/api/health" || true

echo "Alert check passed."
