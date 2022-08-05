package middleware

import (
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/util"

	"net/http"

	"github.com/gin-gonic/gin"
)

func VerifiedEmailAuthorizationMiddleware(c *gin.Context) {
	requestContext := ginext.GetRequestContext(c)
	authentication := requestContext.Authentication
	if !authentication.AppUser.EmailVerified {
		requestContext.SecurityLog.Info("Email not verified")
		c.AbortWithError(http.StatusForbidden, util.Errorf("email not verified"))
		return
	}
}
