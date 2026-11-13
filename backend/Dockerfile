# Build stage: pin Go version from go.mod (go 1.23).
FROM golang:1.23.8-alpine3.20 AS builder

WORKDIR /build

# Build with go.sum: copy module files first for layer caching.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

# Runtime stage: pinned base, no latest tag.
FROM alpine:3.20.8

RUN addgroup -S app && adduser -S -G app app

WORKDIR /app
COPY --from=builder /out/api /app/api
COPY --from=builder /build/migrations /app/migrations

USER app

EXPOSE 8080

ENTRYPOINT ["/app/api"]
