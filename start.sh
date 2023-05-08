#!/bin/sh

set -e

echo "running db migration"
source app.env
/app/migrate -path /app/migration -database "$DB_URL" -verbose up

echo "starting the app"
exec "$@"

