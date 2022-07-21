package auth

import (
	"fmt"
	"net/http"
	ginext "user-manager/cmd/app/gin-extensions"
	session_service "user-manager/cmd/app/services/session"
	"user-manager/db"
	"user-manager/db/generated/models"
	"user-manager/util"

	"github.com/gin-gonic/gin"
)

func PostLogout(c *gin.Context) {
	session_service.RemoveSessionCookie(c)
	requestContext := ginext.GetRequestContext(c)
	securityLog := requestContext.SecurityLog
	tx := requestContext.Tx
	authentication := requestContext.Authentication
	var userSession *models.UserSession
	if authentication != nil {
		userSession = authentication.UserSession
	}

	if userSession == nil {
		c.Status(http.StatusOK)
		return
	}

	ctx, cancelTimeout := db.DefaultQueryContext()
	defer cancelTimeout()
	rows, err := userSession.Delete(ctx, tx)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("could not delete session", err))
		return
	}
	if rows != 1 {
		c.AbortWithError(http.StatusInternalServerError, util.Error(fmt.Sprintf("too many rows affected: %d", rows)))
		return
	}

	securityLog.Info("Logout")
	c.Status(http.StatusOK)
}
