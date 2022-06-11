package main

//go:generate go run ../generate-sqlboiler/main.go ../../db/generated/models
import (
	"user-manager/app"
	"user-manager/db"
	"user-manager/util"

	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"database/sql"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type dbInfo struct {
	dbname   string
	host     string
	port     string
	user     string
	password string
}

func main() {
	util.SetLogJSON(os.Getenv("LOG_JSON") != "")
	log := util.Log("LIFECYCLE")
	exitCode := 0

	defer func() {
		if p := recover(); p != nil {
			stack := string(debug.Stack())
			log.Err(fmt.Errorf("panic: %v\n%v", p, stack))
			exitCode = 1
		}
		log.Info("Shutdown complete. ExitCode: %d", exitCode)
		os.Exit(exitCode)
	}()

	if err := runServer(log); err != nil {
		log.Err(err)
		log.Warn("Server exited abnormally")
		exitCode = 1
	}
}

func runServer(log *util.Logger) error {
	log.Info("Starting up")

	var httpServer *http.Server
	var dbConnection *sql.DB

	port := util.GetEnvOrDefault(log, "PORT", "8080")
	environment := util.GetEnvOrDefault(log, "ENVIRONMENT", "local")
	if environment != "local" {
		gin.SetMode(gin.ReleaseMode)
	}

	dbInfo := &dbInfo{
		dbname:   os.Getenv("DB_NAME"),
		host:     os.Getenv("DB_HOST"),
		port:     os.Getenv("DB_PORT"),
		user:     os.Getenv("DB_USER"),
		password: os.Getenv("DB_PASSWORD"),
	}

	if dbInfo.dbname == "" || dbInfo.host == "" || dbInfo.port == "" || dbInfo.user == "" || dbInfo.password == "" {
		if environment == "local" {
			var cleanup func()
			var err error
			dbInfo, cleanup, err = startLocalDevDb(log)
			if err != nil {
				return err
			}
			defer cleanup()
		} else {
			return fmt.Errorf("missing database connection information")
		}
	}

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbInfo.host, dbInfo.port, dbInfo.user, dbInfo.password, dbInfo.dbname)
	dbConnection, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return err
	}
	defer dbConnection.Close()
	numAttempts := 10
	sleepTime := 500 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(numAttempts)*sleepTime+1*time.Second)
	defer cancel()
	log.Info("Pinging DB")
	err = dbConnection.PingContext(ctx)
	for attempts := 1; err != nil && attempts < numAttempts; attempts++ {
		time.Sleep(sleepTime)
		log.Info("Retry pinging DB")
		err = dbConnection.PingContext(ctx)
	}
	if err != nil {
		return err
	}
	log.Info("Running migrations")
	n, err := db.MigrateUp(dbConnection)
	if err != nil {
		return err
	}
	log.Info("Applied %d migrations", n)

	httpServer = &http.Server{
		Addr:         ":" + port,
		Handler:      app.New(dbConnection),
		ReadTimeout:  1 * time.Minute,
		WriteTimeout: 1 * time.Minute,
	}

	log.Info("Starting http server on port %s", port)
	httpServerError := make(chan error, 1)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				log.Info("Http server closed")
			} else {
				httpServerError <- err
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
		ctx, cancel = context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		err = httpServer.Shutdown(ctx)
		if err != nil {
			return err
		}

		log.Info("Http server has shutdown normally")
	case err = <-httpServerError:
		return err
	}

	return nil
}

func startLocalDevDb(log *util.Logger) (*dbInfo, func(), error) {
	log.Info("Starting local postgres docker container")
	generatedPassword, err := util.MakeRandomURLSafeB64(21)
	if err != nil {
		return nil, nil, err
	}

	dbInfo := &dbInfo{
		dbname:   "postgres",
		host:     "localhost",
		port:     "5432",
		user:     "postgres",
		password: generatedPassword,
	}
	cmd := fmt.Sprintf("docker run --name postgres-local-dev -p 5432:5432 -e POSTGRES_PASSWORD=%s -d postgres", generatedPassword)
	if err = util.RunShellCommand(cmd); err != nil {
		return nil, nil, err
	}
	cleanup := func() {
		log.Info("Cleaning up local postgres docker container")
		if err := util.RunShellCommand("docker rm -f postgres-local-dev"); err != nil {
			panic(err)
		}
	}

	return dbInfo, cleanup, nil
}
