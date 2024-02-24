package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	ginext "user-manager/cmd/app/gin-extensions"
	dm "user-manager/domain-model"
	"user-manager/util/errors"
	"user-manager/util/slices"
)

func RegisterRequireRoleMiddleware(group *gin.RouterGroup, requiredRole dm.UserRole) {
	group.Use(func(ctx *gin.Context) {
		r := ginext.GetRequestContext(ctx)
		securityLog := r.SecurityLog
		user := r.User
		if !user.IsPresent() {
			securityLog.Info("Not a %s: unauthenticated", requiredRole)
			_ = ctx.AbortWithError(http.StatusUnauthorized, errors.Error("not authenticated"))
			return
		}

		receivedRoles := user.UserRoles
		if !slices.Contains(receivedRoles, requiredRole) {
			securityLog.Info("Not a %s: wrong role (%v)", requiredRole, receivedRoles)
			_ = ctx.AbortWithError(http.StatusUnauthorized, errors.Errorf("wrong role. required %s, received %v", requiredRole, receivedRoles))
			return
		}
	})
}
