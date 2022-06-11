package main

import (
	"user-manager/db"
	"user-manager/util"

	"context"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"database/sql"

	_ "github.com/lib/pq"
)

func main() {
	util.SetLogJSON(false)
	log := util.Log("SQL-BOILER")

	exitCode := 0

	defer func() {
		if p := recover(); p != nil {
			stack := string(debug.Stack())
			log.Err(fmt.Errorf("panic: %v\n%v", p, stack))
			exitCode = 1
		}

		log.Info("Done. ExitCode: %d", exitCode)
		os.Exit(exitCode)
	}()

	err := generateSqlBoiler(log)
	if err != nil {
		exitCode = 1
		log.Warn("Aborted")
	}
}

func generateSqlBoiler(log *util.Logger) error {
	outputDir := ""
	if len(os.Args) > 1 {
		outputDir = os.Args[1]
	}
	if outputDir == "" {
		return fmt.Errorf("no output directory provided as commandline argument")
	}

	log.Info("Starting local postgres docker container")

	dbName := "postgres"
	dbHost := "localhost"
	dbPort := "5432"
	dbUser := "postgres"
	dbPassword, err := util.MakeRandomURLSafeB64(21)
	if err != nil {
		return err
	}
	cmd := fmt.Sprintf("docker run --name postgres-migration-container -p 5432:5432 -e POSTGRES_PASSWORD=%s -d postgres", dbPassword)
	if err = util.RunShellCommand(cmd); err != nil {
		return fmt.Errorf("failed to start local postgres")
	}

	defer func() {
		log.Info("Cleaning up local postgres docker container")
		if err := util.RunShellCommand("docker rm -f postgres-migration-container"); err != nil {
			panic(err)
		}
	}()

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName)
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

	file, err := os.Create("/tmp/sqlboiler-config.toml")
	if err != nil {
		return err
	}
	log.Info("Temporary file created in %s", file.Name())
	defer file.Close()
	defer func() {
		log.Info("Removing file %s", file.Name())
		os.Remove(file.Name())
	}()

	toml :=
		fmt.Sprintf(`
    output   = "%s"
    wipe     = true
    no-tests = true
    add-enum-types = true

    [psql]
    dbname = "%s"
    host   = "%s"
    port   = %s
    user   = "%s"
    pass   = "%s"
    sslmode = "disable"
    schema = "public"
    blacklist = ["migrations"]
    `, outputDir, dbName, dbHost, dbPort, dbUser, dbPassword)

	if _, err := file.Write([]byte(toml)); err != nil {
		return err
	}
	if err := util.RunShellCommand(fmt.Sprintf("sqlboiler psql --config=%s", file.Name())); err != nil {
		return err
	}
	log.Info("Generated models under %s", outputDir)
	return nil
}
