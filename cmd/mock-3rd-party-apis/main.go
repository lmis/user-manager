package main

//go:generate go run ../generate-sqlboiler/main.go ../../db/generated/models
import (
	"net/http"
	"strconv"
	middleware "user-manager/cmd/app/middlewares"
	config "user-manager/cmd/mock-3rd-party-apis/config"
	functional_tests "user-manager/cmd/mock-3rd-party-apis/functional-tests"
	mock_util "user-manager/cmd/mock-3rd-party-apis/util"
	email_api "user-manager/third-party-models/email-api"
	"user-manager/util"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	util.Run("MOCK 3RD-PARTY APIS", startServer)
}

func startServer(log util.Logger, dir string) error {
	emails := make(mock_util.Emails)
	testUser := mock_util.TestUser{
		Email:    "test-user-" + util.MakeRandomURLSafeB64(3) + "@example.com",
		Password: []byte("hunter12"),
	}
	log.Info("Starting up")
	config, err := config.GetConfig(log)
	if err != nil {
		return util.Wrap("cannot read config", err)
	}

	app := gin.New()
	app.Use(middleware.RecoveryMiddleware)
	app.POST("/mock-send-email", func(c *gin.Context) {
		var email email_api.EmailTO
		if err := c.BindJSON(&email); err != nil {
			c.AbortWithError(http.StatusBadRequest, util.Wrap("cannot bind to EmailTO", err))
			return
		}
		m, ok := emails[email.To]
		if !ok {
			emails[email.To] = []*email_api.EmailTO{&email}
		} else {
			emails[email.To] = append(m, &email)
		}
		log.Info("Email received %v", email)
	})

	tests := []mock_util.FunctionalTest{
		{
			Description: "Role before sign-up",
			Test:        functional_tests.TestUserEndpointBeforeSignup,
		},
		{
			Description: "Sign-up",
			Test:        functional_tests.TestSignUp,
		},
		{
			Description: "CSRF",
			Test:        functional_tests.TestCallWithMismatchingCsrfTokens,
		},
		{
			Description: "Bad login",
			Test:        functional_tests.TestBadLogin,
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
		testUser = mock_util.TestUser{
			Email:    "test-user-" + util.MakeRandomURLSafeB64(3) + "@example.com",
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

	if err = util.RunHttpServer(log, &http.Server{
		Addr:    ":" + config.Port,
		Handler: app,
	}); err != nil {
		return util.Wrap("issue running http server", err)
	}

	return nil
}
