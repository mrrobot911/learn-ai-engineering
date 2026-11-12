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

docker compose --profile observability up -d --build

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

echo "Observability smoke check passed."
