package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/caarlos0/env/v6"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	dm "user-manager/domain-model"
	email "user-manager/third-party-models/email-api"
	"user-manager/util/command"
	"user-manager/util/db"
	"user-manager/util/errs"
	"user-manager/util/logger"
)

type Config struct {
	DbInfo      db.Info
	EmailApiUrl string `env:"EMAIL_API_URL"`
	Environment string `env:"ENVIRONMENT"`
}

func main() {
	slog.SetDefault(logger.NewLogger(false).With("service", "app"))
	command.Run(startJob)
}

func startJob() error {
	slog.Info("Starting up")

	config := Config{}
	if err := env.Parse(&config, env.Options{RequiredIfNoDef: true}); err != nil {
		return errs.Wrap("error parsing env", err)
	}

	if config.Environment != "local" {
		slog.SetDefault(logger.NewLogger(true).With("service", "app"))
		gin.SetMode(gin.ReleaseMode)
	}

	database, err := db.OpenDbConnection(config.DbInfo)
	if err != nil {
		return errs.Wrap("issue opening db connection", err)
	}
	defer db.CloseOrPanic(database.Client())

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)

	// Run until shutdown signal or server error is received
	var lastEmailSentAt time.Time
	minTimeBetweenSendingEmails := 2 * time.Second
	for {
		select {
		case <-signals:
			slog.Info("Shutdown signal received. About to shut down")
			return nil
		default:
			timeSinceLastEmailSent := time.Since(lastEmailSentAt)
			if timeSinceLastEmailSent < minTimeBetweenSendingEmails {
				time.Sleep(minTimeBetweenSendingEmails - timeSinceLastEmailSent)
			}
			if err = sendOneEmail(database, config); err != nil {
				return errs.Wrap("issue sending email", err)
			}
			lastEmailSentAt = time.Now()
		}
	}
}

func sendOneEmail(database *mongo.Database, config Config) (ret error) {
	maxNumFailedAttempts := int8(3)
	ctx, cancelTimeout := db.DefaultQueryContext(context.Background())
	defer cancelTimeout()

	var mail dm.Mail
	err := database.Collection(dm.MailQueueCollectionName).FindOne(ctx, bson.M{
		"$or": []bson.M{
			{"status": dm.MailStatusPending},
			{"status": dm.MailStatusFailed, "failedAttempts": bson.M{"$lt": maxNumFailedAttempts}},
		},
	}).Decode(&mail)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			slog.Info("No emails to send")
			return nil
		}
		return errs.Wrap("issue getting email from db", err)
	}

	payload, err := json.Marshal(email.EmailTO{
		From:    mail.From,
		To:      mail.To,
		Subject: mail.Subject,
		Body:    mail.Body,
	})

	if err != nil {
		return errs.Wrap("issue marshalling payload for api call", err)
	}
	// TODO: how does this work in GCP?
	_, err = http.Post(config.EmailApiUrl, "application/json", bytes.NewReader(payload))

	// If sending the mail failed, log and continue
	failedAttempts := mail.FailedAttempts
	status := dm.MailStatusSent
	if err != nil {
		status = dm.MailStatusFailed
		failedAttempts++
		slog.Warn(errs.Wrap("issue sending email", err).Error())
	}

	_, err = database.Collection(dm.MailQueueCollectionName).UpdateByID(ctx, mail.ObjectID, bson.M{
		"$set": bson.M{
			"status":         status,
			"failedAttempts": failedAttempts,
			"updatedAt":      time.Now(),
		},
	})

	if err != nil {
		return errs.Wrap("issue updating email in db", err)
	}

	return nil
}
