package resource

import (
	"github.com/gin-gonic/gin"
	ginext "user-manager/cmd/app/gin-extensions"
	dm "user-manager/domain-model"
)

type UserInfoResource struct {
	userSession dm.UserSession
}

func ProvideUserInfoResource(userSession dm.UserSession) *UserInfoResource {
	return &UserInfoResource{userSession}
}

func RegisterUserInfoResource(group *gin.RouterGroup) {
	group.GET("user-info", ginext.WrapEndpointWithoutRequestBody(InitializeUserInfoResource, (*UserInfoResource).Get))
}

type UserInfoTO struct {
	Roles         []dm.UserRole   `json:"roles"`
	EmailVerified bool            `json:"emailVerified"`
	Language      dm.UserLanguage `json:"language"`
}

func (r *UserInfoResource) Get() (UserInfoTO, error) {
	userSession := r.userSession

	userTO := UserInfoTO{}
	if userSession.UserSessionID != "" {
		userTO.Roles = userSession.User.UserRoles
		userTO.EmailVerified = userSession.User.EmailVerified
		userTO.Language = userSession.User.Language
	}
	return userTO, nil
}
