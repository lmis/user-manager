package injector

import (
	domain_model "user-manager/domain-model"
)

var config *domain_model.Config

func SetupConfigProvider(c *domain_model.Config) {
	if c == nil {
		panic("Invalid singleton setup: config is nil")
	}
	config = c
}

func ProvideConfig() *domain_model.Config {
	return config
}
