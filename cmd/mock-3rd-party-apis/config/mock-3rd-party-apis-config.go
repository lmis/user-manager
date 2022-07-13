package config

import (
	"user-manager/util"
)

type Config struct {
	Port   string `env:"MOCK_API_PORT"`
	AppUrl string `env:"APP_URL"`
}

func GetConfig(log util.Logger) (*Config, error) {
	config := &Config{}

	if err := util.ParseEnv(config); err != nil {
		return nil, util.Wrap("error parsing env", err)
	}
	return config, nil
}
