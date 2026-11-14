#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

WEBHOOK_PORT="${WEBHOOK_PORT:-9998}"
AM_PORT="${AM_PORT:-9094}"
WEBHOOK_LOG="$(mktemp)"
RENDERED_AM_CONFIG="$(mktemp)"
NETWORK="am-webhook-check-$$"
AM_CID=""
LISTENER_CID=""

cleanup() {
  set +e
  if [ -n "$AM_CID" ]; then
    docker rm -f "$AM_CID" >/dev/null 2>&1 || true
  fi
  if [ -n "$LISTENER_CID" ]; then
    docker rm -f "$LISTENER_CID" >/dev/null 2>&1 || true
  fi
  docker network rm "$NETWORK" >/dev/null 2>&1 || true
  rm -f "$WEBHOOK_LOG" "$RENDERED_AM_CONFIG" >/dev/null 2>&1 || true
}
trap cleanup EXIT

: > "$WEBHOOK_LOG"
chmod 666 "$WEBHOOK_LOG"

docker network create "$NETWORK" >/dev/null

# Ephemeral webhook listener (not committed) writing JSON payloads to a mounted log file.
LISTENER_CID="$(docker run -d --rm --network "$NETWORK" --name "am-check-listener-$$" \
  -v "$WEBHOOK_LOG:/tmp/webhook.log" \
  python:3.12-alpine python -u -c 'import http.server
path="/tmp/webhook.log"
class H(http.server.BaseHTTPRequestHandler):
    def do_POST(self):
        n=int(self.headers.get("Content-Length","0"))
        body=self.rfile.read(n).decode("utf-8","replace") if n else ""
        with open(path,"a",encoding="utf-8") as f:
            f.write(body+"\n---\n")
        self.send_response(200); self.end_headers()
    def log_message(self, fmt, *args): pass
http.server.HTTPServer(("0.0.0.0", 9999), H).serve_forever()')"

sleep 1

ALERT_WEBHOOK_URL="${ALERT_WEBHOOK_URL:-http://am-check-listener-$$:9999/alerts}"
sed "s|__ALERT_WEBHOOK_URL__|${ALERT_WEBHOOK_URL}|g" \
  observability/alertmanager/alertmanager.yml.tmpl > "$RENDERED_AM_CONFIG"
chmod 644 "$RENDERED_AM_CONFIG"

AM_CID="$(docker run -d --rm --network "$NETWORK" --name "am-check-am-$$" \
  -p "127.0.0.1:${AM_PORT}:9093" \
  -v "$RENDERED_AM_CONFIG:/etc/alertmanager/alertmanager.yml:ro" \
  prom/alertmanager:v0.27.0 \
  --config.file=/etc/alertmanager/alertmanager.yml --storage.path=/alertmanager --web.listen-address=:9093)"

for i in $(seq 1 60); do
  if curl -fsS "http://127.0.0.1:${AM_PORT}/-/healthy" >/dev/null 2>&1; then
    echo "Alertmanager: ok"
    break
  fi
  sleep 1
  if [ "$i" -eq 60 ]; then
    echo "Alertmanager: failed health check" >&2
    docker logs "$AM_CID" >&2 || true
    exit 1
  fi
done

starts_at="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
ends_at="$(date -u -d '+2 minutes' +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || python3 - <<'PY'
import datetime
print((datetime.datetime.utcnow()+datetime.timedelta(minutes=2)).strftime('%Y-%m-%dT%H:%M:%SZ'))
PY
)"

curl -fsS -X POST "http://127.0.0.1:${AM_PORT}/api/v2/alerts" \
  -H 'Content-Type: application/json' \
  -d "[{\"labels\":{\"alertname\":\"QuizSubmission5xxRateHigh\",\"severity\":\"warning\",\"journey\":\"quiz-submission\"},\"annotations\":{\"summary\":\"Visitors cannot submit a quiz\",\"description\":\"synthetic alert for webhook delivery test\"},\"startsAt\":\"${starts_at}\",\"endsAt\":\"${ends_at}\"}]" >/dev/null

for i in $(seq 1 30); do
  if grep -q 'QuizSubmission5xxRateHigh' "$WEBHOOK_LOG"; then
    break
  fi
  sleep 1
  if [ "$i" -eq 30 ]; then
    echo "webhook delivery not observed" >&2
    cat "$WEBHOOK_LOG" >&2 || true
    docker logs "$AM_CID" >&2 || true
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
        if labels.get('alertname') == 'QuizSubmission5xxRateHigh' and ann.get('summary') == 'Visitors cannot submit a quiz':
            found = True
if not found:
    raise SystemExit('webhook JSON missing QuizSubmission5xxRateHigh with expected summary')
print('Alertmanager webhook delivery: ok')
PY

echo "Alertmanager webhook check passed."
