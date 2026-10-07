set -e

mkdir -p /app/db /app/uploads


if [ -z "$ORBIT_TOKEN_SECRET" ] && [ ! -s /app/db/tokens-signature.txt ]; then
    echo "Generating token signature..."
    head -c 48 /dev/urandom | base64 | tr -d '\n' > /app/db/tokens-signature.txt
fi

echo "Running migrations..."
migrate -path ./internal/migrations -database "sqlite3://./db/social_network.db" up

echo "Starting server on :4000"
exec /app/server
