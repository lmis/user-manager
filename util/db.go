package util

import (
	"context"
	"database/sql"
	"time"
)

func CloseOrPanic(db *sql.DB) {
	err := db.Close()
	if err != nil {
		panic(Wrap("closing db failed", err))
	}
}

func CheckConnection(log Logger, db *sql.DB) error {
	numAttempts := 10
	sleepTime := 500 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(numAttempts)*sleepTime+1*time.Second)
	defer cancel()
	log.Info("Pinging DB")
	err := db.PingContext(ctx)
	for attempts := 1; err != nil && attempts < numAttempts; attempts++ {
		time.Sleep(sleepTime)
		log.Info("Retry pinging DB")
		err = db.PingContext(ctx)
	}
	if err != nil {
		return Wrap("issue pinging db", err)
	}

	return nil
}
