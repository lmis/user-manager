package db

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"runtime/debug"
	"time"
	"user-manager/util/errors"
	"user-manager/util/logger"
)

type DbInfo struct {
	Name     string `env:"DB_NAME"`
	Host     string `env:"DB_HOST"`
	Port     int    `env:"DB_PORT"`
	User     string `env:"DB_USER"`
	Password string `env:"DB_PASSWORD"`
}

func OpenDbConnection(log logger.Logger, info DbInfo) (_ *mongo.Database, err error) {
	clientOptions := options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%d", info.Host, info.Port)).SetAuth(options.Credential{
		Username: info.User,
		Password: info.Password,
	})

	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, errors.Wrap("could not open db connection", err)
	}
	defer func() {
		if p := recover(); p != nil {
			CloseOrPanic(client)
			err = errors.Errorf("panicked %s\n%s", p, debug.Stack())
		}
	}()

	if err = CheckConnection(log, client); err != nil {
		CloseOrPanic(client)
		return nil, errors.Wrap("could not check db connection", err)
	}

	return client.Database(info.Name), nil
}
func CloseOrPanic(client *mongo.Client) {
	if client == nil {
		err := client.Disconnect(context.Background())

		if err != nil {
			panic(errors.Wrap("closing db failed", err))
		}
	}
}

func CheckConnection(log logger.Logger, client *mongo.Client) error {
	numAttempts := 10
	sleepTime := 500 * time.Millisecond
	ctx := context.Background() //context.WithTimeout(context.Background(), time.Duration(numAttempts)*sleepTime+1*time.Second)
	//defer cancel()

	log.Info("Pinging DB")
	err := client.Ping(ctx, nil)
	for attempts := 1; err != nil && attempts < numAttempts; attempts++ {
		time.Sleep(sleepTime)
		log.Info("Retry pinging DB")
		err = client.Ping(ctx, nil)
	}
	if err != nil {
		return errors.Wrap("issue pinging db", err)
	}

	return nil
}
