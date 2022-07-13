package router

import (
	"time"
	config "user-manager/cmd/app/config"
	"user-manager/cmd/app/endpoints"
	authendpoints "user-manager/cmd/app/endpoints/auth"
	userendpoints "user-manager/cmd/app/endpoints/user"
	"user-manager/cmd/app/middlewares"

	"database/sql"

	"github.com/gin-gonic/gin"
)

func New(db *sql.DB, config *config.Config) *gin.Engine {
	todo := func(c *gin.Context) {}
	if db == nil {
		panic("Invalid gin engine construction: db is nil")
	}
	if config == nil {
		panic("Invalid gin engine construction: config is nil")
	}

	r := gin.New()
	r.HandleMethodNotAllowed = true
	r.Use(middlewares.RequestContextMiddleware(config))
	r.Use(middlewares.LoggerMiddleware)
	r.Use(middlewares.RecoveryMiddleware)

	{
		api := r.Group("api")
		api.Use(middlewares.CsrfMiddleware(config))
		api.Use(middlewares.DatabaseMiddleware(db))
		api.GET("role",
			middlewares.SessionCheckMiddleware,
			endpoints.GetAuthRole,
		)
		api.POST("sign-up",
			middlewares.TimingObfuscationMiddleware(400*time.Millisecond),
			endpoints.PostSignUp,
		)

		{
			auth := api.Group("auth")
			auth.POST("login",
				middlewares.TimingObfuscationMiddleware(400*time.Millisecond),
				authendpoints.PostLogin,
			)
			auth.POST("logout", authendpoints.PostLogout)
		}

		{
			user := api.Group("user")
			user.Use(middlewares.SessionCheckMiddleware)
			user.Use(middlewares.UserAuthorizationMiddleware)

			user.POST("confirm-email", userendpoints.PostConfirmEmail)
			user.POST("re-trigger-confirmation-email", todo)

			{
				userSettings := api.Group("settings")
				userSettings.Use(middlewares.VerifiedEmailAuthorizationMiddleware)
				userSettings.POST("change-email", todo)
				userSettings.POST("change-password", todo)
				userSettings.POST("generate-temporary-2fa", todo)
				userSettings.POST("enable-2fa", todo)
				userSettings.POST("disable-2fa", todo)
			}
		}
		{
			admin := api.Group("admin")
			admin.Use(middlewares.SessionCheckMiddleware)
			admin.Use(middlewares.AdminAuthorizationMiddleware)

			{
				superAdmin := api.Group("super")
				admin.Use(middlewares.SuperAdminAuthorizationMiddleware)
				superAdmin.POST("add-admin-user", todo)
				superAdmin.POST("change-password", todo)
			}
		}
	}

	return r
}
