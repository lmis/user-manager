package config

import (
	"os"
	"user-manager/util"
)

type Config struct {
	DbInfo      DbInfo
	Port        string
	environment string
}

type DbInfo struct {
	DbName   string
	Host     string
	Port     string
	User     string
	Password string
}

func (conf *Config) IsLocalEnv() bool {
	return conf.environment == "local"
}

func (conf *Config) IsProdEnv() bool {
	return conf.environment == "production"
}

func (conf *Config) IsStagingEnv() bool {
	return conf.environment == "staging"
}

func GetConfig(log util.Logger) (*Config, error) {
	config := &Config{
		environment: getEnvOrDefault(log, "ENVIRONMENT", "local"),
		Port:        getEnvOrDefault(log, "PORT", "8080"),
		DbInfo: DbInfo{
			DbName:   os.Getenv("DB_NAME"),
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
		},
	}

	return config, nil
}

func getEnvOrDefault(log util.Logger, envVar string, defaultVal string) string {
	res := os.Getenv(envVar)
	if res == "" {
		res = defaultVal
		log.Warn("Env variable \"%s\" not set, defaulting to %s", envVar, defaultVal)
	}
	return res
}
