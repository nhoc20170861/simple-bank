#!/bin/sh

set -e
echo "run db migrate"

/app/migrate -path /app/migration -database "$DB_SOURCE" up

echo "run the app"
exec "$@"