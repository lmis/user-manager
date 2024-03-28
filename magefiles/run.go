package main

import (
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var appEnv = map[string]string{
	"ENVIRONMENT":  "local",
	"PORT":         "8080",
	"APP_URL":      "http://localhost:8080",
	"SERVICE_NAME": "TestApp",
	"EMAIL_FROM":   "test-email-from@example.com",
	"DB_NAME":      "db",
	"DB_HOST":      "localhost",
	"DB_PORT":      "27017",
	"DB_USER":      "test",
	"DB_PASSWORD":  "mongo-test-password",
}

// Start checks then starts app, emailer and mock 3rd-party APIs
func Start() error {
	mg.Deps(Build.App, ComposeUpLocalEnvironment)
	return sh.RunWithV(appEnv, "bin/app")
}

// Watch checks then starts app, emailer and mock 3rd-party APIs each time a file changes
func Watch() error {
	//mg.Deps(Build.App, ComposeUpLocalEnvironment)

	return sh.RunV(wgo, "-xdir", "magefiles", "-xfile", "bin/app", "-xfile", ".*"+templGeneratedSuffix, "-xfile", tailwindOut, "mage", "start")
}

// ComposeUpLocalEnvironment starts a local MongoDB instance, emailer service and mock-3rd-party APIs as docker containers
func ComposeUpLocalEnvironment() error {
	mg.Deps(Check, Build.EmailJob, Build.MockApis)
	return sh.Run("docker-compose", "-f", "local-env-docker-compose.yml", "up", "-d")
}

// ComposeDownLocalEnvironment stops the docker containers
func ComposeDownLocalEnvironment() error {
	return sh.Run("docker-compose", "-f", "local-env-docker-compose.yml", "down")
}
