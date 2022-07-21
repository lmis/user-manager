package middlewares

import (
	"database/sql"
	"net/http"
	"time"
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/db"
	"user-manager/db/generated/models"
	domainmodel "user-manager/domain-model"
	"user-manager/util"

	"github.com/gin-gonic/gin"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// TODO: Refresh the session
func SessionCheckMiddleware(c *gin.Context) {
	sessionID, err := c.Request.Cookie("SESSION_ID")
	if err != nil && err != http.ErrNoCookie {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("getting session cookie failed", err))
		return
	}

	if sessionID == nil {
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

	if err != nil && err != sql.ErrNoRows {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("getting session failed", err))
		return
	}
	if session != nil {
		requestContext.Authentication = &domainmodel.Authentication{
			UserSession: session,
			AppUser:     session.R.AppUser,
		}
	}
}
