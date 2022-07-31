package api

import (
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/db/generated/models"

	"github.com/gin-gonic/gin"
)

type AuthRoleTO struct {
	Roles         []models.UserRole `json:"roles"`
	EmailVerified bool              `json:"emailVerified"`
}

func GetAuthRole(requestContext *ginext.RequestContext, _ *gin.Context) (*AuthRoleTO, error) {
	authentication := requestContext.Authentication
	authRole := AuthRoleTO{}
	if authentication != nil {
		authRole.Roles = authentication.UserRoles
		authRole.EmailVerified = authentication.AppUser.EmailVerified
	}
	return &authRole, nil
}
