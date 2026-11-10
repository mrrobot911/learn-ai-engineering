#!/usr/bin/env bash
# Frontend integration tests (T31a / #85).
# Starts an isolated test database and the backend, applies migrations and seed,
# inserts the small fixture data the tests need, then runs the frontend
# integration suite. Never points at the development database.
#
# Usage (from the repository root):
#   bash scripts/frontend-integration.sh
#
# Local mode (default): uses Docker Compose (db service) and runs backend in a container.
# External mode (TEST_INTEGRATION_EXTERNAL=1): uses an externally provided PostgreSQL
# (e.g. CI service) via POSTGRES_HOST/PORT/USER/PASSWORD and DATABASE_URL/TEST_DATABASE_URL,
# runs the backend locally with `go run`, and requires `psql` and `pg_isready` on PATH.
set -eu

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

DB_USER="${POSTGRES_USER:-learn}"
DB_PASSWORD="${POSTGRES_PASSWORD:-learn}"
DB_HOST="${POSTGRES_HOST:-localhost}"
DB_PORT="${POSTGRES_PORT:-5432}"
TEST_DB="${FRONTEND_TEST_DB:-learn_ai_fe_test}"
TEST_API_PORT="${TEST_API_PORT:-18080}"
API_CONTAINER="${FE_TEST_API_CONTAINER:-learn-ai-fe-test-api}"
DB_ADMIN="${POSTGRES_ADMIN_USER:-$DB_USER}"

if [ "${TEST_INTEGRATION_EXTERNAL:-}" = "1" ]; then
  BACKEND_PID=""
  cleanup() {
    if [ -n "$BACKEND_PID" ] && kill -0 "$BACKEND_PID" >/dev/null 2>&1; then
      kill "$BACKEND_PID" >/dev/null 2>&1 || true
      wait "$BACKEND_PID" >/dev/null 2>&1 || true
    fi
  }
  trap cleanup EXIT

  export PGPASSWORD="$DB_PASSWORD"
  ADMIN_PGPASSWORD="${POSTGRES_ADMIN_PASSWORD:-$DB_PASSWORD}"

  echo "==> Waiting for external database at $DB_HOST:$DB_PORT"
  for i in $(seq 1 30); do
    if PGPASSWORD="$ADMIN_PGPASSWORD" pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_ADMIN" -d postgres >/dev/null 2>&1; then
      break
    fi
    if [ "$i" -eq 30 ]; then
      echo "external database did not become ready" >&2
      exit 1
    fi
    sleep 1
  done

  echo "==> Resetting isolated frontend integration database: $TEST_DB (never dev database)"
  PGPASSWORD="$ADMIN_PGPASSWORD" psql -v ON_ERROR_STOP=1 -h "$DB_HOST" -p "$DB_PORT" -U "$DB_ADMIN" -d postgres \
    -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '$TEST_DB' AND pid <> pg_backend_pid();" \
    >/dev/null
  PGPASSWORD="$ADMIN_PGPASSWORD" psql -v ON_ERROR_STOP=1 -h "$DB_HOST" -p "$DB_PORT" -U "$DB_ADMIN" -d postgres \
    -c "DROP DATABASE IF EXISTS \"$TEST_DB\";"
  PGPASSWORD="$ADMIN_PGPASSWORD" psql -v ON_ERROR_STOP=1 -h "$DB_HOST" -p "$DB_PORT" -U "$DB_ADMIN" -d postgres \
    -c "CREATE DATABASE \"$TEST_DB\" OWNER \"$DB_USER\";"

  DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${TEST_DB}?sslmode=disable"
  export DATABASE_URL
  export HOST=127.0.0.1
  export PORT="$TEST_API_PORT"
  export ENV=test
  export SEED_MAINTAINER_EMAIL="${SEED_MAINTAINER_EMAIL:-}"
  export SEED_MAINTAINER_PASSWORD="${SEED_MAINTAINER_PASSWORD:-}"

  echo "==> Starting test backend locally on port $TEST_API_PORT"
  (
    cd backend
    go run ./cmd/api
  ) >/tmp/fe-integration-backend.log 2>&1 &
  BACKEND_PID=$!

  echo "==> Waiting for test backend (migrations + seed run on startup)"
  for i in $(seq 1 90); do
    if curl -fsS "http://127.0.0.1:${TEST_API_PORT}/health" >/dev/null 2>&1; then
      break
    fi
    if ! kill -0 "$BACKEND_PID" >/dev/null 2>&1; then
      echo "test backend exited early; logs:" >&2
      cat /tmp/fe-integration-backend.log >&2 || true
      exit 1
    fi
    if [ "$i" -eq 90 ]; then
      echo "test backend did not become healthy; logs:" >&2
      cat /tmp/fe-integration-backend.log >&2 || true
      exit 1
    fi
    sleep 1
  done

  echo "==> Inserting fixture topic for 'fewer than 5 questions' (2 published questions)"
  psql -v ON_ERROR_STOP=1 -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$TEST_DB" <<'SQL'
INSERT INTO topics (id, slug, title, description, position, status)
VALUES ('topic-few-fixture', 'quiz-few', 'Quiz Few Fixture', 'Fixture with two questions.', 999, 'published')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO questions (id, topic_id, text, explanation, status)
VALUES
  ('q-few-1', 'topic-few-fixture', 'Fixture question one?', 'Explanation one.', 'published'),
  ('q-few-2', 'topic-few-fixture', 'Fixture question two?', 'Explanation two.', 'published')
ON CONFLICT (id) DO NOTHING;

INSERT INTO answer_options (id, question_id, text, position, is_correct, status)
VALUES
  ('o-few-1a', 'q-few-1', 'Option A1', 1, true, 'published'),
  ('o-few-1b', 'q-few-1', 'Option B1', 2, false, 'published'),
  ('o-few-2a', 'q-few-2', 'Option A2', 1, false, 'published'),
  ('o-few-2b', 'q-few-2', 'Option B2', 2, true, 'published')
ON CONFLICT (id) DO NOTHING;
SQL

  echo "==> Running frontend integration tests"
  (
    cd frontend
    TEST_API_BASE_URL="http://127.0.0.1:${TEST_API_PORT}" bun run test:integration
  )

  echo "==> Frontend integration tests passed"
  exit 0
fi

cleanup() {
  docker rm -f "$API_CONTAINER" >/dev/null 2>&1 || true
}
trap cleanup EXIT

echo "==> Starting test database (compose db)"
docker compose up -d db

echo "==> Waiting for database to accept connections"
for i in $(seq 1 30); do
  if docker compose exec -T db pg_isready -U "$DB_ADMIN" -d postgres >/dev/null 2>&1; then
    break
  fi
  if [ "$i" -eq 30 ]; then
    echo "database did not become ready" >&2
    exit 1
  fi
  sleep 1
done

echo "==> Resetting isolated test database: $TEST_DB (never the dev database)"
docker compose exec -T db psql -v ON_ERROR_STOP=1 -U "$DB_ADMIN" -d postgres \
  -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '$TEST_DB' AND pid <> pg_backend_pid();" \
  >/dev/null
docker compose exec -T db psql -v ON_ERROR_STOP=1 -U "$DB_ADMIN" -d postgres \
  -c "DROP DATABASE IF EXISTS \"$TEST_DB\";"
docker compose exec -T db psql -v ON_ERROR_STOP=1 -U "$DB_ADMIN" -d postgres \
  -c "CREATE DATABASE \"$TEST_DB\" OWNER \"$DB_USER\";"

echo "==> Building and starting test backend on port $TEST_API_PORT"
docker rm -f "$API_CONTAINER" >/dev/null 2>&1 || true
docker compose run -d --name "$API_CONTAINER" --build -p "${TEST_API_PORT}:8080" \
  -e DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@db:5432/${TEST_DB}?sslmode=disable" \
  -e HOST=0.0.0.0 \
  -e PORT=8080 \
  -e ENV=test \
  -e SEED_MAINTAINER_EMAIL= \
  -e SEED_MAINTAINER_PASSWORD= \
  --no-deps backend >/dev/null

echo "==> Waiting for test backend (migrations + seed run on startup)"
for i in $(seq 1 60); do
  if curl -fsS "http://127.0.0.1:${TEST_API_PORT}/health" >/dev/null 2>&1; then
    break
  fi
  if [ "$i" -eq 60 ]; then
    echo "test backend did not become healthy; logs:" >&2
    docker logs "$API_CONTAINER" >&2 || true
    exit 1
  fi
  sleep 1
done

echo "==> Inserting fixture topic for 'fewer than 5 questions' (2 published questions)"
docker compose exec -T db psql -v ON_ERROR_STOP=1 -U "$DB_USER" -d "$TEST_DB" <<'SQL'
INSERT INTO topics (id, slug, title, description, position, status)
VALUES ('topic-few-fixture', 'quiz-few', 'Quiz Few Fixture', 'Fixture with two questions.', 999, 'published')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO questions (id, topic_id, text, explanation, status)
VALUES
  ('q-few-1', 'topic-few-fixture', 'Fixture question one?', 'Explanation one.', 'published'),
  ('q-few-2', 'topic-few-fixture', 'Fixture question two?', 'Explanation two.', 'published')
ON CONFLICT (id) DO NOTHING;

INSERT INTO answer_options (id, question_id, text, position, is_correct, status)
VALUES
  ('o-few-1a', 'q-few-1', 'Option A1', 1, true, 'published'),
  ('o-few-1b', 'q-few-1', 'Option B1', 2, false, 'published'),
  ('o-few-2a', 'q-few-2', 'Option A2', 1, false, 'published'),
  ('o-few-2b', 'q-few-2', 'Option B2', 2, true, 'published')
ON CONFLICT (id) DO NOTHING;
SQL

echo "==> Running frontend integration tests"
(
  cd frontend
  TEST_API_BASE_URL="http://127.0.0.1:${TEST_API_PORT}" bun run test:integration
)

echo "==> Frontend integration tests passed"
