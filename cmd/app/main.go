package main

import (
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"time"
	"user-manager/cmd/app/router"
	dm "user-manager/domain-model"
	"user-manager/util/command"
	"user-manager/util/db"
	"user-manager/util/errs"
	util "user-manager/util/http"
	"user-manager/util/logger"
)

func main() {
	slog.SetDefault(logger.NewLogger(false))
	command.Run(runServer)
}

func runServer() error {
	slog.Info("Starting up")

	config, err := dm.GetConfig()
	if err != nil {
		return errs.Wrap("cannot read config", err)
	}

	if !config.IsLocalEnv() {
		slog.SetDefault(logger.NewLogger(true))
		gin.SetMode(gin.ReleaseMode)
	}

	database, err := db.OpenDbConnection(config.DbInfo)
	if err != nil {
		return errs.Wrap("could not open db connection", err)
	}
	defer db.CloseOrPanic(database.Client())

	engine, err := router.New(config, database)
	if err != nil {
		return errs.Wrap("cannot setup router", err)
	}
	if err = util.RunHttpServer(&http.Server{
		Addr:         ":" + config.AppPort,
		Handler:      engine,
		ReadTimeout:  1 * time.Minute,
		WriteTimeout: 1 * time.Minute,
	}); err != nil {
		return errs.Wrap("error running http server", err)
	}

	return nil
}
