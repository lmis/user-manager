package router

import (
	"net/http"
	"time"
	config "user-manager/cmd/app/config"
	"user-manager/cmd/app/endpoints"
	authendpoints "user-manager/cmd/app/endpoints/auth"
	userendpoints "user-manager/cmd/app/endpoints/user"
	usersettingsendpoints "user-manager/cmd/app/endpoints/user/settings"
	"user-manager/cmd/app/middlewares"
	"user-manager/util"

	"database/sql"

	"github.com/gin-gonic/gin"
)

func todo(c *gin.Context) {
	c.AbortWithError(http.StatusInternalServerError, util.Errorf("todo endpoint"))
}

func New(db *sql.DB, config *config.Config) *gin.Engine {
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
		api.Use(middlewares.CsrfMiddleware(config)).
			Use(middlewares.DatabaseMiddleware(db)).
			Use(middlewares.SessionCheckMiddleware).
			GET("role", endpoints.GetAuthRole)

		api.Group("auth").
			Use(middlewares.TimingObfuscationMiddleware(400*time.Millisecond)).
			POST("sign-up", authendpoints.PostSignUp).
			POST("login", authendpoints.PostLogin).
			POST("logout", authendpoints.PostLogout).
			POST("request-password-reset", authendpoints.PostRequestPasswordReset).
			POST("reset-password", authendpoints.PostResetPassword)

		{
			user := api.Group("user")
			user.Use(middlewares.UserAuthorizationMiddleware).
				POST("confirm-email", userendpoints.PostConfirmEmail).
				POST("re-trigger-confirmation-email", userendpoints.PostRetriggerConfirmationEmail)

			user.Group("settings").
				Use(middlewares.VerifiedEmailAuthorizationMiddleware).
				POST("confirm-email-change", usersettingsendpoints.PostConfirmEmailChange).
				POST("change-email", usersettingsendpoints.PostChangeEmail).
				POST("change-password", todo).
				POST("generate-temporary-2fa", todo).
				POST("enable-2fa", todo).
				POST("disable-2fa", todo)
		}
		{
			// TODO: Tests
			admin := api.Group("admin")
			admin.Use(middlewares.AdminAuthorizationMiddleware)

			admin.Group("super").
				Use(middlewares.SuperAdminAuthorizationMiddleware).
				POST("add-admin-user", todo).
				POST("change-password", todo)
		}
	}

	return r
}
