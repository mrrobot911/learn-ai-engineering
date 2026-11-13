# Build stage: frontend (pin Bun version, also recorded in .bun-version).
FROM oven/bun:1.4.2 AS frontend

WORKDIR /app

COPY frontend/package.json frontend/bun.lock ./
RUN bun install --frozen-lockfile

COPY frontend/ .

# Vite embeds VITE_* variables at build time; the app is same-origin behind /api.
ARG VITE_API_MODE=http
ARG VITE_API_BASE_URL=/api
ENV VITE_API_MODE=${VITE_API_MODE}
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}

RUN bun run build

# Build stage: backend (pin Go version from backend/go.mod, go 1.23).
FROM golang:1.23.8-alpine3.20 AS backend

WORKDIR /build

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

# Runtime stage: pinned base, backend serves the frontend build itself.
FROM alpine:3.20.8

RUN addgroup -S app && adduser -S -G app app

WORKDIR /app
COPY --from=backend /out/api /app/api
COPY --from=backend /build/migrations /app/migrations
COPY --from=frontend /app/dist /app/static

ENV STATIC_DIR=/app/static

USER app

EXPOSE 8080

ENTRYPOINT ["/app/api"]
