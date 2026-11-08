#!/usr/bin/env bash
# Frontend integration tests (T31a / #85).
# Starts an isolated test database and the backend, applies migrations and seed,
# inserts the small fixture data the tests need, then runs the frontend
# integration suite. Never points at the development database.
#
# Usage (from the repository root):
#   bash scripts/frontend-integration.sh
#
# Requirements: Docker with Compose, Go (to build the backend image via Compose),
# Bun. The test backend listens on http://127.0.0.1:${TEST_API_PORT:-18080}.
set -eu

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

DB_USER="${POSTGRES_USER:-learn}"
DB_PASSWORD="${POSTGRES_PASSWORD:-learn}"
TEST_DB="${FRONTEND_TEST_DB:-learn_ai_fe_test}"
TEST_API_PORT="${TEST_API_PORT:-18080}"
API_CONTAINER="${FE_TEST_API_CONTAINER:-learn-ai-fe-test-api}"
# The compose db image makes POSTGRES_USER the superuser (can create/drop databases).
DB_ADMIN="${POSTGRES_ADMIN_USER:-$DB_USER}"

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
