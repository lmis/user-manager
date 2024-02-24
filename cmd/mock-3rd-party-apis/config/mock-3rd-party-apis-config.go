package config

import (
	env "github.com/caarlos0/env/v6"
	"user-manager/util/errs"
)

type Config struct {
	Port   string `env:"MOCK_API_PORT"`
	AppUrl string `env:"APP_URL"`
}

func GetConfig() (*Config, error) {
	config := &Config{}

	var target interface{} = config
	if err := env.Parse(target, env.Options{RequiredIfNoDef: true}); err != nil {
		return nil, errs.Wrap("error parsing env", err)
	}
	return config, nil
}
