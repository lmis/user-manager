package config

import (
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
