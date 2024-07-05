package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/url"
	ginext "user-manager/cmd/app/gin-extensions"
	dm "user-manager/domain-model"
	"user-manager/util/slices"
)

func RegisterLoginRedirectIfRoleMissingMiddleware(group *gin.RouterGroup, requiredRole dm.UserRole) {
	group.Use(func(ctx *gin.Context) {
		r := ginext.GetRequestContext(ctx)
		logger := r.Logger
		user := r.User
		if !user.IsPresent() {
			logger.Info(fmt.Sprintf("Not a %s: unauthenticated", requiredRole))
			login := "/auth/login?redirectUrl=" + url.QueryEscape(ctx.Request.URL.Path)
			ginext.HXLocationOrRedirect(ctx, login)
			return
		}

		receivedRoles := user.UserRoles
		if !slices.Contains(receivedRoles, requiredRole) {
			login := "/auth/login?redirectUrl=" + url.QueryEscape(ctx.Request.URL.Path)
			ginext.HXLocationOrRedirect(ctx, login)
			return
		}
	})
}
