package endpoints

import (
	"net/http"
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/db/generated/models"

	"github.com/gin-gonic/gin"
)

type AuthRoleTO struct {
	Role models.UserRole `json:"role"`
}

func GetAuthRole(c *gin.Context) {
	requestContext := ginext.GetRequestContext(c)
	authRole := AuthRoleTO{}
	if requestContext.Authentication != nil {
		authRole.Role = requestContext.Authentication.AppUser.Role
	}
	c.JSON(http.StatusOK, authRole)
}
