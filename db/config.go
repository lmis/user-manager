package db

//go:generate go run ../cmd/migrator/main.go generate generated/models
import (
	"database/sql"
	"fmt"
	"runtime/debug"
	"user-manager/util/errors"
	"user-manager/util/logger"
)

type Info struct {
	DbName   string `env:"DB_NAME"`
	Host     string `env:"DB_HOST"`
	Port     int    `env:"DB_PORT"`
	User     string `env:"DB_USER"`
	Password string `env:"DB_PASSWORD"`
}

func (dbInfo *Info) OpenDbConnection(log logger.Logger) (dbConnection *sql.DB, err error) {
	defer func() {
		if p := recover(); p != nil {
			CloseOrPanic(dbConnection)
			err = errors.Errorf("panicked %s\n%s", p, debug.Stack())
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
		return nil, errors.Wrap("could not open db connection", err)
	}

	if err = CheckConnection(log, dbConnection); err != nil {
		CloseOrPanic(dbConnection)
		return nil, errors.Wrap("could not check db connection", err)
	}

	return dbConnection, nil
}
