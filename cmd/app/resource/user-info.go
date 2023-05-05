package resource

import (
	"github.com/gin-gonic/gin"
	ginext "user-manager/cmd/app/gin-extensions"
	dm "user-manager/domain-model"
)

func RegisterUserInfoResource(group *gin.RouterGroup) {
	group.GET("user-info", ginext.WrapEndpointWithoutRequestBody(Get))
}

type UserInfoTO struct {
	Roles         []dm.UserRole   `json:"roles"`
	EmailVerified bool            `json:"emailVerified"`
	Language      dm.UserLanguage `json:"language"`
}

func Get(_ *gin.Context, r *ginext.RequestContext) (UserInfoTO, error) {
	userSession := r.UserSession

	userTO := UserInfoTO{}
	if userSession.UserSessionID != "" {
		userTO.Roles = userSession.User.UserRoles
		userTO.EmailVerified = userSession.User.EmailVerified
		userTO.Language = userSession.User.Language
	}
	return userTO, nil
}
