package main

// TODO: Extract and reuse common parts  from server/main.go

//go:generate go run ../generate-sqlboiler/main.go ../../db/generated/models
import (
	"user-manager/config"
	"user-manager/db"
	"user-manager/db/generated/models"
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
		return util.Wrap("startJob", "issue reading config", err)
	}

	db, err = config.OpenDbConnection(log)
	if err != nil {
		return util.Wrap("startJob", "issue opening db connection", err)
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
			err = sendOneEmail(log, db, config)
			lastEmailSentAt = time.Now()
			if err != nil {
				return util.Wrap("startJob", "issue sending email", err)
			}
		}
	}
}

func sendOneEmail(log util.Logger, database *sql.DB, config *config.Config) error {
	var maxNumFailedAttempts int16 = 3
	ctx, cancelTimeout := db.DefaultQueryContext()
	defer cancelTimeout()

	log.Info("BEGIN Transaction")
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return util.Wrap("sendOneEmail", "issue beginning transaction", err)
	}
	defer func() {
		log.Info("ROLLBACK Transaction")
		err = tx.Rollback()
		// TODO: wrap?
		if err != nil {
			panic(err)
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
		return util.Wrap("sendOneEmail", "issue getting email from db", err)
	}

	// send
	if config.IsLocalEnv() {
		log.Info("%s: %s", mail.Email, mail.Content)
	} else {
		// address := "test-recipient@example.com" // TODO
		// if (config.IsProdEnv()) {
		// 	address = mail.Email
		// }
		// TODO: call some api
	}

	// If sending the mail failed, log and continue
	if err != nil {
		mail.Status = models.EmailStatusERROR
		mail.NumberOfFailedAttempts++
		log.Warn(util.Wrap("sendOneEmail", "issue sending email", err).Error())
	} else {
		mail.Status = models.EmailStatusSENT
	}

	rows, err := mail.Update(ctx, tx, boil.Infer())
	if err != nil {
		return util.Wrap("sendOneEmail", "issue updating email in db", err)
	}
	if rows != 1 {
		return util.Wrap("sendOneEmail", fmt.Sprintf("wrong number of rows affected: %d", rows), err)
	}

	log.Info("COMMIT")
	if err := tx.Commit(); err != nil {
		return util.Wrap("sendOneEmail", "issue committing to db", err)
	}

	return nil
}
