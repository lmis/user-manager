package middlewares

import (
	"time"
	"user-manager/db"
	"user-manager/db/generated/models"
	"user-manager/domainmodel"
	ginext "user-manager/gin-extensions"
	"user-manager/util"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func SessionCheckMiddleware(c *gin.Context) {
	sessionID, err := c.Request.Cookie("SESSION_ID")
	if err != nil && err != http.ErrNoCookie {
		ginext.LogAndAbortWithError(c, http.StatusInternalServerError, util.Wrap("SessionCheckMiddleware", "getting session cookie failed", err))
		return
	}

	if sessionID == nil {
		c.Next()
		return
	}

	requestContext := ginext.GetRequestContext(c)
	ctx, cancelTimeout := db.DefaultQueryContext()
	defer cancelTimeout()

	utcNow := time.Now().UTC()
	session, err := models.UserSessions(models.UserSessionWhere.UserSessionID.EQ(sessionID.Value),
		models.UserSessionWhere.TimeoutAt.GT(utcNow),
		qm.Load(models.UserSessionRels.AppUser)).
		One(ctx, requestContext.Tx)

	if err != nil {
		ginext.LogAndAbortWithError(c, http.StatusInternalServerError, util.Wrap("SessionCheckMiddleware", "getting session failed", err))
		return
	}
	if session != nil {
		requestContext.Authentication = &domainmodel.Authentication{UserID: session.AppUserID, Role: session.R.AppUser.Role, UserSession: session}
	}
}

func UserAuthorizationMiddleware(c *gin.Context) {
	requestContext := ginext.GetRequestContext(c)
	if err := requireRole(requestContext, models.UserRoleUSER); err != nil {
		requestContext.SecurityLog.Info("Not a user: %v", err)
		c.AbortWithStatus(getStatusCode(err))
		return
	}

	c.Next()
}

func AdminAuthorizationMiddleware(c *gin.Context) {
	requestContext := ginext.GetRequestContext(c)
	if err := requireRole(requestContext, models.UserRoleADMIN); err != nil {
		requestContext.SecurityLog.Info("Not an admin: %v", err)
		c.AbortWithStatus(getStatusCode(err))
		return
	}

	c.Next()
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
