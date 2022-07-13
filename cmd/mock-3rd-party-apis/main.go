package main

//go:generate go run ../generate-sqlboiler/main.go ../../db/generated/models
import (
	"net/http"
	"user-manager/cmd/app/middlewares"
	config "user-manager/cmd/mock-3rd-party-apis/config"
	flowtests "user-manager/cmd/mock-3rd-party-apis/flow-tests"
	"user-manager/util"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	util.Run("MOCK 3RD-PARTY APIS", startServer)
}

type EmailTO struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func startServer(log util.Logger) error {
	emails := make(map[string][]EmailTO)
	log.Info("Starting up")
	config, err := config.GetConfig(log)
	if err != nil {
		return util.Wrap("cannot read config", err)
	}

	app := gin.New()
	app.Use(middlewares.RecoveryMiddleware)
	app.POST("/mock-send-email", func(c *gin.Context) {
		var email EmailTO
		if err := c.BindJSON(&email); err != nil {
			c.AbortWithError(http.StatusBadRequest, util.Wrap("cannot bind to EmailTO", err))
			return
		}
		m, ok := emails[email.To]
		if !ok {
			emails[email.To] = []EmailTO{email}
		} else {
			emails[email.To] = append(m, email)
		}
	})

	app.POST("/trigger-test/:n", func(c *gin.Context) {
		n := c.Param("n")
		switch n {
		case "1":
			respondToTestRequest(c, flowtests.TestRoleBeforeSignup(config))
			return
		default:
			c.Status(http.StatusNotFound)
			return
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
