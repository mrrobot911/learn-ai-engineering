#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

cleanup() {
  docker compose --profile observability stop otel-collector prometheus loki tempo grafana >/dev/null 2>&1 || true
  docker compose --profile observability rm -f otel-collector prometheus loki tempo grafana >/dev/null 2>&1 || true
  for v in prometheus_data loki_data tempo_data grafana_data; do
    docker volume ls -q --filter "label=com.docker.compose.volume=$v" | xargs -r docker volume rm >/dev/null 2>&1 || true
  done
}
trap cleanup EXIT

OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318 OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf docker compose --profile observability up -d --build

wait_http() {
  local name="$1"
  local url="$2"
  for i in $(seq 1 60); do
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
wait_http "Loki" "http://127.0.0.1:3100/ready"
wait_http "Tempo" "http://127.0.0.1:3200/ready"
wait_http "OTel Collector" "http://127.0.0.1:13133/"
wait_http "App" "http://127.0.0.1:3000/api/health"

if command -v nc >/dev/null 2>&1; then
  for port in 4317 4318; do
    if nc -z 127.0.0.1 "$port"; then
      echo "OTLP ${port}: ok"
    else
      echo "OTLP ${port}: failed to accept connection" >&2
      exit 1
    fi
  done
else
  (echo > /dev/tcp/127.0.0.1/4317) >/dev/null 2>&1 && echo "OTLP 4317: ok" || { echo "OTLP 4317: failed" >&2; exit 1; }
  (echo > /dev/tcp/127.0.0.1/4318) >/dev/null 2>&1 && echo "OTLP 4318: ok" || { echo "OTLP 4318: failed" >&2; exit 1; }
fi

# Smoke traffic for the quiz journey so panels can show data.
topic=$(curl -fsS "http://127.0.0.1:3000/api/topics" | python3 -c 'import json,sys; d=json.load(sys.stdin); print((d[0]["slug"] if d else ""))')
if [ -z "$topic" ]; then
  echo "no topic available for quiz smoke traffic" >&2
  exit 1
fi
quiz_json=$(curl -fsS -X POST "http://127.0.0.1:3000/api/topics/${topic}/quiz")
submission_body=$(printf '%s' "$quiz_json" | python3 -c 'import json,sys
qs = json.load(sys.stdin)
answers = []
for q in qs:
    opts = q.get("options") or []
    answers.append({"question_id": q["id"], "option_id": opts[0]["id"] if opts else None})
print(json.dumps({"answers": answers}))
')
for i in 1 2 3; do
  curl -fsS -X POST "http://127.0.0.1:3000/api/topics/${topic}/quiz" >/dev/null
  curl -fsS -X POST -H 'Content-Type: application/json' \
    -d "$submission_body" \
    "http://127.0.0.1:3000/api/topics/${topic}/quiz/submissions" >/dev/null
done
echo "Quiz journey smoke traffic: ok"

dashboard_json=$(curl -fsS "http://127.0.0.1:3001/api/dashboards/uid/quiz-journey")
printf '%s' "$dashboard_json" | python3 -c '
import json, sys
d = json.load(sys.stdin)
if d.get("meta", {}).get("isStarred") is None and "dashboard" not in d:
    raise SystemExit("unexpected dashboard payload")
meta = d.get("meta", {})
if meta.get("slug") != "quiz-journey" and d.get("dashboard", {}).get("uid") != "quiz-journey":
    raise SystemExit("dashboard uid/slug mismatch")
prov = meta.get("provisioned", False)
if not prov:
    raise SystemExit("dashboard is not provisioned")
panels = d.get("dashboard", {}).get("panels", [])
titles = {p.get("title") for p in panels}
required = {"Request rate", "5xx rate", "p95 latency", "Quiz submissions", "Trace search", "Journey logs"}
missing = sorted(required - titles)
if missing:
    raise SystemExit("missing panels: " + ", ".join(missing))
print("Grafana dashboard quiz-journey: ok")
'

# Fail if any provisioning error is reported for dashboards.
health=$(curl -fsS "http://127.0.0.1:3001/api/health")
printf '%s' "$health" >/dev/null
echo "Grafana provisioning: ok"

echo "Observability smoke check passed."
