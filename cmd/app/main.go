package main

//go:generate go run ../generate-sqlboiler/main.go ../../db/generated/models
import (
	"embed"
	config "user-manager/cmd/app/config"
	"user-manager/cmd/app/router"
	emailservice "user-manager/cmd/app/services/email"
	"user-manager/util"

	"net/http"
	"time"

	"database/sql"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

//go:embed translations/*
var translationsFS embed.FS

func main() {
	util.Run("LIFECYCLE", runServer)
}

func runServer(log util.Logger) error {
	log.Info("Starting up")

	err := emailservice.Initialize(log, translationsFS)
	if err != nil {
		return util.Wrap("cannot initialize email service", err)
	}

	var dbConnection *sql.DB

	config, err := config.GetConfig()
	if err != nil {
		return util.Wrap("cannot read config", err)
	}

	if !config.IsLocalEnv() {
		gin.SetMode(gin.ReleaseMode)
	}

	dbConnection, err = config.DbInfo.OpenDbConnection(log)
	if err != nil {
		return util.Wrap("could not open db connection", err)
	}
	defer util.CloseOrPanic(dbConnection)

	if err = util.RunHttpServer(log, &http.Server{
		Addr:         ":" + config.AppPort,
		Handler:      router.New(dbConnection, config),
		ReadTimeout:  1 * time.Minute,
		WriteTimeout: 1 * time.Minute,
	}); err != nil {
		return util.Wrap("error running http server", err)
	}

	return nil
}
