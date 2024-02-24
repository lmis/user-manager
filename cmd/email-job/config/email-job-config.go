package config

import (
	env "github.com/caarlos0/env/v6"
	"user-manager/db"
	"user-manager/util/errs"
)

type Config struct {
	DbInfo      db.Info
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

func GetConfig() (*Config, error) {
	config := &Config{}

	var target interface{} = config
	if err := env.Parse(target, env.Options{RequiredIfNoDef: true}); err != nil {
		return nil, errs.Wrap("error parsing env", err)
	}

	return config, nil
}
