package main

//go:generate go run ../generate-sqlboiler/main.go ../../db/generated/models
import (
	"user-manager/config"
	"user-manager/db"
	"user-manager/util"

	"database/sql"

	_ "github.com/lib/pq"
)

func main() {
	util.Run("MIGRATOR", runMigrations)
}

func runMigrations(log util.Logger) error {
	var dbConnection *sql.DB

	config, err := config.GetConfig(log)
	if err != nil {
		return util.Wrap("runMigrations", "cannot read config", err)
	}

	dbConnection, err = config.OpenDbConnection(log)
	if err != nil {
		return util.Wrap("runMigrations", "could not open db connection", err)
	}
	defer util.CloseOrPanic(dbConnection)

	log.Info("Running migrations")
	n, err := db.MigrateUp(dbConnection)
	if err != nil {
		return util.Wrap("runMigrations", "could not run migration", err)
	}
	log.Info("Applied %d migrations", n)
	return nil
}
