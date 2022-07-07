#!/bin/sh
export DB_NAME=postgres
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=$(cat /dev/random | head -c 15 | base64)
_DOCKER_POSTGRES_NAME=postgres-local-dev-$(cat /proc/sys/kernel/random/uuid)

killIfExists() {
    if [[ -n "$1" ]]
    then
      kill $1
    fi
}

cleanup() {
    echo "------ CLEANUP ------"
    killIfExists $_APP_PID
    killIfExists $_EMAILER_PID

    sleep 1
    docker rm $_DOCKER_POSTGRES_NAME -f > /dev/null
}

trap cleanup exit

echo "------ START LOCAL DB ($_DOCKER_POSTGRES_NAME) ------"
docker run --name $_DOCKER_POSTGRES_NAME -p 5432:5432 -e POSTGRES_PASSWORD=$DB_PASSWORD -d postgres > /dev/null

echo "------ RUN MIGRATOR ------"
go run cmd/migrator/main.go

if [[ $? = 0 ]]
then
    echo "------ START APP ------"
    go run cmd/app/main.go &
    _APP_PID=$!
    echo "------ START EMAILER ------"
    go run cmd/email-job/main.go &
    _EMAILER_PID=$!

    sleep infinity
else
  echo "----- ABORTING -----"
fi
