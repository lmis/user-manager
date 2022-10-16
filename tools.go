//go:build tools
// +build tools

package main

import (
	_ "github.com/google/wire/cmd/wire"
	_ "github.com/kisielk/errcheck"
	_ "github.com/volatiletech/sqlboiler/v4"
	_ "github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql"
)
