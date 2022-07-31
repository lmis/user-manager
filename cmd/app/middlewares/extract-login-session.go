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
	domain_model "user-manager/domain-model"
	"user-manager/util"

	"github.com/gin-gonic/gin"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func ExtractLoginSession(c *gin.Context) {
	sessionID, err := session_service.GetSessionCookie(c, models.UserSessionTypeLOGIN)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("getting session cookie failed", err))
		return
	}

	if sessionID == "" {
		return
	}

	requestContext := ginext.GetRequestContext(c)
	ctx, cancelTimeout := db.DefaultQueryContext()
	defer cancelTimeout()

	session, err := models.UserSessions(models.UserSessionWhere.UserSessionID.EQ(sessionID),
		models.UserSessionWhere.TimeoutAt.GT(time.Now()),
		models.UserSessionWhere.UserSessionType.EQ(models.UserSessionTypeLOGIN),
		qm.Load(models.UserSessionRels.AppUser),
		qm.Load(models.AppUserRels.AppUserRoles)).
		One(ctx, requestContext.Tx)

	if err != nil && err != sql.ErrNoRows {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("getting session failed", err))
		return
	}
	if session != nil {
		appUserRoles := session.R.AppUser.R.AppUserRoles
		userRoles := make([]models.UserRole, len(appUserRoles))
		for i, role := range appUserRoles {
			userRoles[i] = role.Role
		}

		requestContext.Authentication = &domain_model.Authentication{
			UserSession: session,
			UserRoles:   userRoles,
			AppUser:     session.R.AppUser,
		}

		session.TimeoutAt = time.Now().Add(session_service.LoginSessionDuriation)

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
}
