# syntax=docker/dockerfile:1

# ---------- Build stage ----------
FROM golang:1.26.0-alpine AS builder

# go-sqlite3 uses cgo, so we need a C toolchain
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Cache go modules separately from source changes
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the backend source
COPY cmd ./cmd
COPY internal ./internal
COPY database ./database

# CGO must stay enabled for mattn/go-sqlite3
ENV CGO_ENABLED=1
RUN go build -o /app/server ./cmd/server

# Build the golang-migrate CLI with the sqlite3 driver baked in (it needs
# cgo too, hence building it here rather than downloading a release binary)
RUN go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# ---------- Runtime stage ----------
FROM alpine:3.20

RUN apk add --no-cache ca-certificates sqlite-libs tzdata

WORKDIR /app

COPY --from=builder /app/server ./server
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate

# Bring migrations along so the entrypoint can apply them at container start
COPY internal/migrations ./internal/migrations

COPY entrypoint.sh ./entrypoint.sh
RUN chmod +x ./entrypoint.sh

# The app expects ./db (sqlite file) and ./uploads (user media) relative
# to its working directory — create them so they exist even before a
# volume is mounted.
RUN mkdir -p ./db ./uploads/avatars ./uploads/posts ./uploads/groups/avatars

EXPOSE 4000

# ORBIT_TOKEN_SECRET is read from the environment at runtime (see
# internal/app/tokens/tokens.go) — set it via `docker run -e` or
# docker-compose's `environment:` / `env_file:`.

ENTRYPOINT ["./entrypoint.sh"]