package middleware

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"
	ginext "user-manager/cmd/app/gin-extensions"
	session_service "user-manager/cmd/app/services/session"
	"user-manager/db"
	"user-manager/db/generated/models"
	"user-manager/util"

	"github.com/gin-gonic/gin"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func RequireSudoMode(c *gin.Context) {
	sudoSessionId, err := session_service.GetSessionCookie(c, models.UserSessionTypeSUDO)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("getting session cookie failed", err))
		return
	}

	if sudoSessionId == "" {
		c.AbortWithError(http.StatusForbidden, util.Error("sudo session cookie missing"))
		return
	}

	requestContext := ginext.GetRequestContext(c)
	ctx, cancelTimeout := db.DefaultQueryContext()
	defer cancelTimeout()

	session, err := models.UserSessions(models.UserSessionWhere.UserSessionID.EQ(sudoSessionId),
		models.UserSessionWhere.TimeoutAt.GT(time.Now()),
		models.UserSessionWhere.UserSessionType.EQ(models.UserSessionTypeLOGIN),
		qm.Load(models.UserSessionRels.AppUser),
		qm.Load(models.AppUserRels.AppUserRoles)).
		One(ctx, requestContext.Tx)

	// TODO BELOW

	if err != nil && err != sql.ErrNoRows {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("getting sudo session failed", err))
		return
	}

	if session == nil {
		c.AbortWithError(http.StatusForbidden, util.Error("sudo session not found on db"))
		return
	}

	session.TimeoutAt = time.Now().Add(session_service.SudoSessionDuration)

	ctx, cancelTimeout = db.DefaultQueryContext()
	defer cancelTimeout()
	rows, err := session.Update(ctx, requestContext.Tx, boil.Infer())
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("issue updating session in db", err))
		return
	}
	if rows != 1 {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap(fmt.Sprintf("wrong number of rows affected: %d", rows), err))
		return
	}
}
