package main

//go:generate go run github.com/google/wire/cmd/wire ./...

import (
	"database/sql"
	"embed"
	"user-manager/cmd/app/injector"
	"user-manager/cmd/app/router"
	"user-manager/db"
	domain_model "user-manager/domain-model"
	"user-manager/util/command"
	"user-manager/util/errors"
	httputil "user-manager/util/http"
	"user-manager/util/logger"

	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

//go:embed translations/*
var translationsFS embed.FS
var database *sql.DB

func main() {
	command.Run("LIFECYCLE", runServer)
}

func runServer(log logger.Logger) error {
	log.Info("Starting up")

	err := injector.SetupEmailTemplatesProviders(translationsFS)
	if err != nil {
		return errors.Wrap("cannot initialize email service", err)
	}

	config, err := domain_model.GetConfig()
	if err != nil {
		return errors.Wrap("cannot read config", err)
	}
	injector.SetupConfigProvider(config)

	if !config.IsLocalEnv() {
		gin.SetMode(gin.ReleaseMode)
	}

	database, err = config.DbInfo.OpenDbConnection(log)
	if err != nil {
		return errors.Wrap("could not open db connection", err)
	}
	defer db.CloseOrPanic(database)
	injector.SetupDatabaseProvider(database)

	if err = httputil.RunHttpServer(log, &http.Server{
		Addr:         ":" + config.AppPort,
		Handler:      router.New(),
		ReadTimeout:  1 * time.Minute,
		WriteTimeout: 1 * time.Minute,
	}); err != nil {
		return errors.Wrap("error running http server", err)
	}

	return nil
}
