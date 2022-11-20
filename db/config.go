package db

import (
	"database/sql"
	"fmt"
	"runtime/debug"
	"user-manager/util"
)

type DbInfo struct {
	DbName   string `env:"DB_NAME"`
	Host     string `env:"DB_HOST"`
	Port     int    `env:"DB_PORT"`
	User     string `env:"DB_USER"`
	Password string `env:"DB_PASSWORD"`
}

func (dbInfo *DbInfo) OpenDbConnection(log util.Logger) (dbConnection *sql.DB, err error) {
	defer func() {
		if p := recover(); p != nil {
			CloseOrPanic(dbConnection)
			err = util.Errorf("panicked %s\n%s", p, debug.Stack())
		}
	}()

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbInfo.Host,
		dbInfo.Port,
		dbInfo.User,
		dbInfo.Password,
		dbInfo.DbName)

	dbConnection, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, util.Wrap("could not open db connection", err)
	}

	if err = CheckConnection(log, dbConnection); err != nil {
		CloseOrPanic(dbConnection)
		return nil, util.Wrap("could not check db connection", err)
	}

	return dbConnection, nil
}
