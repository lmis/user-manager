package main

//go:generate go run ../generate-sqlboiler/main.go ../../db/generated/models
import (
	"path"
	"user-manager/cmd/migrator/config"
	"user-manager/db/migrate"
	"user-manager/util"

	"database/sql"

	_ "github.com/lib/pq"
)

func main() {
	util.Run("MIGRATOR", runMigrations)
}

func runMigrations(log util.Logger, dir string) error {
	var dbConnection *sql.DB

	config, err := config.GetConfig(log)
	if err != nil {
		return util.Wrap("cannot read config", err)
	}

	dbConnection, err = config.DbInfo.OpenDbConnection(log)
	if err != nil {
		return util.Wrap("could not open db connection", err)
	}
	defer util.CloseOrPanic(dbConnection)

	log.Info("Running migrations")
	n, err := migrate.MigrateUp(dbConnection, path.Join(dir, "../../db/migrate"))
	if err != nil {
		return util.Wrap("could not run migration", err)
	}
	log.Info("Applied %d migrations", n)
	return nil
}
