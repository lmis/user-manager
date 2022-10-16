//go:build wireinject

package middleware

import (
	"user-manager/cmd/app/injector"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitializeCsrfMiddleware(c *gin.Context) *CsrfMiddleware {
	wire.Build(ProvideCsrfMiddleware, injector.AllDependencies)
	return &CsrfMiddleware{}
}

func InitializeDatabaseMiddleware(c *gin.Context) *DatabaseMiddleware {
	wire.Build(ProvideDatabaseMiddleware, injector.AllDependencies)
	return &DatabaseMiddleware{}
}

func InitializeExtractLoginSessionMiddleware(c *gin.Context) *ExtractLoginSessionMiddleware {
	wire.Build(ProvideExtractLoginSessionMiddleware, injector.AllDependencies)
	return &ExtractLoginSessionMiddleware{}
}

func InitializeRequireRoleMiddleware(c *gin.Context) *RequireRoleMiddleware {
	wire.Build(ProvideRequireRoleMiddleware, injector.AllDependencies)
	return &RequireRoleMiddleware{}
}

func InitializeRequireSudoModeMiddleware(c *gin.Context) *RequireSudoModeMiddleware {
	wire.Build(ProvideRequireSudoModeMiddleware, injector.AllDependencies)
	return &RequireSudoModeMiddleware{}
}

func InitializeVerifiedEmailAuthorizationMiddleware(c *gin.Context) *VerifiedEmailAuthorizationMiddleware {
	wire.Build(ProvideVerifiedEmailAuthorizationMiddleware, injector.AllDependencies)
	return &VerifiedEmailAuthorizationMiddleware{}
}
