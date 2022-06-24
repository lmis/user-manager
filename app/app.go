package app

import (
	"user-manager/app/endpoints"
	"user-manager/app/middlewares"
	"user-manager/config"

	"database/sql"

	"github.com/gin-gonic/gin"
)

func New(db *sql.DB, config *config.Config) *gin.Engine {
	if db == nil {
		panic("Invalid gin engine construction: db is nil")
	}

	r := gin.New()
	r.Use(middlewares.LoggerMiddleware)
	r.Use(middlewares.RecoveryMiddleware)

	{
		api := r.Group("api")
		api.Use(middlewares.DatabaseMiddleware(db))
		api.Use(middlewares.CsrfMiddleware(config))
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
