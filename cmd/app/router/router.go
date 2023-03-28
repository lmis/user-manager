package router

import (
	"github.com/gin-gonic/gin"
	"time"
	"user-manager/cmd/app/middleware"
	"user-manager/cmd/app/resource"
	dm "user-manager/domain-model"
)

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
	registerAdminGroup(api.Group("admin"))
	registerUserGroup(api.Group("user"))
}

func registerAuthGroup(auth *gin.RouterGroup) {
	middleware.RegisterTimingObfuscationMiddleware(auth, 400*time.Millisecond)

	resource.RegisterSignUpResource(auth)
	resource.RegisterLoginResource(auth)
	resource.RegisterLogoutResource(auth)
	resource.RegisterResetPasswordResource(auth)
}

func registerAdminGroup(admin *gin.RouterGroup) {
	middleware.RegisterRequireRoleMiddlware(admin, dm.UserRoleAdmin)

	registerSuperAdminGroup(admin.Group("super-admin"))
}

func registerSuperAdminGroup(superAdmin *gin.RouterGroup) {
	middleware.RegisterRequireRoleMiddlware(superAdmin, dm.UserRoleSuperAdmin)

	// POST("add-admin-user", todo).
	// POST("change-password", todo)
}

func registerUserGroup(user *gin.RouterGroup) {
	middleware.RegisterRequireRoleMiddlware(user, dm.UserRoleUser)

	resource.RegisterEmailConfirmationResource(user)

	registerSettingsGroup(user.Group("settings"))
}

func registerSettingsGroup(settings *gin.RouterGroup) {
	middleware.RegisterVerifiedEmailAuthorizationMiddleware(settings)

	resource.RegisterSettingsResource(settings)
	// POST("generate-temporary-second-factor-token"

	registerSensitiveSettingsGroup(settings.Group("sensitive-settings"))
}
func registerSensitiveSettingsGroup(sensitiveSettings *gin.RouterGroup) {
	middleware.RegisterRequireSudoModeMiddleware(sensitiveSettings)

	resource.RegisterSensitiveSettingsResource(sensitiveSettings)
	// 	POST("change-password", todo).
	// 	POST("second-factor", todo)
}
