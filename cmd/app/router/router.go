package router

import (
	"net/http"
	"time"
	"user-manager/cmd/app/middleware"
	"user-manager/cmd/app/resource"
	domain_model "user-manager/domain-model"
	"user-manager/util"

	"github.com/gin-gonic/gin"
)

func todo(c *gin.Context) {
	c.AbortWithError(http.StatusInternalServerError, util.Errorf("todo endpoint"))
}

func New() *gin.Engine {
	r := gin.New()
	r.HandleMethodNotAllowed = true
	middleware.RegisterRequestContextMiddleware(r)
	middleware.RegisterLoggerMiddleware(r)
	middleware.RegisterRecoveryMiddleware(r)

	registerApiGroup(r.Group("api"))

	return r
}

func registerApiGroup(api *gin.RouterGroup) {
	middleware.RegisterCsrfMiddleware(api)
	middleware.RegisterDatabaseMiddleware(api)
	middleware.RegisterExtractLoginSessionMiddleware(api)

	resource.RegisterUserInfoResource(api)

	registerAuthGroup(api.Group("auth"))
	registerUserGroup(api.Group("user"))
	registerAdminGroup(api.Group("admin"))
}

func registerAdminGroup(admin *gin.RouterGroup) {
	middleware.RegisterRequireRoleMiddlware(admin, domain_model.USER_ROLE_ADMIN)

	registerSuperAdminGroup(admin.Group("super-admin"))
}

func registerSuperAdminGroup(superAdmin *gin.RouterGroup) {
	middleware.RegisterRequireRoleMiddlware(superAdmin, domain_model.USER_ROLE_SUPER_ADMIN)
	superAdmin.POST("add-admin-user", todo).
		POST("change-password", todo)
}

func registerAuthGroup(auth *gin.RouterGroup) {
	middleware.RegisterTimingObfuscationMiddleware(auth, 400*time.Millisecond)
	resource.RegisterLoginResource(auth)
	resource.RegisterLogoutResource(auth)
	// 	POST("request-password-reset", ginext.WrapEndpointWithoutResponseBody(auth.PostRequestPasswordReset)).
	// 	POST("reset-password", ginext.WrapEndpoint(auth.PostResetPassword))
}

func registerUserGroup(user *gin.RouterGroup) {
	middleware.RegisterRequireRoleMiddlware(user, domain_model.USER_ROLE_USER)
	resource.RegisterSignUpResource(user)
	resource.RegisterEmailConfirmationResource(user)

	registerSettingsGroup(user.Group("settings"))
}

func registerSettingsGroup(settings *gin.RouterGroup) {
	middleware.RegisterVerifiedEmailAuthorizationMiddleware(settings)
	// settings.POST("sudo", ginext.WrapEndpoint(user_settings_endpoint.PostSudo)).
	// 	POST("language", ginext.WrapEndpointWithoutResponseBody(user_settings_endpoint.PostLanguage)).
	// 	POST("confirm-email-change", ginext.WrapEndpoint(user_settings_endpoint.PostConfirmEmailChange)).
	// 	POST("generate-temporary-2fa", todo)
	registerSensitiveSettingsGroup(settings.Group("sensitive-settings"))
}
func registerSensitiveSettingsGroup(sensitiveSettings *gin.RouterGroup) {
	middleware.RegisterRequireSudoModeMiddleware(sensitiveSettings)
	// sensitiveSettings.POST("change-email", ginext.WrapEndpointWithoutResponseBody(sensitive_user_settings_endpoint.PostChangeEmail)).
	// 	POST("change-password", todo).
	// 	POST("2fa", todo)
}
