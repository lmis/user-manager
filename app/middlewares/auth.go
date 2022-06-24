package middlewares

import (
	"user-manager/db/generated/models"
	ginext "user-manager/gin-extensions"
	"user-manager/util"

	"net/http"

	"github.com/gin-gonic/gin"
)

func UserAuthorizationMiddleware(c *gin.Context) {
	requestContext := ginext.GetRequestContext(c)
	if err := requireRole(requestContext, models.UserRoleUSER); err != nil {
		requestContext.SecurityLog.Info("Not a user: %v", err)
		c.AbortWithStatus(getStatusCode(err))
	}
}

func AdminAuthorizationMiddleware(c *gin.Context) {
	requestContext := ginext.GetRequestContext(c)
	if err := requireRole(requestContext, models.UserRoleADMIN); err != nil {
		requestContext.SecurityLog.Info("Not an admin: %v", err)
		c.AbortWithStatus(getStatusCode(err))
	}
}

func SuperAdminAuthorizationMiddleware(c *gin.Context) {
	// TODO: Super admin user role
	c.AbortWithStatus(http.StatusForbidden)
}

func VerifiedEmailAuthorizationMiddleware(c *gin.Context) {
	requestContext := ginext.GetRequestContext(c)
	authentication := requestContext.Authentication
	if !authentication.EmailVerified {
		requestContext.SecurityLog.Info("Email not verified")
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
}

type AuthError struct {
	StatusCode int
	Err        error
}

func (e *AuthError) Error() string {
	return e.Err.Error()
}

func getStatusCode(err error) int {
	authError, ok := err.(*AuthError)
	if !ok {
		return http.StatusInternalServerError
	}
	return authError.StatusCode
}

func requireRole(requestContext *ginext.RequestContext, role models.UserRole) error {
	authentication := requestContext.Authentication
	if authentication == nil {
		return &AuthError{
			http.StatusUnauthorized,
			util.Error("requireRole", "not authenticated"),
		}
	}

	if authentication.Role != models.UserRoleUSER {
		return &AuthError{
			http.StatusForbidden,
			util.Errorf("requireRole", "wrong user role. required %s, received %s", models.UserRoleUSER, authentication.Role),
		}
	}

	return nil
}
