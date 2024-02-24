package main

import (
	"embed"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
	"user-manager/cmd/app/router"
	"user-manager/db"
	dm "user-manager/domain-model"
	"user-manager/util/command"
	"user-manager/util/errs"
	util "user-manager/util/http"
	"user-manager/util/logger"

	_ "github.com/lib/pq"
)

//go:embed translations/*
var translationsFS embed.FS

func main() {
	command.Run("LIFECYCLE", runServer)
}

func runServer(log logger.Logger) error {
	log.Info("Starting up")

	config, err := dm.GetConfig()
	if err != nil {
		return errs.Wrap("cannot read config", err)
	}

	if !config.IsLocalEnv() {
		gin.SetMode(gin.ReleaseMode)
	}

	database, err := db.OpenDbConnection(log, config.DbInfo)
	if err != nil {
		return errs.Wrap("could not open db connection", err)
	}
	defer db.CloseOrPanic(database.Client())

	engine, err := router.New(translationsFS, config, database)
	if err != nil {
		return errs.Wrap("cannot setup router", err)
	}
	if err = util.RunHttpServer(log, &http.Server{
		Addr:         ":" + config.AppPort,
		Handler:      engine,
		ReadTimeout:  1 * time.Minute,
		WriteTimeout: 1 * time.Minute,
	}); err != nil {
		return errs.Wrap("error running http server", err)
	}

	return nil
}
