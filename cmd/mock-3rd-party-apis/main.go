package main

//go:generate go run ../generate-sqlboiler/main.go ../../db/generated/models
import (
	"net/http"
	"user-manager/cmd/app/middlewares"
	config "user-manager/cmd/mock-3rd-party-apis/config"
	flowtests "user-manager/cmd/mock-3rd-party-apis/flow-tests"
	emailapi "user-manager/third-party-models/email-api"
	"user-manager/util"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	util.Run("MOCK 3RD-PARTY APIS", startServer)
}

func startServer(log util.Logger) error {
	emails := make(map[string][]emailapi.EmailTO)
	log.Info("Starting up")
	config, err := config.GetConfig(log)
	if err != nil {
		return util.Wrap("cannot read config", err)
	}

	app := gin.New()
	app.Use(middlewares.RecoveryMiddleware)
	app.POST("/mock-send-email", func(c *gin.Context) {
		var email emailapi.EmailTO
		if err := c.BindJSON(&email); err != nil {
			c.AbortWithError(http.StatusBadRequest, util.Wrap("cannot bind to EmailTO", err))
			return
		}
		m, ok := emails[email.To]
		if !ok {
			emails[email.To] = []emailapi.EmailTO{email}
		} else {
			emails[email.To] = append(m, email)
		}
		log.Info("Email received %v", email)
	})

	app.POST("/trigger-test/:n", func(c *gin.Context) {
		n := c.Param("n")
		switch n {
		case "1":
			respondToTestRequest(c, flowtests.TestRoleBeforeSignup(config))
		case "2":
			respondToTestRequest(c, flowtests.TestSignUp(config, emails))
		default:
			c.Status(http.StatusNotFound)
		}
	})

	if err = util.RunHttpServer(log, &http.Server{
		Addr:    ":" + config.Port,
		Handler: app,
	}); err != nil {
		return util.Wrap("issue running http server", err)
	}

	return nil
}

func respondToTestRequest(c *gin.Context, err error) {
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	} else {
		c.Status(http.StatusNoContent)
	}
}
