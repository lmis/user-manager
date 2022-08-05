package db

import (
	"context"
	"database/sql"
	"time"
	"user-manager/util"
)

func CloseOrPanic(connection *sql.DB) {
	err := connection.Close()
	if err != nil {
		panic(util.Wrap("closing db failed", err))
	}
}

func CheckConnection(log util.Logger, connection *sql.DB) error {
	numAttempts := 10
	sleepTime := 500 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(numAttempts)*sleepTime+1*time.Second)
	defer cancel()
	log.Info("Pinging DB")
	err := connection.PingContext(ctx)
	for attempts := 1; err != nil && attempts < numAttempts; attempts++ {
		time.Sleep(sleepTime)
		log.Info("Retry pinging DB")
		err = connection.PingContext(ctx)
	}
	if err != nil {
		return util.Wrap("issue pinging db", err)
	}

	return nil
}
