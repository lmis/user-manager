package config

import (
	"user-manager/db"
	"user-manager/util"
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

func GetConfig(log util.Logger) (*Config, error) {
	config := &Config{}

	if err := util.ParseEnv(config); err != nil {
		return nil, util.Wrap("error parsing env", err)
	}

	return config, nil
}
