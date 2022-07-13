package authendpoints

import (
	"fmt"
	"net/http"
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/db"
	"user-manager/db/generated/models"
	"user-manager/util"

	"github.com/gin-gonic/gin"
)

func PostLogout(c *gin.Context) {
	SetSessionCookie(c, "")
	requestContext := ginext.GetRequestContext(c)
	tx := requestContext.Tx
	authentication := requestContext.Authentication
	var userSession *models.UserSession
	if authentication != nil {
		userSession = &authentication.UserSession
	}

	if userSession == nil {
		c.AbortWithError(http.StatusBadRequest, util.Error("logout without session present"))
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

	c.Status(http.StatusOK)
}