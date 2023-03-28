package injector

import (
	dm "user-manager/domain-model"
)

var config *dm.Config

func SetupConfigProvider(c *dm.Config) {
	if c == nil {
		panic("Invalid singleton setup: config is nil")
	}
	config = c
}

func ProvideConfig() *dm.Config {
	return config
}
