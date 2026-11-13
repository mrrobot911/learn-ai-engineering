# Observability stack

Local observability for the quiz journey. Started with Docker Compose profile `observability` (not used by Render or CI/CD).

## Start

From the repository root, with telemetry exported by the backend:

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318 OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf docker compose --profile observability up --build
```

Stop (keeps volumes):

```bash
docker compose --profile observability down
```

Never use `docker compose --profile observability down -v`: it removes volumes, including the application `db_data` volume.

## Grafana

- URL: `http://127.0.0.1:3001`
- Quiz journey dashboard: `http://127.0.0.1:3001/d/quiz-journey/quiz-journey`
- Dashboard UID: `quiz-journey`

The dashboard is provisioned from `observability/grafana/dashboards/quiz-journey.json` via `observability/grafana/provisioning/dashboards/dashboards.yaml`. Datasources (Prometheus, Loki, Tempo) are provisioned from `observability/grafana/provisioning/datasources/datasources.yaml`, including trace/log correlation (`trace_id` in logs links to Tempo, traces link back to logs).

## Smoke check

```bash
./observability/smoke-check.sh
```

Starts the profile, verifies service health and OTLP ports, generates quiz journey traffic, and queries the Grafana API to confirm dashboard `quiz-journey` exists, is provisioned, and contains the required panels.
