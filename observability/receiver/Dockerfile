FROM golang:1.23.8-alpine3.20 AS builder
WORKDIR /build
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/receiver ./cmd/receiver

FROM alpine:3.20.8
RUN apk add --no-cache ca-certificates && addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=builder /out/receiver /app/receiver
USER app
EXPOSE 8081
ENTRYPOINT ["/app/receiver"]
