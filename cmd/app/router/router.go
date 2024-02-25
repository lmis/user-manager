package router

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
	"user-manager/cmd/app/resource"
	middleware2 "user-manager/cmd/app/router/middleware"
	dm "user-manager/domain-model"
	"user-manager/util/errs"
)

func New(config *dm.Config, database *mongo.Database) (*gin.Engine, error) {
	r := gin.New()
	r.HandleMethodNotAllowed = true

	err := middleware2.RegisterRequestContextMiddleware(r, database, config)
	if err != nil {
		return nil, errs.Wrap("cannot setup RequestContextMiddleware", err)
	}
	middleware2.RegisterLoggerMiddleware(r)
	middleware2.RegisterRecoveryMiddleware(r)

	err = registerApiGroup(r.Group("api"))
	if err != nil {
		return nil, errs.Wrap("cannot setup ApiGroup", err)
	}

	return r, nil
}

func registerApiGroup(api *gin.RouterGroup) error {
	middleware2.RegisterCsrfMiddleware(api)
	middleware2.RegisterExtractLoginSessionMiddleware(api)

	resource.RegisterUserInfoResource(api)

	registerAuthGroup(api.Group("auth"))
	registerAdminGroup(api.Group("admin"))
	registerUserGroup(api.Group("user"))
	return nil
}

func registerAuthGroup(auth *gin.RouterGroup) {
	middleware2.RegisterTimingObfuscationMiddleware(auth, 400*time.Millisecond)

	resource.RegisterSignUpResource(auth)
	resource.RegisterLoginResource(auth)
	resource.RegisterLogoutResource(auth)
	resource.RegisterResetPasswordResource(auth)
}

func registerAdminGroup(admin *gin.RouterGroup) {
	middleware2.RegisterRequireRoleMiddleware(admin, dm.UserRoleAdmin)

	registerSuperAdminGroup(admin.Group("super-admin"))
}

func registerSuperAdminGroup(superAdmin *gin.RouterGroup) {
	middleware2.RegisterRequireRoleMiddleware(superAdmin, dm.UserRoleSuperAdmin)

	// POST("add-admin-user", todo).
	// POST("change-password", todo)
}

func registerUserGroup(user *gin.RouterGroup) {
	middleware2.RegisterRequireRoleMiddleware(user, dm.UserRoleUser)

	resource.RegisterEmailConfirmationResource(user)

	registerSettingsGroup(user.Group("settings"))
}

func registerSettingsGroup(settings *gin.RouterGroup) {
	middleware2.RegisterVerifiedEmailAuthorizationMiddleware(settings)

	resource.RegisterSettingsResource(settings)
	// POST("generate-temporary-second-factor-token"

	registerSensitiveSettingsGroup(settings.Group("sensitive-settings"))
}
func registerSensitiveSettingsGroup(sensitiveSettings *gin.RouterGroup) {
	middleware2.RegisterRequireSudoModeMiddleware(sensitiveSettings)

	resource.RegisterSensitiveSettingsResource(sensitiveSettings)
	resource.RegisterChangePasswordResource(sensitiveSettings)
	// 	POST("second-factor", todo)
}
