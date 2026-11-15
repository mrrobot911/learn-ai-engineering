FROM golang:1.23.8-alpine3.20 AS builder
WORKDIR /build
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/receiver ./cmd/receiver

FROM alpine:3.20.8
ARG OPENCODE_VERSION=1.3.2
# apk packages are not pinned: Alpine 3.20.x repos do not retain every package
# revision, so exact apk versions make image builds non-reproducible over time.
# Base images and opencode-ai are pinned for reproducibility.
RUN apk add --no-cache ca-certificates git nodejs npm \
	&& npm install -g opencode-ai@${OPENCODE_VERSION} \
	&& npm cache clean --force \
	&& addgroup -S app && adduser -S -G app app \
	&& mkdir -p /data/incidents /data/repo \
	&& chown -R app:app /data
ENV INCIDENT_BUNDLE_DIR=/data/incidents \
	REPO_DIR=/data/repo
WORKDIR /app
COPY --from=builder /out/receiver /app/receiver
USER app
EXPOSE 8081
ENTRYPOINT ["/app/receiver"]
