package config

import (
	"database/sql"
	"fmt"
	"runtime/debug"
	"user-manager/util"
)

func (config *Config) GetPsqlInfo() (string, error) {
	dbInfo := config.DbInfo
	host := dbInfo.Host
	if host == "" {
		return "", util.Error("GetPsqlInfo", "no host defined")
	}
	port := dbInfo.Port
	if port == "" {
		return "", util.Error("GetPsqlInfo", "no port defined")
	}
	user := dbInfo.User
	if user == "" {
		return "", util.Error("GetPsqlInfo", "no user defined")
	}
	password := dbInfo.Password
	if password == "" {
		return "", util.Error("GetPsqlInfo", "no password defined")
	}
	dbname := dbInfo.DbName
	if dbname == "" {
		return "", util.Error("GetPsqlInfo", "no dbname defined")
	}

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname), nil
}

func (config *Config) OpenDbConnection(log util.Logger) (dbConnection *sql.DB, err error) {
	defer func() {
		if p := recover(); p != nil {
			util.CloseOrPanic(dbConnection)
			err = util.Errorf("OpenDbConnection", "panicked %s\n%s", p, debug.Stack())
		}
	}()
	psqlInfo, err := config.GetPsqlInfo()
	if err != nil {
		return nil, util.Wrap("OpenDbConnection", "issue in database connection information", err)
	}

	dbConnection, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, util.Wrap("OpenDbConnection", "could not open db connection", err)
	}

	err = util.CheckConnection(log, dbConnection)
	if err != nil {
		util.CloseOrPanic(dbConnection)
		return nil, util.Wrap("OpenDbConnection", "could not check db connection", err)
	}

	return dbConnection, nil
}
