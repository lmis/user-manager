package db

import (
	"context"
	"database/sql"
	"time"
	"user-manager/util/errors"
	"user-manager/util/logger"
)

func CloseOrPanic(connection *sql.DB) {
	err := connection.Close()
	if err != nil {
		panic(errors.Wrap("closing db failed", err))
	}
}

func CheckConnection(log logger.Logger, connection *sql.DB) error {
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
		return errors.Wrap("issue pinging db", err)
	}

	return nil
}
