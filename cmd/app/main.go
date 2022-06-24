package main

//go:generate go run ../generate-sqlboiler/main.go ../../db/generated/models
import (
	"user-manager/app"
	"user-manager/config"
	"user-manager/util"

	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"database/sql"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	util.Run("LIFECYCLE", runServer)
}

func runServer(log util.Logger) error {
	log.Info("Starting up")

	var httpServer *http.Server
	var dbConnection *sql.DB

	config, err := config.GetConfig(log)
	if err != nil {
		return util.Wrap("runServer", "cannot read config", err)
	}

	if !config.IsLocalEnv() {
		gin.SetMode(gin.ReleaseMode)
	}

	dbConnection, err = config.OpenDbConnection(log)
	if err != nil {
		return util.Wrap("runServer", "could not open db connection", err)
	}
	defer util.CloseOrPanic(dbConnection)

	httpServer = &http.Server{
		Addr:         ":" + config.Port,
		Handler:      app.New(dbConnection, config),
		ReadTimeout:  1 * time.Minute,
		WriteTimeout: 1 * time.Minute,
	}

	log.Info("Starting http server on port %s", config.Port)
	httpServerError := make(chan error, 1)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				log.Info("Http server closed")
			} else {
				httpServerError <- util.Wrap("runServer", "httpServer stopped with unexpected error", err)
			}
		}
	}()

	// Block until shutdown signal or server error is received
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	select {
	case <-signals:
		log.Info("Shutdown signal received. About to shut down")

		log.Info("Shutting down http server down gracefully")
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		err = httpServer.Shutdown(ctx)
		if err != nil {
			return util.Wrap("runServer", "httpServer shutdown error", err)
		}

		log.Info("Http server has shutdown normally")
	case err = <-httpServerError:
		return err
	}

	return nil
}
