package router

import (
	"database/sql"
	"embed"
	"github.com/gin-gonic/gin"
	"time"
	"user-manager/cmd/app/middleware"
	"user-manager/cmd/app/resource"
	dm "user-manager/domain-model"
	"user-manager/util/errors"
)

func New(translationsFS embed.FS, config *dm.Config, database *sql.DB) (*gin.Engine, error) {
	r := gin.New()
	r.HandleMethodNotAllowed = true

	err := middleware.RegisterRequestContextMiddleware(r, translationsFS, config)
	if err != nil {
		return nil, errors.Wrap("cannot setup RequestContextMiddleware", err)
	}
	middleware.RegisterLoggerMiddleware(r)
	middleware.RegisterRecoveryMiddleware(r)

	err = registerApiGroup(r.Group("api"), database)
	if err != nil {
		return nil, errors.Wrap("cannot setup ApiGroup", err)
	}

	return r, nil
}

func registerApiGroup(api *gin.RouterGroup, database *sql.DB) error {
	middleware.RegisterCsrfMiddleware(api)
	err := middleware.RegisterDatabaseMiddleware(api, database)
	if err != nil {
		return errors.Error("cannot setup DatabaseMiddleware")
	}
	middleware.RegisterExtractLoginSessionMiddleware(api)

	resource.RegisterUserInfoResource(api)

	registerAuthGroup(api.Group("auth"))
	registerAdminGroup(api.Group("admin"))
	registerUserGroup(api.Group("user"))
	return nil
}

func registerAuthGroup(auth *gin.RouterGroup) {
	middleware.RegisterTimingObfuscationMiddleware(auth, 400*time.Millisecond)

	resource.RegisterSignUpResource(auth)
	resource.RegisterLoginResource(auth)
	resource.RegisterLogoutResource(auth)
	resource.RegisterResetPasswordResource(auth)
}

func registerAdminGroup(admin *gin.RouterGroup) {
	middleware.RegisterRequireRoleMiddleware(admin, dm.UserRoleAdmin)

	registerSuperAdminGroup(admin.Group("super-admin"))
}

func registerSuperAdminGroup(superAdmin *gin.RouterGroup) {
	middleware.RegisterRequireRoleMiddleware(superAdmin, dm.UserRoleSuperAdmin)

	// POST("add-admin-user", todo).
	// POST("change-password", todo)
}

func registerUserGroup(user *gin.RouterGroup) {
	middleware.RegisterRequireRoleMiddleware(user, dm.UserRoleUser)

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
	resource.RegisterChangePasswordResource(sensitiveSettings)
	// 	POST("second-factor", todo)
}
