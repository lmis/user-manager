package middleware

import (
	ginext "user-manager/cmd/app/gin-extensions"

	"net/http"

	"github.com/gin-gonic/gin"
)

func VerifiedEmailAuthorizationMiddleware(c *gin.Context) {
	requestContext := ginext.GetRequestContext(c)
	authentication := requestContext.Authentication
	if !authentication.AppUser.EmailVerified {
		requestContext.SecurityLog.Info("Email not verified")
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
}
