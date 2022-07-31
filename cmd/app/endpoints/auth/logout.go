package auth

import (
	"fmt"
	ginext "user-manager/cmd/app/gin-extensions"
	session_service "user-manager/cmd/app/services/session"
	"user-manager/db"
	"user-manager/db/generated/models"
	"user-manager/util"

	"github.com/gin-gonic/gin"
)

func PostLogout(requestContext *ginext.RequestContext, c *gin.Context) error {
	session_service.RemoveSessionCookie(c, models.UserSessionTypeLOGIN)
	securityLog := requestContext.SecurityLog
	tx := requestContext.Tx
	authentication := requestContext.Authentication
	var userSession *models.UserSession
	if authentication != nil {
		userSession = authentication.UserSession
	}

	if userSession == nil {
		return nil
	}

	ctx, cancelTimeout := db.DefaultQueryContext()
	defer cancelTimeout()
	rows, err := userSession.Delete(ctx, tx)
	if err != nil {
		return util.Wrap("could not delete session", err)
	}
	if rows != 1 {
		return util.Error(fmt.Sprintf("too many rows affected: %d", rows))
	}

	securityLog.Info("Logout")
	return nil
}
