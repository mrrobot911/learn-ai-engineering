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

The alert is verified in four independent steps, from fastest to slowest. Each one fails on its own link of the chain `backend -> OTel collector -> Prometheus rule -> Alertmanager -> webhook`, so a failure points at one place.

| Step | Command | Time | Proves |
| --- | --- | --- | --- |
| 1. Rule logic | `docker run --rm --entrypoint promtool -v "$PWD/observability/prometheus:/work" -w /work prom/prometheus:v2.55.1 test rules alerts_test.yml` | seconds | AC1, AC2: fires above 5% with enough traffic and `for: 2m`, resolves, ignores 4xx, other routes, low and no traffic |
| 2. Webhook delivery | `./observability/alertmanager-webhook-check.sh` | ~1 min | AC3: the real `alertmanager` service renders its config from `ALERT_WEBHOOK_URL` and a throwaway listener receives a JSON notification |
| 3. Real pipeline, fast | `./observability/alert-integration-check.sh` | ~2 min | AC4, first half: real requests produce metrics the rule selects; 2xx/4xx traffic keeps the rule inactive and sends nothing; 5xx traffic activates the rule |
| 4. Real pipeline, full | `./observability/alert-integration-check.sh full` | ~4-5 min | AC4: 5xx traffic through the unchanged `for: 2m` window ends in a webhook notification |

Steps 2-4 need Docker. They run in an isolated Compose project named `alert-check` (own containers, network and volumes), so they do not touch a running dev stack and need no free host ports. They start only `db`, `backend`, `otel-collector`, `prometheus` and `alertmanager` (step 2: only `alertmanager`). Nothing is published on the host; requests go through a short-lived curl container and the webhook listener is a throwaway container whose code is written to a temp directory at run time (not committed). Add `AC_KEEP=1` to leave the stack running for debugging; remove it with `docker compose -p alert-check down -v`.

Step 3 stops as soon as the rule becomes `pending`, which already proves that its expression is true on live data, so it does not wait for the `for: 2m` window. Run step 4 once before handing the work over; it is the only slow one. Times exclude the first run's image build and pulls.

The checks use `observability/alert-test.compose.yml` on top of `docker-compose.yml`. It sets `ENV=test`, wires the backend to the collector, and sets `OTEL_METRIC_EXPORT_INTERVAL=5000` because the SDK exports metrics only every 60 s by default. Alert thresholds, windows, minimum traffic and `for` are not changed.

### Test-only fault injection

Step 3 and 4 need real 5xx responses on `POST /topics/{topicSlug}/quiz/submissions` without stopping PostgreSQL, so the backend has a startup-time fault hook.

- Environment variable: `FAULT_INJECT_QUIZ_SUBMISSION_5XX=true`
- **Disabled by default.** Normal behavior is unchanged when unset or not `true`.
- **Designated test environment only:** it is enabled only when `ENV=test`; startup fails if it is requested in `development` or `production`.
- **Not a runtime or public control:** there is no HTTP endpoint or toggle; it is read once from the environment at startup.
- **Scope:** only `POST /topics/{topicSlug}/quiz/submissions` is affected and it returns a deterministic 500 `internal_error`.
- **Render safety:** `render.yaml` sets `ENV=production` and does not define this variable; production startup rejects it, so it cannot be enabled in the deployed environment.

The integration check enables it by recreating only the backend container with `FAULT_INJECT_QUIZ_SUBMISSION_5XX=true`, and runs without it before that. To disable it, recreate the backend without the variable.
