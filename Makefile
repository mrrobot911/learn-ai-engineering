.PHONY: test test-integration check up down

# Unit tests (backend + frontend). Does not require a database.
test:
	(cd backend && go test ./...)
	(cd frontend && bun install --frozen-lockfile && bun run test)

# Integration tests on an isolated test database (never the dev database).
# Backend runs with -p 1 because packages share one database.
test-integration:
	docker compose up -d db
	@echo "==> Waiting for database"
	@for i in $$(seq 1 30); do \
		if docker compose exec -T db pg_isready -U "$${POSTGRES_ADMIN_USER:-$${POSTGRES_USER:-learn}}" -d postgres >/dev/null 2>&1; then break; fi; \
		if [ "$$i" -eq 30 ]; then echo "database did not become ready" >&2; exit 1; fi; \
		sleep 1; \
	done
	@echo "==> Resetting isolated backend integration database"
	@DB_USER="$${POSTGRES_USER:-learn}"; DB_ADMIN="$${POSTGRES_ADMIN_USER:-$$DB_USER}"; TEST_DB="$${BACKEND_TEST_DB:-learn_ai_engineering_test}"; \
	docker compose exec -T db psql -v ON_ERROR_STOP=1 -U "$$DB_ADMIN" -d postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '$$TEST_DB' AND pid <> pg_backend_pid();" >/dev/null; \
	docker compose exec -T db psql -v ON_ERROR_STOP=1 -U "$$DB_ADMIN" -d postgres -c "DROP DATABASE IF EXISTS \"$$TEST_DB\";"; \
	docker compose exec -T db psql -v ON_ERROR_STOP=1 -U "$$DB_ADMIN" -d postgres -c "CREATE DATABASE \"$$TEST_DB\" OWNER \"$$DB_USER\";";
	@DB_USER="$${POSTGRES_USER:-learn}"; DB_PASSWORD="$${POSTGRES_PASSWORD:-learn}"; TEST_DB="$${BACKEND_TEST_DB:-learn_ai_engineering_test}"; \
	TEST_DATABASE_URL="postgres://$$DB_USER:$$DB_PASSWORD@localhost:$${DB_PORT:-5432}/$$TEST_DB?sslmode=disable" \
		$(MAKE) -s _test-integration-backend TEST_DATABASE_URL="$$TEST_DATABASE_URL"
	bash scripts/frontend-integration.sh

_test-integration-backend:
	(cd backend && TEST_DATABASE_URL="$(TEST_DATABASE_URL)" go test -tags=integration -p 1 ./...)

# Checks that do not change files.
check:
	@echo "==> gofmt"
	@test -z "$$(cd backend && gofmt -l .)" || (echo "gofmt needed:"; cd backend && gofmt -l .; exit 1)
	@echo "==> go vet"
	(cd backend && go vet ./...)
	@echo "==> frontend typecheck"
	(cd frontend && bun install --frozen-lockfile && bun run typecheck)
	@echo "==> frontend lint"
	(cd frontend && bun run lint)
	@echo "==> frontend format:check"
	(cd frontend && bun run format:check)
	@echo "==> redocly lint"
	bunx @redocly/cli lint openapi.yaml

up:
	docker compose up --build

down:
	docker compose down
