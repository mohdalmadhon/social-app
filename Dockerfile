ARG GO_VERSION=1.26.0
FROM golang:${GO_VERSION}-bookworm AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY database ./database

ENV CGO_ENABLED=1
RUN go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

RUN go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest \
    && cp "$(go env GOPATH)/bin/migrate" /out/migrate

FROM debian:bookworm-slim
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates tzdata \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /out/server /app/server
COPY --from=builder /out/migrate /usr/local/bin/migrate

COPY internal/migrations ./internal/migrations
COPY images ./images

COPY uploads ./uploads

COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN sed -i 's/\r$//' /usr/local/bin/docker-entrypoint.sh \
    && chmod +x /usr/local/bin/docker-entrypoint.sh \
    && mkdir -p /app/db /app/uploads

EXPOSE 4000
VOLUME ["/app/db", "/app/uploads"]

ENTRYPOINT ["docker-entrypoint.sh"]
