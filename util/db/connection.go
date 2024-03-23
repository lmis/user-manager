package db

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log/slog"
	"runtime/debug"
	"time"
	"user-manager/util/errs"
)

type Info struct {
	Name     string `env:"DB_NAME"`
	Host     string `env:"DB_HOST"`
	Port     int    `env:"DB_PORT"`
	User     string `env:"DB_USER"`
	Password string `env:"DB_PASSWORD"`
}

func OpenDbConnection(info Info) (_ *mongo.Database, err error) {
	clientOptions := options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%d", info.Host, info.Port)).SetAuth(options.Credential{
		Username: info.User,
		Password: info.Password,
	})

	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, errs.Wrap("could not open db connection", err)
	}
	defer func() {
		if p := recover(); p != nil {
			CloseOrPanic(client)
			err = errs.Errorf("panicked %s\n%s", p, debug.Stack())
		}
	}()

	if err = CheckConnection(client); err != nil {
		CloseOrPanic(client)
		return nil, errs.Wrap("could not check db connection", err)
	}

	return client.Database(info.Name), nil
}
func CloseOrPanic(client *mongo.Client) {
	if client == nil {
		err := client.Disconnect(context.Background())

		if err != nil {
			panic(errs.Wrap("closing db failed", err))
		}
	}
}

func CheckConnection(client *mongo.Client) error {
	numAttempts := 10
	sleepTime := 500 * time.Millisecond
	ctx := context.Background() //context.WithTimeout(context.Background(), time.Duration(numAttempts)*sleepTime+1*time.Second)
	//defer cancel()

	slog.Info("Pinging DB")
	err := client.Ping(ctx, nil)
	for attempts := 1; err != nil && attempts < numAttempts; attempts++ {
		time.Sleep(sleepTime)
		slog.Info("Retry pinging DB")
		err = client.Ping(ctx, nil)
	}
	if err != nil {
		return errs.Wrap("issue pinging db", err)
	}

	slog.Info("Pinging DB successful")
	return nil
}
