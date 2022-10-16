package injector

import "database/sql"

var database *sql.DB

func SetupDatabaseProvider(db *sql.DB) {
	if db == nil {
		panic("Invalid singleton setup: db is nil")
	}
	database = db
}
func ProvideDatabase() *sql.DB {
	return database
}
