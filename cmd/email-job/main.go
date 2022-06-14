package main

//go:generate go run ../generate-sqlboiler/main.go ../../db/generated/models
import (
	"user-manager/db"
	"user-manager/db/generated/models"
	"user-manager/util"

	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"database/sql"

	_ "github.com/lib/pq"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type dbInfo struct {
	dbname   string
	host     string
	port     string
	user     string
	password string
}

func main() {
	util.SetLogJSON(os.Getenv("LOG_JSON") != "")
	log := util.Log("EMAILER")
	exitCode := 0

	defer func() {
		if p := recover(); p != nil {
			stack := string(debug.Stack())
			log.Err(fmt.Errorf("panic: %v\n%v", p, stack))
			exitCode = 1
		}
		log.Info("Shutdown complete. ExitCode: %d", exitCode)
		os.Exit(exitCode)
	}()

	if err := runServer(log); err != nil {
		log.Err(err)
		log.Warn("Server exited abnormally")
		exitCode = 1
	}
}

func runServer(log util.Logger) error {
	log.Info("Starting up")

	var dbConnection *sql.DB

	port := util.GetEnvOrDefault(log, "PORT", "8080")
	environment := util.GetEnvOrDefault(log, "ENVIRONMENT", "local")

	dbInfo := &dbInfo{
		dbname:   os.Getenv("DB_NAME"),
		host:     os.Getenv("DB_HOST"),
		port:     os.Getenv("DB_PORT"),
		user:     os.Getenv("DB_USER"),
		password: os.Getenv("DB_PASSWORD"),
	}

	if dbInfo.dbname == "" || dbInfo.host == "" || dbInfo.port == "" || dbInfo.user == "" || dbInfo.password == "" {
		return fmt.Errorf("missing database connection information")
	}

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbInfo.host, dbInfo.port, dbInfo.user, dbInfo.password, dbInfo.dbname)
	dbConnection, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return err
	}
	defer util.CloseOrPanic(dbConnection)
	numAttempts := 10
	sleepTime := 500 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(numAttempts)*sleepTime+1*time.Second)
	defer cancel()
	log.Info("Pinging DB")
	err = dbConnection.PingContext(ctx)
	for attempts := 1; err != nil && attempts < numAttempts; attempts++ {
		time.Sleep(sleepTime)
		log.Info("Retry pinging DB")
		err = dbConnection.PingContext(ctx)
	}
	if err != nil {
		return err
	}

	log.Info("Starting http server on port %s", port)

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
			err = sendOneEmail(log, dbConnection, environment == "prod")
			lastEmailSentAt = time.Now()
			if err != nil {
				return err
			}
		}
	}
}

func sendOneEmail(log util.Logger, database *sql.DB, sendRealEmails bool) error {
	var maxNumFailedAttempts int16 = 3
	ctx, cancelTimeout := db.DefaultQueryContext()
	defer cancelTimeout()

	log.Info("BEGIN Transaction")
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	mail, err := models.MailQueues(
		models.MailQueueWhere.Status.EQ(models.EmailStatusPENDING),
		qm.Or2(qm.Expr(models.MailQueueWhere.Status.EQ(models.EmailStatusERROR), models.MailQueueWhere.NumberOfFailedAttempts.LT(maxNumFailedAttempts))),
		qm.OrderBy(models.MailQueueColumns.Priority+"desc"),
	).One(ctx, tx)

	if err != nil {
		return err
	}

	// send
	err = nil // todo

	if err != nil {
		mail.Status = models.EmailStatusERROR
		mail.NumberOfFailedAttempts++
		log.Err(err)
	} else {
		mail.Status = models.EmailStatusSENT
	}

	rows, err := mail.Update(ctx, tx, boil.Infer())
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("wrong number of rows affected: %d", rows)
	}

	log.Info("COMMIT")
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
