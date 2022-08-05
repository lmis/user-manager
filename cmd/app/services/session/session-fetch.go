package session

import (
	"context"
	"time"
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/db"
	"user-manager/db/generated/models"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func FetchSession(requestContext *ginext.RequestContext, sessionId string, sessionType models.UserSessionType) (*models.UserSession, error) {
	return db.Fetch(func(ctx context.Context) (*models.UserSession, error) {
		return models.UserSessions(models.UserSessionWhere.UserSessionID.EQ(sessionId),
			models.UserSessionWhere.TimeoutAt.GT(time.Now()),
			models.UserSessionWhere.UserSessionType.EQ(sessionType),
			qm.Load(models.UserSessionRels.AppUser),
			qm.Load(models.AppUserRels.AppUserRoles)).
			One(ctx, requestContext.Tx)
	})
}
