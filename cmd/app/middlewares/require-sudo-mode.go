package middleware

import (
	"context"
	"net/http"
	"time"
	ginext "user-manager/cmd/app/gin-extensions"
	session_service "user-manager/cmd/app/services/session"
	"user-manager/db"
	"user-manager/db/generated/models"
	"user-manager/util"

	"github.com/gin-gonic/gin"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func RequireSudoMode(c *gin.Context) {
	requestContext := ginext.GetRequestContext(c)

	sudoSessionId, err := session_service.GetSessionCookie(c, models.UserSessionTypeSUDO)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("getting session cookie failed", err))
		return
	}

	if sudoSessionId == "" {
		c.AbortWithError(http.StatusForbidden, util.Error("sudo session cookie missing"))
		return
	}

	session, err := session_service.FetchSession(requestContext, sudoSessionId, models.UserSessionTypeSUDO)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("getting sudo session failed", err))
		return
	}

	if session == nil {
		c.AbortWithError(http.StatusForbidden, util.Error("sudo session not found on db"))
		return
	}

	session.TimeoutAt = time.Now().Add(session_service.SudoSessionDuration)

	if err := db.ExecSingleMutation(func(ctx context.Context) (int64, error) { return session.Update(ctx, requestContext.Tx, boil.Infer()) }); err != nil {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("issue updating session in db", err))
		return
	}
}
