package main

//go:generate go run ../migrator/main.go generate ../../db/generated/models
import (
	"embed"
	config "user-manager/cmd/app/config"
	"user-manager/cmd/app/router"
	emailservice "user-manager/cmd/app/services/email"
	"user-manager/db"
	"user-manager/util"

	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

//go:embed translations/*
var translationsFS embed.FS

func main() {
	util.Run("LIFECYCLE", runServer)
}

func runServer(log util.Logger, dir string) error {
	log.Info("Starting up")

	err := emailservice.Initialize(log, translationsFS)
	if err != nil {
		return util.Wrap("cannot initialize email service", err)
	}

	config, err := config.GetConfig()
	if err != nil {
		return util.Wrap("cannot read config", err)
	}

	if !config.IsLocalEnv() {
		gin.SetMode(gin.ReleaseMode)
	}

	connection, err := config.DbInfo.OpenDbConnection(log)
	if err != nil {
		return util.Wrap("could not open db connection", err)
	}
	defer db.CloseOrPanic(connection)

	if err = util.RunHttpServer(log, &http.Server{
		Addr:         ":" + config.AppPort,
		Handler:      router.New(connection, config),
		ReadTimeout:  1 * time.Minute,
		WriteTimeout: 1 * time.Minute,
	}); err != nil {
		return util.Wrap("error running http server", err)
	}

	return nil
}
