package config

import (
	"user-manager/util/env"
	"user-manager/util/errors"
	"user-manager/util/logger"
)

type Config struct {
	Port   string `env:"MOCK_API_PORT"`
	AppUrl string `env:"APP_URL"`
}

func GetConfig(log logger.Logger) (*Config, error) {
	config := &Config{}

	if err := env.ParseEnv(config); err != nil {
		return nil, errors.Wrap("error parsing env", err)
	}
	return config, nil
}
