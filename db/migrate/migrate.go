package migrate

import (
	"database/sql"
	"user-manager/util"

	sql_migrate "github.com/rubenv/sql-migrate"
)

func MigrateUp(db *sql.DB, dir string) (int, error) {
	sql_migrate.SetTable("migrations")
	migrations := &sql_migrate.FileMigrationSource{Dir: "../../db/migrate"}

	numApplied, err := sql_migrate.Exec(db, "postgres", migrations, sql_migrate.Up)
	if err != nil {
		return 0, util.Wrap("issue executing migration", err)
	}
	return numApplied, nil
}
