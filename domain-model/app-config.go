package domain_model

import (
	"time"
	"user-manager/db"
	"user-manager/util"
)

type Config struct {
	DbInfo      db.DbInfo
	AppPort     string `env:"PORT"`
	AppUrl      string `env:"APP_URL"`
	ServiceName string `env:"SERVICE_NAME"`
	EmailFrom   string `env:"EMAIL_FROM"`
	Environment string `env:"ENVIRONMENT"`
}

const (
	LOGIN_SESSION_DURATION        = 60 * time.Minute
	SUDO_SESSION_DURATION         = 10 * time.Minute
	DEVICE_SESSION_DURATION       = 30 * 24 * time.Hour
	PASSWORD_RESET_TOKEN_DURATION = 1 * time.Hour
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

	if err := util.ParseEnv(config); err != nil {
		return nil, util.Wrap("issue parsing environment", err)
	}

	return config, nil
}
