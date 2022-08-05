package middleware

import (
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/db/generated/models"
	"user-manager/util"
	"user-manager/util/slices"

	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireRoleMiddleware(requiredRole models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestContext := ginext.GetRequestContext(c)
		securityLog := requestContext.SecurityLog
		authentication := requestContext.Authentication
		if authentication == nil {
			securityLog.Info("Not a %s: unauthenticated", requiredRole)
			c.AbortWithError(http.StatusUnauthorized, util.Error("not authenticated"))
			return
		}

		receivedRoles := authentication.UserRoles
		if !slices.Contains(receivedRoles, requiredRole) {
			securityLog.Info("Not a %s: wrong role (%v)", requiredRole, receivedRoles)
			c.AbortWithError(http.StatusUnauthorized, util.Errorf("wrong role. required %s, received %v", requiredRole, receivedRoles))
			return
		}
	}
}
