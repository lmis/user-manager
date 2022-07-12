package config

import (
	"os"
	"user-manager/util"
)

type Config struct {
	DbInfo      DbInfo
	AppPort     string
	AppUrl      string
	ServiceName string
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

// TODO: Make this more strict with regards to missing env variables for non-local environments
func GetConfig(log util.Logger) (*Config, error) {
	environment := getEnvOrDefault(log, "ENVIRONMENT", "local")
	config := &Config{
		environment: environment,
		AppPort:     getEnvOrDefault(log, "PORT", "8080"),
		AppUrl:      getEnvOrDefault(log, "APP_URL", "http://localhost:8080/"),
		ServiceName: getEnvOrDefault(log, "SERVICE_NAME", "TestApp"),
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
