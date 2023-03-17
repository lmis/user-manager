package resource

import (
	ginext "user-manager/cmd/app/gin-extensions"
	domain_model "user-manager/domain-model"

	"github.com/gin-gonic/gin"
)

type UserInfoResource struct {
	userSession domain_model.UserSession
}

func ProvideUserInfoResource(userSession domain_model.UserSession) *UserInfoResource {
	return &UserInfoResource{userSession}
}

func RegisterUserInfoResource(group *gin.RouterGroup) {
	group.GET("user-info", ginext.WrapEndpointWithoutRequestBody(InitializeUserInfoResource, (*UserInfoResource).Get))
}

type UserInfoTO struct {
	Roles         []domain_model.UserRole   `json:"roles"`
	EmailVerified bool                      `json:"emailVerified"`
	Language      domain_model.UserLanguage `json:"language"`
}

func (r *UserInfoResource) Get() (*UserInfoTO, error) {
	userSession := r.userSession

	userTO := UserInfoTO{}
	if userSession.UserSessionID != "" {
		userTO.Roles = userSession.User.UserRoles
		userTO.EmailVerified = userSession.User.EmailVerified
		userTO.Language = userSession.User.Language
	}
	return &userTO, nil
}
