//go:build tools
// +build tools

package main

import (
  _ "github.com/google/wire/cmd/wire"
  _ "github.com/golangci/golangci-lint/cmd/golangci-lint"
)
