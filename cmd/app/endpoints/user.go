package api

import (
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/db/generated/models"

	"github.com/gin-gonic/gin"
)

type UserTO struct {
	Roles         []models.UserRole   `json:"roles"`
	EmailVerified bool                `json:"emailVerified"`
	Language      models.UserLanguage `json:"language"`
}

func GetUser(requestContext *ginext.RequestContext, _ *gin.Context) (*UserTO, error) {
	authentication := requestContext.Authentication
	userTO := UserTO{}
	if authentication != nil {
		userTO.Roles = authentication.UserRoles
		userTO.EmailVerified = authentication.AppUser.EmailVerified
		userTO.Language = authentication.AppUser.Language
	}
	return &userTO, nil
}
