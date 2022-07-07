#!/bin/sh
echo "go mod tidy"
go mod tidy
echo "go generate app"
go generate cmd/app/main.go
echo "go fmt"
go fmt ./...
echo "go vet"
go vet ./...
echo "errcheck"
errcheck -exclude .errcheck-ignore -asserts $(go list ./... | grep -v db/generated/)
