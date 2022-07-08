package main

import (
	"user-manager/db"
	"user-manager/util"

	"fmt"
	"os"

	"database/sql"

	_ "github.com/lib/pq"
)

func main() {
	util.Run("SQL-BOILER", generateSqlBoiler)
}

func generateSqlBoiler(log util.Logger) error {
	outputDir := ""
	if len(os.Args) > 1 {
		outputDir = os.Args[1]
	}
	if outputDir == "" {
		return util.Errorf("generateSqlBoiler", "no output directory provided as commandline argument")
	}

	log.Info("Starting local postgres docker container")

	dbName := "postgres"
	dbHost := "localhost"
	dbPort := "5432"
	dbUser := "postgres"
	dbPassword := util.MakeRandomURLSafeB64(21)
	cmd := fmt.Sprintf("docker run --name postgres-migration-container -p 5432:5432 -e POSTGRES_PASSWORD=%s -d postgres > /dev/null", dbPassword)
	if err := util.RunShellCommand(cmd); err != nil {
		return util.Wrap("generateSqlBoiler", "failed to start local postgres", err)
	}

	defer func() {
		log.Info("Cleaning up local postgres docker container")
		if err := util.RunShellCommand("docker rm -f postgres-migration-container > /dev/null"); err != nil {
			panic(util.Wrap("generateSqlBoiler", "cleanup of local docker container failed", err))
		}
	}()

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName)
	dbConnection, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return util.Wrap("generateSqlBoiler", "cannot open db", err)
	}
	defer util.CloseOrPanic(dbConnection)
	err = util.CheckConnection(log, dbConnection)
	if err != nil {
		return util.Wrap("generateSqlBoiler", "issue checking connection", err)
	}
	log.Info("Running migrations")
	n, err := db.MigrateUp(dbConnection)
	if err != nil {
		return util.Wrap("generateSqlBoiler", "issue migrating up", err)
	}
	log.Info("Applied %d migrations", n)

	file, err := os.Create("/tmp/sqlboiler-config.toml")
	if err != nil {
		return util.Wrap("generateSqlBoiler", "issue creating tmp file", err)
	}
	log.Info("Temporary file created in %s", file.Name())
	defer func() {
		err = file.Close()
		if err != nil {
			panic(util.Wrap("generateSqlBoiler", "issue closing temp file", err))
		}
		log.Info("Removing file %s", file.Name())
		err = os.Remove(file.Name())
		if err != nil {
			panic(util.Wrap("generateSqlBoiler", "issue removing temp file", err))
		}
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
		return util.Wrap("generateSqlBoiler", "issue writing to temp file", err)
	}
	if err := util.RunShellCommand(fmt.Sprintf("sqlboiler psql --config=%s", file.Name())); err != nil {
		return util.Wrap("generateSqlBoiler", "issue running sqlboiler", err)
	}
	log.Info("Generated models under %s", outputDir)
	return nil
}
