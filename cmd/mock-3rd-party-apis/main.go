package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"user-manager/cmd/app/router/middleware"
	"user-manager/cmd/mock-3rd-party-apis/config"
	test "user-manager/cmd/mock-3rd-party-apis/functional-tests"
	"user-manager/cmd/mock-3rd-party-apis/util"
	email "user-manager/third-party-models/email-api"
	"user-manager/util/command"
	"user-manager/util/errs"
	httputil "user-manager/util/http"
	"user-manager/util/logger"
	"user-manager/util/random"

	"github.com/gin-gonic/gin"
)

func main() {
	slog.SetDefault(logger.NewLogger(false).With("service", "mock-3rd-party-apis"))
	command.Run(startServer)
}

func startServer() error {
	emails := make(util.Emails)
	slog.Info("Starting up")
	conf, err := config.GetConfig()
	if err != nil {
		return errs.Wrap("cannot read config", err)
	}

	app := gin.New()
	app.Use(middleware.RecoveryMiddleware)

	registerMockEmailApi(app, emails)
	registerFunctionalTests(conf, app, emails)

	if err = httputil.RunHttpServer(&http.Server{
		Addr:    ":" + conf.Port,
		Handler: app,
	}); err != nil {
		return errs.Wrap("issue running http server", err)
	}

	return nil
}

func registerMockEmailApi(app *gin.Engine, emails util.Emails) {
	app.POST("/mock-send-email", func(c *gin.Context) {
		var mail email.EmailTO
		if err := c.BindJSON(&mail); err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, errs.Wrap("cannot bind to EmailTO", err))
			return
		}
		m, ok := emails[mail.To]
		if !ok {
			emails[mail.To] = []email.EmailTO{mail}
		} else {
			emails[mail.To] = append(m, mail)
		}
		slog.Info("Email received", "mail", fmt.Sprintf("%v", mail))
	})
}

func registerFunctionalTests(config *config.Config, app *gin.Engine, emails util.Emails) {
	testUser := util.TestUser{}
	tests := []util.FunctionalTest{
		{
			Description: "Sign-up",
			Test:        test.TestSignUp,
		},
		{
			Description: "Password reset",
			Test:        test.TestPasswordReset,
		},
		{
			Description: "CSRF",
			Test:        test.TestCallWithMismatchingCsrfTokens,
		},
		{
			Description: "Simple login",
			Test:        test.TestSimpleLogin,
		},
		{
			Description: "Change password",
			Test:        test.TestChangePassword,
		},
	}
	app.GET("/tests/:n", func(c *gin.Context) {
		n := c.Param("n")
		testNumber, err := strconv.Atoi(n)
		if err != nil || testNumber < 0 || testNumber >= len(tests) {
			c.Status(http.StatusNotFound)
			return
		}
		c.String(http.StatusOK, tests[testNumber].Description)
	})
	app.POST("/tests/reset", func(c *gin.Context) {
		testUser = util.TestUser{
			Email:    "test-user-" + random.MakeRandomURLSafeB64(5) + "@example.com",
			Password: []byte("hunter12"),
		}
	})
	app.POST("/tests/:n/trigger", func(c *gin.Context) {
		n := c.Param("n")
		testNumber, err := strconv.Atoi(n)
		if err != nil || testNumber < 0 || testNumber >= len(tests) {
			c.Status(http.StatusNotFound)
			return
		}
		if err := tests[testNumber].Test(config, emails, &testUser); err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		c.Status(http.StatusNoContent)
	})
}
