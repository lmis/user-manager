#!/bin/sh
export DB_NAME=postgres
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=$(cat /dev/random | head -c 15 | base64)

echo "start local db"
docker run --name postgres-local-dev -p 5432:5432 -e POSTGRES_PASSWORD=$DB_PASSWORD -d postgres

go run cmd/migrator/main.go
go run cmd/app/main.go &
go run cmd/email-job/main.go &
