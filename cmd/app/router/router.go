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
	ginext "user-manager/cmd/app/gin-extensions"
	middleware "user-manager/cmd/app/middlewares"
	"user-manager/db/generated/models"
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
			Use(middleware.ExtractLoginSession).
			GET("role", ginext.WrapEndpointWithoutRequestBody(api_endpoint.GetAuthRole))

		api.Group("auth").
			Use(middleware.TimingObfuscationMiddleware(400*time.Millisecond)).
			POST("sign-up", ginext.WrapEndpointWithoutResponseBody(auth.PostSignUp)).
			POST("login", ginext.WrapEndpoint(auth.PostLogin)).
			POST("login-with-second-factor", ginext.WrapEndpoint(auth.PostLoginWithSecondFactor)).
			POST("logout", ginext.WrapEndpointWithoutResponseBody(auth.PostLogout)).
			POST("request-password-reset", ginext.WrapEndpointWithoutResponseBody(auth.PostRequestPasswordReset)).
			POST("reset-password", ginext.WrapEndpoint(auth.PostResetPassword))

		{
			user := api.Group("user")
			user.Use(middleware.RequireRoleMiddleware(models.UserRoleUSER)).
				POST("confirm-email", ginext.WrapEndpoint(user_endpoint.PostConfirmEmail)).
				POST("re-trigger-confirmation-email", ginext.WrapEndpointWithoutRequestBody(user_endpoint.PostRetriggerConfirmationEmail))

			settings := user.Group("settings")
			settings.Use(middleware.VerifiedEmailAuthorizationMiddleware).
				POST("sudo", todo).
				POST("language", todo).
				POST("confirm-email-change", ginext.WrapEndpoint(user_settings_endpoint.PostConfirmEmailChange)).
				POST("generate-temporary-2fa", todo)
			settings.Group("sensitive").
				Use(middleware.RequireSudoMode).
				POST("change-email", ginext.WrapEndpointWithoutResponseBody(sensitive_user_settings_endpoint.PostChangeEmail)).
				POST("change-password", todo).
				POST("2fa", todo)
		}
		{
			// TODO: Tests
			admin := api.Group("admin")
			admin.Use(middleware.RequireRoleMiddleware(models.UserRoleADMIN))

			admin.Group("super").
				Use(middleware.RequireRoleMiddleware(models.UserRoleADMIN)). // TODO: super-admin
				POST("add-admin-user", todo).
				POST("change-password", todo)
		}
	}

	return r
}
