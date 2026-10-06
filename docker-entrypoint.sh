#!/bin/sh
set -e

mkdir -p /app/db /app/uploads

# Token signing secret: use ORBIT_TOKEN_SECRET if given, otherwise generate
# one once and keep it in the db volume (the app reads ./db/tokens-signature.txt).
if [ -z "$ORBIT_TOKEN_SECRET" ] && [ ! -s /app/db/tokens-signature.txt ]; then
    echo "Generating token signature..."
    head -c 48 /dev/urandom | base64 | tr -d '\n' > /app/db/tokens-signature.txt
fi

echo "Running migrations..."
migrate -path ./internal/migrations -database "sqlite3://./db/social_network.db" up

echo "Starting server on :4000"
exec /app/server
