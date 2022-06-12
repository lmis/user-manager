package util

import "database/sql"

func CloseOrPanic(db *sql.DB) {
	err := db.Close()
	if err != nil {
		panic(err)
	}
}
