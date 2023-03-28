package config

import (
	"user-manager/db"
	"user-manager/util/env"
	"user-manager/util/errors"
)

type Config struct {
	DbInfo db.Info
}

func GetConfig() (*Config, error) {
	config := &Config{}

	if err := env.ParseEnv(config); err != nil {
		return nil, errors.Wrap("error parsing env", err)
	}

	return config, nil
}
