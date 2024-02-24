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
	user := r.User

	if !user.IsPresent() {
		return UserInfoTO{}, nil
	}

	return UserInfoTO{
		Roles:         user.UserRoles,
		EmailVerified: user.EmailVerified,
		Language:      user.Language,
	}, nil
}
