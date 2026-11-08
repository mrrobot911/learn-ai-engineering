# Build stage: pin Bun version (also recorded in .bun-version).
FROM oven/bun:1.4.2 AS builder

WORKDIR /app

# Install dependencies using the lockfile for reproducible builds.
COPY package.json bun.lock ./
RUN bun install --frozen-lockfile

COPY . .

# Vite embeds VITE_* variables at build time.
ARG VITE_API_MODE=http
ARG VITE_API_BASE_URL=/api
ENV VITE_API_MODE=${VITE_API_MODE}
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}

RUN bun run build

# Runtime stage: pin the static server version.
FROM caddy:2.8.4-alpine

COPY Caddyfile /etc/caddy/Caddyfile
COPY --from=builder /app/dist /srv

EXPOSE 80
