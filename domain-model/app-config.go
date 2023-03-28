package domain_model

import (
	"time"
	"user-manager/db"
	"user-manager/util/env"
	"user-manager/util/errors"
)

type Config struct {
	DbInfo      db.Info
	AppPort     string `env:"PORT"`
	AppUrl      string `env:"APP_URL"`
	ServiceName string `env:"SERVICE_NAME"`
	EmailFrom   string `env:"EMAIL_FROM"`
	Environment string `env:"ENVIRONMENT"`
}

const (
	LoginSessionDuration       = 60 * time.Minute
	SudoSessionDuration        = 10 * time.Minute
	DeviceSessionDuration      = 30 * 24 * time.Hour
	PasswordResetTokenDuration = 1 * time.Hour
)

func (conf *Config) IsLocalEnv() bool {
	return conf.Environment == "local"
}

func (conf *Config) IsProdEnv() bool {
	return conf.Environment == "production"
}

func (conf *Config) IsStagingEnv() bool {
	return conf.Environment == "staging"
}

// TODO: Make this more strict with regards to missing env variables for non-local environments
func GetConfig() (*Config, error) {
	config := &Config{}

	if err := env.ParseEnv(config); err != nil {
		return nil, errors.Wrap("issue parsing environment", err)
	}

	return config, nil
}
