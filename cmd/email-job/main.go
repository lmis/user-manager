package main

//go:generate go run ../generate-sqlboiler/main.go ../../db/generated/models
import (
	"bytes"
	"encoding/json"
	"net/http"
	config "user-manager/cmd/email-job/config"
	"user-manager/db"
	"user-manager/db/generated/models"
	emailapi "user-manager/third-party-models/email-api"
	"user-manager/util"

	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"database/sql"

	_ "github.com/lib/pq"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func main() {
	util.Run("EMAILER", startJob)
}

func startJob(log util.Logger) error {
	log.Info("Starting up")

	var db *sql.DB

	config, err := config.GetConfig(log)
	if err != nil {
		return util.Wrap("issue reading config", err)
	}

	db, err = config.DbInfo.OpenDbConnection(log)
	if err != nil {
		return util.Wrap("issue opening db connection", err)
	}
	defer util.CloseOrPanic(db)

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
			if err = sendOneEmail(log, db, config); err != nil {
				return util.Wrap("issue sending email", err)
			}
			lastEmailSentAt = time.Now()
		}
	}
}

func sendOneEmail(log util.Logger, database *sql.DB, config *config.Config) (ret error) {
	shouldCommit := false
	var maxNumFailedAttempts int16 = 3
	ctx, cancelTimeout := db.DefaultQueryContext()
	defer cancelTimeout()

	log.Info("BEGIN Transaction")
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return util.Wrap("issue beginning transaction", err)
	}
	defer func() {
		if shouldCommit {
			log.Info("COMMIT")
			if err := tx.Commit(); err != nil {
				ret = util.Wrap("issue committing to db", err)
			}
		} else {
			log.Info("ROLLBACK Transaction")
			if err = tx.Rollback(); err != nil {
				ret = util.Wrap("issue rolling back transaction", err)
			}
		}
	}()

	mail, err := models.MailQueues(
		models.MailQueueWhere.Status.EQ(models.EmailStatusPENDING),
		qm.Or2(qm.Expr(models.MailQueueWhere.Status.EQ(models.EmailStatusERROR), models.MailQueueWhere.NumberOfFailedAttempts.LT(maxNumFailedAttempts))),
		qm.OrderBy(models.MailQueueColumns.Priority+" DESC"),
	).One(ctx, tx)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return util.Wrap("issue getting email from db", err)
	}

	payload, err := json.Marshal(emailapi.EmailTO{
		From:    mail.FromAddress,
		To:      mail.ToAddress,
		Subject: mail.Subject,
		Body:    mail.Content,
	})

	if err != nil {
		return util.Wrap("issue marshalling payload for api call", err)
	}
	// TODO: how does this work in GCP?
	_, err = http.Post(config.EmailApiUrl, "application/json", bytes.NewReader(payload))

	// If sending the mail failed, log and continue
	if err != nil {
		mail.Status = models.EmailStatusERROR
		mail.NumberOfFailedAttempts++
		log.Warn(util.Wrap("issue sending email", err).Error())
	} else {
		mail.Status = models.EmailStatusSENT
	}

	rows, err := mail.Update(ctx, tx, boil.Infer())
	if err != nil {
		return util.Wrap("issue updating email in db", err)
	}
	if rows != 1 {
		return util.Wrap(fmt.Sprintf("wrong number of rows affected: %d", rows), err)
	}

	shouldCommit = true
	return nil
}
