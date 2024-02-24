package domain_model

import (
	env "github.com/caarlos0/env/v6"
	"time"
	"user-manager/db"
	"user-manager/util/errs"
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

func GetConfig() (*Config, error) {
	config := &Config{}

	var target interface{} = config
	if err := env.Parse(target, env.Options{RequiredIfNoDef: true}); err != nil {
		return nil, errs.Wrap("issue parsing environment", err)
	}

	return config, nil
}
