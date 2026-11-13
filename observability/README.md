# Observability stack

Local observability for the quiz journey. Started with Docker Compose profile `observability` (not used by Render or CI/CD).

## Start

From the repository root, with telemetry exported by the backend:

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318 OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf docker compose --profile observability up --build
```

To deliver quiz submission 5xx alerts to a webhook, set `ALERT_WEBHOOK_URL` (documented in `.env.example`), for example a local test listener:

```bash
ALERT_WEBHOOK_URL=http://host.docker.internal:9999/alerts OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318 OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf docker compose --profile observability up --build
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

## Alert: quiz submission 5xx rate

- Rule file (provisioned): `observability/prometheus/alerts.yml`
- Alert name: `QuizSubmission5xxRateHigh`
- Summary (user impact): `Visitors cannot submit a quiz`
- Signal: 5xx rate of `POST /topics/{topicSlug}/quiz/submissions` (`route="/topics/{topicSlug}/quiz/submissions"`, `status=~"5.."`)
- Threshold: `> 5%` 5xx responses
- Window: `2m` rate, condition held `for: 2m`
- Minimum traffic: `>= 5` requests to the endpoint in `2m` (`increase(...[2m]) >= 5`)
- Does not fire: on 4xx responses, on other endpoints, or when there is no traffic
- Delivery: Alertmanager webhook at `ALERT_WEBHOOK_URL` (environment variable), JSON notification with `send_resolved: true`
- Scope: local `observability` profile only; not deployed

## Smoke check

```bash
./observability/smoke-check.sh
```

Starts the profile, verifies service health and OTLP ports, generates quiz journey traffic, and queries the Grafana API to confirm dashboard `quiz-journey` exists, is provisioned, and contains the required panels.

## Alert check

```bash
./observability/alert-check.sh
```

Starts the profile with a local webhook listener (test stub, not committed), verifies the `QuizSubmission5xxRateHigh` rule is loaded, confirms 2xx/4xx-only traffic sends no alert, then induces 5xx responses on `POST /topics/{topicSlug}/quiz/submissions` and waits for the webhook delivery after the alert window. Requires Docker and several minutes to complete.
