package api

import (
	"net/http"
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/db/generated/models"

	"github.com/gin-gonic/gin"
)

type AuthRoleTO struct {
	Role          models.UserRole `json:"role"`
	EmailVerified bool            `json:"emailVerified"`
}

func GetAuthRole(c *gin.Context) {
	requestContext := ginext.GetRequestContext(c)
	authentication := requestContext.Authentication
	authRole := AuthRoleTO{}
	if authentication != nil {
		authRole.Role = authentication.AppUser.Role
		authRole.EmailVerified = authentication.AppUser.EmailVerified
	}
	c.JSON(http.StatusOK, authRole)
}
