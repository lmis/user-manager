package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"time"
	ginext "user-manager/cmd/app/gin-extensions"
	session_service "user-manager/cmd/app/services/session"
	"user-manager/db"
	"user-manager/db/generated/models"
	domain_model "user-manager/domain-model"
	"user-manager/util"
	"user-manager/util/slices"

	"github.com/gin-gonic/gin"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func ExtractLoginSession(c *gin.Context) {
	sessionId, err := session_service.GetSessionCookie(c, models.UserSessionTypeLOGIN)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("getting session cookie failed", err))
		return
	}

	if sessionId == "" {
		return
	}

	requestContext := ginext.GetRequestContext(c)

	session, err := session_service.FetchSessionAndUser(requestContext, sessionId, models.UserSessionTypeLOGIN)

	if err != nil && err != sql.ErrNoRows {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("fetching session failed", err))
		return
	}
	if session != nil {
		appUserRoles, err := db.Fetch(func(ctx context.Context) (models.AppUserRoleSlice, error) {
			return models.AppUserRoles(models.AppUserRoleWhere.AppUserID.EQ(session.AppUserID)).
				All(ctx, requestContext.Tx)
		})

		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, util.Wrap("fetching app user roles failed", err))
			return
		}
		userRoles := slices.Map(appUserRoles, func(role *models.AppUserRole) models.UserRole { return role.Role })

		requestContext.Authentication = &domain_model.Authentication{
			UserSession: session,
			UserRoles:   userRoles,
			AppUser:     session.R.AppUser,
		}

		session.TimeoutAt = time.Now().Add(session_service.LoginSessionDuriation)

		if err := db.ExecSingleMutation(func(ctx context.Context) (int64, error) { return session.Update(ctx, requestContext.Tx, boil.Infer()) }); err != nil {
			c.AbortWithError(http.StatusInternalServerError, util.Wrap("issue updating session in db", err))
			return
		}
	}
}
