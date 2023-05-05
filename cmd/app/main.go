package main

//go:generate go run github.com/google/wire/cmd/wire ./...

import (
	"database/sql"
	"embed"
	"user-manager/cmd/app/router"
	"user-manager/db"
	dm "user-manager/domain-model"
	"user-manager/util/command"
	"user-manager/util/errors"
	util "user-manager/util/http"
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

	config, err := dm.GetConfig()
	if err != nil {
		return errors.Wrap("cannot read config", err)
	}

	if !config.IsLocalEnv() {
		gin.SetMode(gin.ReleaseMode)
	}

	database, err = config.DbInfo.OpenDbConnection(log)
	if err != nil {
		return errors.Wrap("could not open db connection", err)
	}
	defer db.CloseOrPanic(database)

	engine, err := router.New(translationsFS, config, database)
	if err != nil {
		return errors.Wrap("cannot setup router", err)
	}
	if err = util.RunHttpServer(log, &http.Server{
		Addr:         ":" + config.AppPort,
		Handler:      engine,
		ReadTimeout:  1 * time.Minute,
		WriteTimeout: 1 * time.Minute,
	}); err != nil {
		return errors.Wrap("error running http server", err)
	}

	return nil
}
