package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	config "user-manager/cmd/email-job/config"
	"user-manager/db"
	. "user-manager/db/generated/models/postgres/public/enum"
	"user-manager/db/generated/models/postgres/public/model"
	. "user-manager/db/generated/models/postgres/public/table"
	emailapi "user-manager/third-party-models/email-api"
	"user-manager/util/command"
	"user-manager/util/errors"
	"user-manager/util/logger"

	. "github.com/go-jet/jet/v2/postgres"
	_ "github.com/lib/pq"
)

func main() {
	command.Run("EMAILER", startJob)
}

func startJob(log logger.Logger) error {
	log.Info("Starting up")

	config, err := config.GetConfig(log)
	if err != nil {
		return errors.Wrap("issue reading config", err)
	}

	connection, err := config.DbInfo.OpenDbConnection(log)
	if err != nil {
		return errors.Wrap("issue opening db connection", err)
	}
	defer db.CloseOrPanic(connection)

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
			if err = sendOneEmail(log, connection, config); err != nil {
				return errors.Wrap("issue sending email", err)
			}
			lastEmailSentAt = time.Now()
		}
	}
}

func sendOneEmail(log logger.Logger, database *sql.DB, config *config.Config) (ret error) {
	shouldCommit := false
	maxNumFailedAttempts := int16(3)
	ctx, cancelTimeout := db.DefaultQueryContext()
	defer cancelTimeout()

	log.Info("BEGIN Transaction")
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap("issue beginning transaction", err)
	}
	defer func() {
		if shouldCommit {
			log.Info("COMMIT")
			if err := tx.Commit(); err != nil {
				ret = errors.Wrap("issue committing to db", err)
			}
		} else {
			log.Info("ROLLBACK Transaction")
			if err = tx.Rollback(); err != nil {
				ret = errors.Wrap("issue rolling back transaction", err)
			}
		}
	}()

	maybeMail, err := db.Fetch(
		SELECT(MailQueue.AllColumns).
			FROM(MailQueue).
			WHERE(
				MailQueue.Status.EQ(EmailStatus.Pending).
					OR(
						MailQueue.Status.EQ(EmailStatus.Error).
							AND(MailQueue.NumberOfFailedAttempts.LT(Int16(maxNumFailedAttempts))))).
			ORDER_BY(MailQueue.Priority.DESC()).
			QueryContext,
		func(x *model.MailQueue) *model.MailQueue { return x },
		tx)

	if err != nil {
		return errors.Wrap("issue getting email from db", err)
	}
	if maybeMail.IsEmpty() {
		return
	}

	mail := maybeMail.OrPanic()
	payload, err := json.Marshal(emailapi.EmailTO{
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
	status := EmailStatus.Sent
	if err != nil {
		status = EmailStatus.Error
		numberOfFailedAttempts++
		log.Warn(errors.Wrap("issue sending email", err).Error())
	}

	err = db.ExecSingleMutation(
		MailQueue.UPDATE(MailQueue.Status, MailQueue.NumberOfFailedAttempts, MailQueue.UpdatedAt).
			SET(status, numberOfFailedAttempts, time.Now()).
			WHERE(MailQueue.MailQueueID.EQ(Int64(mail.MailQueueID))).
			ExecContext,
		tx)

	if err != nil {
		return errors.Wrap("issue updating email in db", err)
	}

	shouldCommit = true
	return nil
}
