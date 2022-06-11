package endpoints

import (
	"net/http"
	"user-manager/db/generated/models"
	ginext "user-manager/gin-extensions"

	"github.com/gin-gonic/gin"
)

type AuthRoleTO struct {
	Role models.UserRole `json:"role"`
}

func GetAuthRole(c *gin.Context) {
	requestContext := ginext.GetRequestContext(c)
	authRole := AuthRoleTO{}
	if requestContext.Authentication != nil {
		authRole.Role = requestContext.Authentication.Role
	}
	c.JSON(http.StatusOK, authRole)
}
