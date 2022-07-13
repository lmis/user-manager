package config

import (
	"user-manager/db"
	"user-manager/util"
)

type Config struct {
	DbInfo db.DbInfo
}

func GetConfig(log util.Logger) (*Config, error) {
	config := &Config{}

	if err := util.ParseEnv(config); err != nil {
		return nil, util.Wrap("error parsing env", err)
	}

	return config, nil
}
