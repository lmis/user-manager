package resource

import (
	ginext "user-manager/cmd/app/gin-extensions"
	domain_model "user-manager/domain-model"
	"user-manager/util/nullable"

	"github.com/gin-gonic/gin"
)

type UserInfoResource struct {
	userSession nullable.Nullable[*domain_model.UserSession]
}

func ProvideUserInfoResource(userSession nullable.Nullable[*domain_model.UserSession]) *UserInfoResource {
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
	if userSession.IsPresent {
		userTO.Roles = userSession.OrPanic().User.UserRoles
		userTO.EmailVerified = userSession.OrPanic().User.EmailVerified
		userTO.Language = userSession.OrPanic().User.Language
	}
	return &userTO, nil
}
