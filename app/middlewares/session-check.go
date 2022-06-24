package middlewares

import (
	"net/http"
	"time"
	"user-manager/db"
	"user-manager/db/generated/models"
	"user-manager/domainmodel"
	appuser "user-manager/domainmodel/id/appUser"
	ginext "user-manager/gin-extensions"
	"user-manager/util"

	"github.com/gin-gonic/gin"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func SessionCheckMiddleware(c *gin.Context) {
	sessionID, err := c.Request.Cookie("SESSION_ID")
	if err != nil && err != http.ErrNoCookie {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("SessionCheckMiddleware", "getting session cookie failed", err))
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
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("SessionCheckMiddleware", "getting session failed", err))
		return
	}
	if session != nil {
		requestContext.Authentication = &domainmodel.Authentication{
			UserID:        appuser.ID(session.AppUserID),
			Role:          session.R.AppUser.Role,
			UserSession:   session,
			EmailVerified: session.R.AppUser.EmailVerified,
		}
	}
}
