package middlewares

import (
	"fmt"
	"time"
	"user-manager/db"
	"user-manager/db/generated/models"
	"user-manager/domainmodel"
	ginext "user-manager/gin-extensions"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func SessionCheckMiddleware(c *gin.Context) {
	sessionID, err := c.Request.Cookie("SESSION_ID")
	if err != nil && err != http.ErrNoCookie {
		ginext.LogAndAbortWithError(c, http.StatusInternalServerError, err)
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
		ginext.LogAndAbortWithError(c, http.StatusInternalServerError, err)
		return
	}
	if session != nil {
		requestContext.Authentication = &domainmodel.Authentication{UserID: session.AppUserID, Role: session.R.AppUser.Role, UserSession: session}
	}
}

func UserAuthorizationMiddleware(c *gin.Context) {
	if ok := requireRole(c, models.UserRoleUSER); !ok {
		return
	}

	c.Next()
}

func AdminAuthorizationMiddleware(c *gin.Context) {
	if ok := requireRole(c, models.UserRoleADMIN); !ok {
		return
	}

	c.Next()
}

func requireRole(c *gin.Context, role models.UserRole) bool {
	requestContext := ginext.GetRequestContext(c)
	authentication := requestContext.Authentication
	if authentication == nil {
		ginext.LogAndAbortWithError(c, http.StatusUnauthorized, fmt.Errorf("not authenticated"))
		return false
	}

	if authentication.Role != models.UserRoleUSER {
		ginext.LogAndAbortWithError(c, http.StatusForbidden, fmt.Errorf("wrong user role. required %s, received %s", models.UserRoleUSER, authentication.Role))
		return false
	}

	return true
}
