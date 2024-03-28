package main

import (
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"user-manager/cmd/app/router/middleware"
	emailapi "user-manager/third-party-models/email-api"
	"user-manager/util/command"
	"user-manager/util/errs"
	httputil "user-manager/util/http"
	"user-manager/util/logger"
)

type Emails map[string][]emailapi.EmailTO

type Config struct {
	Port string `env:"MOCK_API_PORT"`
}

func main() {
	slog.SetDefault(logger.NewLogger(false))
	command.Run(startServer)
}

func startServer() error {
	emails := make(Emails)
	slog.Info("Starting up")

	config := &Config{}
	if err := env.Parse(config, env.Options{RequiredIfNoDef: true}); err != nil {
		return errs.Wrap("error parsing env", err)
	}

	app := gin.New()
	app.Use(middleware.RecoveryMiddleware)

	app.POST("/mock-send-email", func(c *gin.Context) {
		var mail emailapi.EmailTO
		if err := c.BindJSON(&mail); err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, errs.Wrap("cannot bind to EmailTO", err))
			return
		}
		m, ok := emails[mail.To]
		if !ok {
			emails[mail.To] = []emailapi.EmailTO{mail}
		} else {
			emails[mail.To] = append(m, mail)
		}
		slog.Info("Email received", "mail", fmt.Sprintf("%v", mail))
	})

	app.GET("/mock-emails/:address", func(c *gin.Context) {
		address := c.Param("address")
		subjectQuery := c.Query("subject")
		slog.Info("Querying emails", "address", address, "subjectQuery", subjectQuery, "emails", fmt.Sprintf("%v", emails))
		var mails []emailapi.EmailTO
		for _, mail := range emails[address] {
			if subjectQuery != "" && mail.Subject != subjectQuery {
				continue
			}

			mails = append(mails, mail)
		}

		c.JSON(http.StatusOK, mails)

	})

	if err := httputil.RunHttpServer(&http.Server{
		Addr:    ":" + config.Port,
		Handler: app,
	}); err != nil {
		return errs.Wrap("issue running http server", err)
	}

	return nil
}
