package util

import (
	env "github.com/caarlos0/env/v6"
)

func ParseEnv(target interface{}) error {
	opts := env.Options{RequiredIfNoDef: true}
	return env.Parse(target, opts)
}
