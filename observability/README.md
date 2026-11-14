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

## Alert checks

The alert testing architecture is split into three focused checks:

1. **Prometheus rule tests (deterministic)**

   ```bash
   docker run --rm -v "$PWD/observability/prometheus:/work" -w /work prom/prometheus:v2.55.1 promtool test rules alerts_test.yml
   ```

   Validates `QuizSubmission5xxRateHigh` rule semantics from the committed `alerts.yml`: firing above threshold with minimum traffic, resolve, and no firing on 4xx-only traffic, other endpoints, or no/insufficient traffic.

2. **Alertmanager webhook delivery (isolated)**

   ```bash
   ./observability/alertmanager-webhook-check.sh
   ```

   Renders `alertmanager.yml.tmpl` with `ALERT_WEBHOOK_URL`, starts a standalone Alertmanager and an ephemeral local webhook listener (not committed), posts a synthetic alert via the Alertmanager API, and asserts JSON webhook delivery with the expected user-impact summary.

3. **Minimal endpoint-to-webhook integration (AC4)**

   ```bash
   ./observability/alert-integration-check.sh
   ```

   Starts only the minimal pipeline (`db`, `backend`, `otel-collector`, `prometheus`, `alertmanager`) with an ephemeral webhook listener, confirms 2xx/4xx-only submission traffic sends no alert, then induces deterministic 5xx responses on `POST /topics/{topicSlug}/quiz/submissions` and observes webhook delivery after the alert window. Requires Docker and several minutes (alert `rate`/`for` windows are unchanged).

### Test-only fault injection for the integration check

The integration check uses a narrowly scoped, startup-time fault injection hook for `POST /topics/{topicSlug}/quiz/submissions` so it can produce deterministic 5xx responses without stopping PostgreSQL or other shared dependencies.

- Environment variable: `FAULT_INJECT_QUIZ_SUBMISSION_5XX=true`
- **Disabled by default.** Normal behavior is unchanged when unset/false.
- **Designated test environment only:** the backend enables it only when `ENV=test`; startup rejects attempts to enable it in `development` or `production`.
- **Not a runtime/public control:** there is no HTTP endpoint or unauthenticated toggle; it is read only from environment at startup.
- **Scope:** affects only `POST /topics/{topicSlug}/quiz/submissions` and returns a deterministic 500 `internal_error` response.
- **Render safety:** `render.yaml` sets `ENV=production` and does not define this variable; production startup rejects enabling it, so it cannot be accidentally enabled in the deployed environment.

`alert-integration-check.sh` enables it by restarting the backend with `ENV=test` and `FAULT_INJECT_QUIZ_SUBMISSION_5XX=true` for the 5xx phase, and disables it by starting/restarting without it (or with it unset/false). Recreate/restart the backend process to change the setting.
