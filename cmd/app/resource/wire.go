//go:build wireinject

package resource

import (
	"user-manager/cmd/app/injector"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitializeEmailConfirmationResource(c *gin.Context) *EmailConfirmationResource {
	wire.Build(ProvideEmailConfirmationResource, injector.AllDependencies)
	return &EmailConfirmationResource{}
}

func InitializeLoginResource(c *gin.Context) *LoginResource {
	wire.Build(ProvideLoginResource, injector.AllDependencies)
	return &LoginResource{}
}

func InitializeSignUpResource(c *gin.Context) *SignUpResource {
	wire.Build(ProvideSignUpResource, injector.AllDependencies)
	return &SignUpResource{}
}

func InitializeUserInfoResource(c *gin.Context) *UserInfoResource {
	wire.Build(ProvideUserInfoResource, injector.AllDependencies)
	return &UserInfoResource{}
}

func InitializeLogoutResource(c *gin.Context) *LogoutResource {
	wire.Build(ProvideLogoutResource, injector.AllDependencies)
	return &LogoutResource{}
}
