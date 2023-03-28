package main

import (
	"embed"
	"fmt"
	"os"
	"user-manager/cmd/migrator/config"
	"user-manager/db"
	"user-manager/util/command"
	"user-manager/util/errors"
	"user-manager/util/exec"
	"user-manager/util/logger"
	"user-manager/util/random"

	"github.com/go-jet/jet/v2/generator/postgres"
	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
)

//go:embed migrations/*
var migrationsFS embed.FS

func main() {
	command.Run("MIGRATOR", runMigrator)
}

func runMigrator(log logger.Logger) error {
	if len(os.Args) > 1 && os.Args[1] == "generate" {
		return generateSqlBoiler(log)
	}
	conf, err := config.GetConfig()
	if err != nil {
		return errors.Wrap("cannot read config", err)
	}
	return applyMigrations(&conf.DbInfo, log)
}

func applyMigrations(dbInfo *db.Info, log logger.Logger) error {
	connection, err := dbInfo.OpenDbConnection(log)
	if err != nil {
		return errors.Wrap("could not open db connection", err)
	}
	defer db.CloseOrPanic(connection)

	log.Info("Running migrations")
	migrate.SetTable("migrations")

	migrations := &migrate.EmbedFileSystemMigrationSource{FileSystem: migrationsFS, Root: "migrations"}
	numApplied, err := migrate.Exec(connection, "postgres", migrations, migrate.Up)
	if err != nil {
		return errors.Wrap("issue executing migration", err)
	}

	if err != nil {
		return errors.Wrap("could not run migration", err)
	}
	log.Info("Applied %d migrations", numApplied)
	return nil
}

func generateSqlBoiler(log logger.Logger) error {
	outputDir := ""
	if len(os.Args) > 2 {
		outputDir = os.Args[2]
	}
	if outputDir == "" {
		return errors.Error("no output directory provided as commandline argument")
	}

	log.Info("Starting local postgres docker container")

	dbInfo := &db.Info{
		DbName:   "postgres",
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: random.MakeRandomURLSafeB64(21),
	}
	cmd := fmt.Sprintf("docker run --name postgres-migration-container -p 5432:5432 -e POSTGRES_PASSWORD=%s -d postgres > /dev/null", dbInfo.Password)
	if err := exec.RunShellCommand(cmd); err != nil {
		return errors.Wrap("failed to start local postgres", err)
	}

	defer func() {
		log.Info("Cleaning up local postgres docker container")
		if err := exec.RunShellCommand("docker rm -f postgres-migration-container > /dev/null"); err != nil {
			panic(errors.Wrap("cleanup of local docker container failed", err))
		}
	}()

	dbConnection, err := dbInfo.OpenDbConnection(log)
	if err != nil {
		return errors.Wrap("cannot open db", err)
	}
	defer db.CloseOrPanic(dbConnection)

	log.Info("Running migrations")
	if err = applyMigrations(dbInfo, log); err != nil {
		return errors.Wrap("issue executing migration", err)
	}

	jetConfig := postgres.DBConnection{
		Host:       dbInfo.Host,
		Port:       dbInfo.Port,
		User:       dbInfo.User,
		Password:   dbInfo.Password,
		DBName:     dbInfo.DbName,
		SchemaName: "public",
		SslMode:    "disable",
	}
	if err = postgres.Generate(outputDir, jetConfig); err != nil {
		return errors.Wrap("issue running jet", err)
	}
	log.Info("Generated models under %s", outputDir)
	return nil
}
