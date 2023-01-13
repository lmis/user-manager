package config

import (
	"user-manager/db"
	"user-manager/util/env"
	"user-manager/util/errors"
	"user-manager/util/logger"
)

type Config struct {
	DbInfo      db.DbInfo
	EmailApiUrl string `env:"EMAIL_API_URL"`
	Environment string `env:"ENVIRONMENT"`
}

func (conf *Config) IsLocalEnv() bool {
	return conf.Environment == "local"
}

func (conf *Config) IsProdEnv() bool {
	return conf.Environment == "production"
}

func (conf *Config) IsStagingEnv() bool {
	return conf.Environment == "staging"
}

func GetConfig(log logger.Logger) (*Config, error) {
	config := &Config{}

	if err := env.ParseEnv(config); err != nil {
		return nil, errors.Wrap("error parsing env", err)
	}

	return config, nil
}
