package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	dm "user-manager/domain-model"
	"user-manager/util/errs"
	"user-manager/util/slices"
)

func RegisterRequireRoleMiddleware(group *gin.RouterGroup, requiredRole dm.UserRole) {
	group.Use(func(ctx *gin.Context) {
		r := GetRequestContext(ctx)
		logger := r.Logger
		user := r.User
		if !user.IsPresent() {
			logger.Info(fmt.Sprintf("Not a %s: unauthenticated", requiredRole))
			_ = ctx.AbortWithError(http.StatusUnauthorized, errs.Error("not authenticated"))
			return
		}

		receivedRoles := user.UserRoles
		if !slices.Contains(receivedRoles, requiredRole) {
			logger.Info("Required role missing", "required", requiredRole, "received", receivedRoles)
			_ = ctx.AbortWithError(http.StatusUnauthorized, errs.Errorf("wrong role. required %s, received %v", requiredRole, receivedRoles))
			return
		}
	})
}
