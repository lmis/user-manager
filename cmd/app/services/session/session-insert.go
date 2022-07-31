package session

import (
	"time"
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/db"
	"user-manager/db/generated/models"
	app_user "user-manager/domain-model/id/appUser"
	"user-manager/util"

	"github.com/volatiletech/sqlboiler/v4/boil"
)

func InsertSession(requestContext *ginext.RequestContext, sessionType models.UserSessionType, appUserId app_user.ID, duration time.Duration) (string, error) {
	sessionId := util.MakeRandomURLSafeB64(21)

	session := models.UserSession{
		UserSessionID:   sessionId,
		AppUserID:       int64(appUserId),
		TimeoutAt:       time.Now().Add(duration),
		UserSessionType: sessionType,
	}
	ctx, cancelTimeout := db.DefaultQueryContext()
	defer cancelTimeout()
	if err := session.Insert(ctx, requestContext.Tx, boil.Infer()); err != nil {
		return "", util.Wrap("cannot insert session", err)
	}
	return sessionId, nil
}
