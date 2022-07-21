package router

import (
	"net/http"
	"time"
	config "user-manager/cmd/app/config"
	api_endpoint "user-manager/cmd/app/endpoints"
	"user-manager/cmd/app/endpoints/auth"
	user_endpoint "user-manager/cmd/app/endpoints/user"
	user_settings_endpoint "user-manager/cmd/app/endpoints/user/settings"
	sensitive_user_settings_endpoint "user-manager/cmd/app/endpoints/user/settings/sensitive"
	middleware "user-manager/cmd/app/middlewares"
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
	r.Use(middleware.RequestContextMiddleware(config))
	r.Use(middleware.LoggerMiddleware)
	r.Use(middleware.RecoveryMiddleware)

	{
		api := r.Group("api")
		api.Use(middleware.CsrfMiddleware(config)).
			Use(middleware.DatabaseMiddleware(db)).
			Use(middleware.SessionCheckMiddleware).
			GET("role", api_endpoint.GetAuthRole)

		api.Group("auth").
			Use(middleware.TimingObfuscationMiddleware(400*time.Millisecond)).
			POST("sign-up", auth.PostSignUp).
			POST("login", auth.PostLogin).
			POST("logout", auth.PostLogout).
			POST("request-password-reset", auth.PostRequestPasswordReset).
			POST("reset-password", auth.PostResetPassword)

		{
			user := api.Group("user")
			user.Use(middleware.UserAuthorizationMiddleware).
				POST("confirm-email", user_endpoint.PostConfirmEmail).
				POST("re-trigger-confirmation-email", user_endpoint.PostRetriggerConfirmationEmail)

			settings := user.Group("settings")
			settings.Use(middleware.VerifiedEmailAuthorizationMiddleware).
				POST("confirm-email-change", user_settings_endpoint.PostConfirmEmailChange).
				POST("generate-temporary-2fa", todo)
			settings.Group("sensitive").
				Use(middleware.RequireLoginCredentials).
				POST("change-email", sensitive_user_settings_endpoint.PostChangeEmail).
				POST("change-password", sensitive_user_settings_endpoint.PostChangePassword).
				POST("2fa", sensitive_user_settings_endpoint.PostSecondFactorSetting)
		}
		{
			// TODO: Tests
			admin := api.Group("admin")
			admin.Use(middleware.AdminAuthorizationMiddleware)

			admin.Group("super").
				Use(middleware.SuperAdminAuthorizationMiddleware).
				POST("add-admin-user", todo).
				POST("change-password", todo)
		}
	}

	return r
}
