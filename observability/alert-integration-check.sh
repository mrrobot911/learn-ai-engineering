#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

WEBHOOK_PORT="${WEBHOOK_PORT:-9997}"
WEBHOOK_LOG="$(mktemp)"
LISTENER_PID=""

cleanup() {
  set +e
  if [ -n "$LISTENER_PID" ]; then
    kill "$LISTENER_PID" >/dev/null 2>&1 || true
    wait "$LISTENER_PID" >/dev/null 2>&1 || true
  fi
  ENV=test FAULT_INJECT_QUIZ_SUBMISSION_5XX=false \
    docker compose --profile observability stop prometheus alertmanager otel-collector backend >/dev/null 2>&1 || true
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

export ALERT_WEBHOOK_URL="http://host.docker.internal:${WEBHOOK_PORT}/alerts"
export OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318
export OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf
export ENV=test
export FAULT_INJECT_QUIZ_SUBMISSION_5XX=false

docker compose --profile observability up -d --build db backend otel-collector prometheus alertmanager

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

wait_http "Prometheus" "http://127.0.0.1:9090/-/healthy"
wait_http "Alertmanager" "http://127.0.0.1:9093/-/healthy"

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

backend_cid="$(docker compose --profile observability ps -q backend)"
if [ -z "$backend_cid" ]; then
  echo "backend container not found" >&2
  exit 1
fi
network="$(docker inspect -f '{{range $k,$v := .NetworkSettings.Networks}}{{$k}}{{end}}' "$backend_cid")"

api_get() {
  local path="$1"
  docker run --rm --network "$network" curlimages/curl:8.10.1 -fsS "http://backend:8080${path}"
}
api_post() {
  local path="$1" body="${2-}"
  if [ -n "$body" ]; then
    docker run --rm --network "$network" curlimages/curl:8.10.1 -fsS \
      -H 'Content-Type: application/json' -d "$body" "http://backend:8080${path}"
  else
    docker run --rm --network "$network" curlimages/curl:8.10.1 -fsS -X POST \
      "http://backend:8080${path}"
  fi
}
api_post_code() {
  local path="$1" body="$2"
  docker run --rm --network "$network" curlimages/curl:8.10.1 -sS -o /dev/null -w '%{http_code}' \
    -H 'Content-Type: application/json' -d "$body" "http://backend:8080${path}"
}

# Wait for backend/API and seed data.
topic=""
for i in $(seq 1 90); do
  topics_json=$(api_get "/topics" 2>/dev/null || true)
  topic=$(printf '%s' "$topics_json" | python3 -c 'import json,sys
try:
    d=json.load(sys.stdin)
    print((d[0]["slug"] if isinstance(d,list) and d else ""))
except Exception:
    print("")' 2>/dev/null || true)
  if [ -n "$topic" ]; then
    echo "Backend/topics: ok (${topic})"
    break
  fi
  sleep 2
  if [ "$i" -eq 90 ]; then
    echo "backend/topics not ready" >&2
    docker compose --profile observability logs --tail=200 backend >&2 || true
    exit 1
  fi
done

quiz_json=$(api_post "/topics/${topic}/quiz")
submission_body=$(printf '%s' "$quiz_json" | python3 -c 'import json,sys
qs=json.load(sys.stdin)
answers=[]
for q in qs:
    opts=q.get("options") or []
    answers.append({"question_id": q["id"], "option_id": opts[0]["id"] if opts else None})
print(json.dumps({"answers": answers}))
')

echo "Phase 1: only 2xx/4xx responses on submission endpoint (must not trigger alert)"
for i in $(seq 1 20); do
  code=$(api_post_code "/topics/${topic}/quiz/submissions" "$submission_body")
  if [ "${code:0:1}" = "5" ]; then
    echo "unexpected 5xx in phase 1: $code" >&2
    exit 1
  fi
  code4=$(api_post_code "/topics/${topic}/quiz/submissions" '{}')
  if [ "${code4:0:1}" = "5" ]; then
    echo "unexpected 5xx on 4xx request in phase 1: $code4" >&2
    exit 1
  fi
  sleep 1
done
sleep 30
alerts_phase1=$(curl -fsS "http://127.0.0.1:9093/api/v2/alerts" 2>/dev/null || echo "[]")
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

echo "Phase 2: enable test-only submission fault injection and wait for webhook delivery"
export FAULT_INJECT_QUIZ_SUBMISSION_5XX=true
docker compose --profile observability up -d --no-deps --force-recreate backend
sleep 2

# Wait for backend healthy after restart.
backend_cid=""
for i in $(seq 1 90); do
  backend_cid="$(docker compose --profile observability ps -q backend)"
  if [ -n "$backend_cid" ]; then
    topics_json=$(api_get "/topics" 2>/dev/null || true)
    topic2=$(printf '%s' "$topics_json" | python3 -c 'import json,sys
try:
    d=json.load(sys.stdin)
    print((d[0]["slug"] if isinstance(d,list) and d else ""))
except Exception:
    print("")' 2>/dev/null || true)
    if [ -n "$topic2" ]; then
      echo "Backend restarted with fault injection: ok"
      break
    fi
  fi
  sleep 2
  if [ "$i" -eq 90 ]; then
    echo "backend not ready after fault-injection restart" >&2
    docker compose --profile observability logs --tail=200 backend >&2 || true
    exit 1
  fi
done
network="$(docker inspect -f '{{range $k,$v := .NetworkSettings.Networks}}{{$k}}{{end}}' "$backend_cid")"

probe=$(api_post_code "/topics/${topic}/quiz/submissions" "$submission_body")
if [ "${probe:0:1}" != "5" ]; then
  echo "expected deterministic 5xx from injected fault, got $probe" >&2
  exit 1
fi

end=$((SECONDS + 300))
while [ $SECONDS -lt $end ]; do
  code=$(api_post_code "/topics/${topic}/quiz/submissions" "$submission_body" || echo "000")
  if [ "$code" = "000" ]; then
    echo "request failed during 5xx generation" >&2
    exit 1
  fi
  sleep 2
done

alerts=""
for i in $(seq 1 120); do
  alerts=$(curl -fsS "http://127.0.0.1:9093/api/v2/alerts" 2>/dev/null || echo "[]")
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
    echo "--- prometheus rule state ---" >&2
    curl -fsS "http://127.0.0.1:9090/api/v1/rules" 2>/dev/null | python3 -m json.tool 2>/dev/null | grep -B2 -A20 'QuizSubmission5xxRateHigh' >&2 || true
    docker compose --profile observability logs --tail=200 backend otel-collector prometheus alertmanager >&2 || true
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
    for a in (obj.get('alerts') or []):
        labels = a.get('labels') or {}
        ann = a.get('annotations') or {}
        if labels.get('alertname') == 'QuizSubmission5xxRateHigh' and 'Visitors cannot submit a quiz' in ann.get('summary', ''):
            found = True
if not found:
    raise SystemExit('webhook JSON missing QuizSubmission5xxRateHigh with summary "Visitors cannot submit a quiz"')
print('Webhook JSON delivery: ok')
PY

echo "Alert integration check passed."
