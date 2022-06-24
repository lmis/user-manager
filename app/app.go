package app

import (
	"user-manager/app/endpoints"
	"user-manager/app/middlewares"
	"user-manager/util"

	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func recoveryHandler(c *gin.Context, requestErr interface{}) {
	err, ok := requestErr.(error)
	if !ok {
		err = fmt.Errorf("%v", requestErr)
	}
	c.AbortWithError(http.StatusInternalServerError, util.Wrap("recoveryHandler", "recovered from panic", err))
}

func New(db *sql.DB, environment string) *gin.Engine {
	if db == nil {
		panic("Invalid gin engine construction: db is nil")
	}

	r := gin.New()
	r.Use(middlewares.LoggerMiddleware)
	r.Use(gin.CustomRecovery(recoveryHandler))

	{
		api := r.Group("api")
		api.Use(middlewares.DatabaseMiddleware(db))
		api.Use(middlewares.CsrfMiddleware(environment))
		api.GET("role", middlewares.SessionCheckMiddleware, endpoints.GetAuthRole)
		api.POST("sign-up") // TODO

		{
			auth := api.Group("auth")
			auth.POST("login", endpoints.PostLogin)
			auth.POST("logout", endpoints.PostLogout)
		}

		{
			user := api.Group("user")
			user.Use(middlewares.SessionCheckMiddleware)
			user.Use(middlewares.UserAuthorizationMiddleware)

			user.POST("confirm-email")

			{
				userSettings := api.Group("settings")
				userSettings.POST("change-email")
				userSettings.POST("enable-2fa")
				userSettings.POST("disable-2fa")
			}
		}

		{
			admin := api.Group("admin")
			admin.Use(middlewares.SessionCheckMiddleware)
			admin.Use(middlewares.AdminAuthorizationMiddleware)

			{
				superAdmin := api.Group("super")
				superAdmin.POST("add-user") // TODO

			}
		}
	}

	return r
}
