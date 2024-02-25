package router

import (
	"embed"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"io/fs"
	"net/http"
	"time"
	"user-manager/cmd/app/resource"
	"user-manager/cmd/app/router/components"
	"user-manager/cmd/app/router/middleware"
	dm "user-manager/domain-model"
	"user-manager/util/errs"
)

//go:embed assets/*
var assetsFS embed.FS

func New(config *dm.Config, database *mongo.Database) (*gin.Engine, error) {
	r := gin.New()
	r.HandleMethodNotAllowed = true

	err := middleware.RegisterRequestContextMiddleware(r, database, config)
	if err != nil {
		return nil, errs.Wrap("cannot setup RequestContextMiddleware", err)
	}
	middleware.RegisterLoggerMiddleware(r)
	middleware.RegisterRecoveryMiddleware(r)

	subFS, err := fs.Sub(assetsFS, "assets")
	if err != nil {
		return nil, errs.Wrap("cannot create FS rooted at assets for assetsFS", err)
	}
	r.StaticFS("/assets", http.FS(subFS))

	r.GET("/index.html", func(c *gin.Context) {
		c.Set("Content-Type", "text/html")
		if err := components.Index(middleware.GetRequestContext(c), c.Writer); err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, errs.Wrap("cannot render index.html", err))
		}
		if !c.IsAborted() {
			c.Status(200)
		}
	})

	err = registerApiGroup(r.Group("api"))
	if err != nil {
		return nil, errs.Wrap("cannot setup ApiGroup", err)
	}

	return r, nil
}

func registerApiGroup(api *gin.RouterGroup) error {
	middleware.RegisterCsrfMiddleware(api)
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
