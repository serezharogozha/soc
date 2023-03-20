#!/bin/sh

until PGHOST=${DB_HOST} PGDATABASE=${DB_NAME} PGUSER=${DB_USER} PGPASSWORD=${DB_PASSWORD} psql -c '\q'; do
  >&2 echo "Postgres is unavailable - sleeping"
  sleep 5
done

if [[ ! -f .env]]; then
  echo "Creating .env file"
  cp .env.example .env
fi

echo "*** Begin migrate ***"
migrate -source file://${PWD}/migrations/ -database postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable\&x-migrations-table=migrations up
echo "*** End migrate ***"

go run ./cmd/main.go

eval "$@"

