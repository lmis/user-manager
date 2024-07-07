package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/cmd/app/router/render"
	dm "user-manager/domain-model"
	"user-manager/util/errs"
	"user-manager/util/slices"
)

func RegisterLoginRedirectIfRoleMissingMiddleware(group *gin.RouterGroup, requiredRole dm.UserRole) {
	group.Use(func(ctx *gin.Context) {
		r := ginext.GetRequestContext(ctx)
		logger := r.Logger
		user := r.User
		if !user.IsPresent() {
			logger.Info(fmt.Sprintf("Not a %s: unauthenticated", requiredRole))
			abortAndSendLoginPage(ctx, r)
			return
		}

		receivedRoles := user.UserRoles
		if !slices.Contains(receivedRoles, requiredRole) {
			logger.Info(fmt.Sprintf("Not a %s: %s", requiredRole, receivedRoles))
			abortAndSendLoginPage(ctx, r)
			return
		}
	})
}

func abortAndSendLoginPage(ctx *gin.Context, r *dm.RequestContext) {
	ginext.HXRetarget(ctx, "closest body")
	component := render.FullPage(ctx, "Login", render.LoginForm())

	ctx.Set("Content-Type", "text/html")
	if err := component.Render(ctx, ctx.Writer); err != nil {
		_ = ctx.AbortWithError(http.StatusInternalServerError, errs.Wrap("error rendering login form page", err))
		return
	}

	ctx.AbortWithStatus(200)
}
