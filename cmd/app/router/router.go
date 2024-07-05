package router

import (
	"embed"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"io/fs"
	"net/http"
	"time"
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/cmd/app/resource"
	"user-manager/cmd/app/router/middleware"
	"user-manager/cmd/app/router/render"
	dm "user-manager/domain-model"
	"user-manager/util/errs"
)

//go:embed assets/*
var assetsFS embed.FS

func New(config *dm.Config, database *mongo.Database) (*gin.Engine, error) {
	r := gin.New()

	err := middleware.RegisterRequestContextMiddleware(r, database, config)
	if err != nil {
		return nil, errs.Wrap("cannot setup RequestContextMiddleware", err)
	}
	middleware.RegisterLoggerMiddleware(r)
	middleware.RegisterRecoveryMiddleware(r)

	err = registerAssets(r)
	if err != nil {
		return nil, errs.Wrap("cannot setup assets", err)
	}

	r.NoRoute(func(c *gin.Context) {
		ginext.HXLocationOrRedirect(c, "/user/home")
	})
	err = registerGroups(r.Group(""))
	if err != nil {
		return nil, errs.Wrap("cannot setup ApiGroup", err)
	}

	return r, nil
}

func registerAssets(r *gin.Engine) error {
	subFS, err := fs.Sub(assetsFS, "assets")
	if err != nil {
		return errs.Wrap("cannot create FS rooted at assets for assetsFS", err)
	}
	r.StaticFS("/assets", http.FS(subFS))

	return nil
}

func registerGroups(root *gin.RouterGroup) error {
	middleware.RegisterCsrfMiddleware(root)
	middleware.RegisterExtractLoginSessionMiddleware(root)

	resource.RegisterUserInfoResource(root)

	registerAuthGroup(root.Group("auth"))
	registerAdminGroup(root.Group("admin"))
	registerUserGroup(root.Group("user"))
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
	middleware.RegisterLoginRedirectIfRoleMissingMiddleware(admin, dm.UserRoleAdmin)

	registerSuperAdminGroup(admin.Group("super-admin"))

	// TODO: Add redirect middleware for unmatched paths
}

func registerSuperAdminGroup(superAdmin *gin.RouterGroup) {
	middleware.RegisterLoginRedirectIfRoleMissingMiddleware(superAdmin, dm.UserRoleSuperAdmin)

	// POST("add-admin-user", todo).
	// POST("change-password", todo)
}

func registerUserGroup(user *gin.RouterGroup) {
	middleware.RegisterLoginRedirectIfRoleMissingMiddleware(user, dm.UserRoleUser)

	resource.RegisterEmailConfirmationResource(user)
	user.GET("home", ginext.WrapTemplWithoutPayload(render.UserHome))

	registerSettingsGroup(user.Group("settings"))

	// TODO: Add redirect middleware for unmatched paths
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
