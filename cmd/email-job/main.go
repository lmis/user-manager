package main

import (
	"bytes"
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"user-manager/cmd/email-job/config"
	"user-manager/db"
	dm "user-manager/domain-model"
	email "user-manager/third-party-models/email-api"
	"user-manager/util/command"
	"user-manager/util/errors"
	"user-manager/util/logger"

	_ "github.com/lib/pq"
)

func main() {
	command.Run("EMAILER", startJob)
}

func startJob(log logger.Logger) error {
	log.Info("Starting up")

	conf, err := config.GetConfig()
	if err != nil {
		return errors.Wrap("issue reading config", err)
	}

	database, err := db.OpenDbConnection(log, conf.DbInfo)
	if err != nil {
		return errors.Wrap("issue opening db connection", err)
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
			log.Info("Shutdown signal received. About to shut down")
			return nil
		default:
			timeSinceLastEmailSent := time.Since(lastEmailSentAt)
			if timeSinceLastEmailSent < minTimeBetweenSendingEmails {
				time.Sleep(minTimeBetweenSendingEmails - timeSinceLastEmailSent)
			}
			if err = sendOneEmail(log, database, conf); err != nil {
				return errors.Wrap("issue sending email", err)
			}
			lastEmailSentAt = time.Now()
		}
	}
}

func sendOneEmail(log logger.Logger, database *mongo.Database, config *config.Config) (ret error) {
	maxNumFailedAttempts := int8(3)
	ctx, cancelTimeout := db.DefaultQueryContext(context.Background())
	defer cancelTimeout()

	var mail dm.Mail
	err := database.Collection(dm.MailQueueCollectionName).FindOne(ctx, bson.M{
		"$or": []bson.M{
			{"status": dm.MailStatusPending},
			{"status": dm.MailStatusFailed, "numberOfFailedAttempts": bson.M{"$lt": maxNumFailedAttempts}},
		},
	}).Decode(&mail)

	if err != nil {
		return errors.Wrap("issue getting email from db", err)
	}

	payload, err := json.Marshal(email.EmailTO{
		From:    mail.FromAddress,
		To:      mail.ToAddress,
		Subject: mail.Subject,
		Body:    mail.Content,
	})

	if err != nil {
		return errors.Wrap("issue marshalling payload for api call", err)
	}
	// TODO: how does this work in GCP?
	_, err = http.Post(config.EmailApiUrl, "application/json", bytes.NewReader(payload))

	// If sending the mail failed, log and continue
	numberOfFailedAttempts := mail.NumberOfFailedAttempts
	status := dm.MailStatusSent
	if err != nil {
		status = dm.MailStatusFailed
		numberOfFailedAttempts++
		log.Warn(errors.Wrap("issue sending email", err).Error())
	}

	_, err = database.Collection(dm.MailQueueCollectionName).UpdateByID(ctx, mail.ID, bson.M{
		"$set": bson.M{
			"status":                 status,
			"numberOfFailedAttempts": numberOfFailedAttempts,
			"updatedAt":              time.Now(),
		},
	})

	if err != nil {
		return errors.Wrap("issue updating email in db", err)
	}

	return nil
}
