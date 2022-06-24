#!/bin/sh
echo "go generate app"
go generate cmd/app/main.go
echo "go mod tidy"
go mod tidy
echo "go fmt"
go fmt ./...
echo "go vet"
go vet ./...
echo "errcheck"
../../bin/errcheck -exclude .errcheck-ignore -asserts $(go list ./... | grep -v db/generated/)
