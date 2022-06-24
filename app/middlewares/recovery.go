package middlewares

import (
	"fmt"
	"net/http"
	"user-manager/util"

	"github.com/gin-gonic/gin"
)

var RecoveryMiddleware = gin.CustomRecovery(recoveryHandler)

func recoveryHandler(c *gin.Context, requestErr interface{}) {
	err, ok := requestErr.(error)
	if !ok {
		err = fmt.Errorf("%v", requestErr)
	}
	c.AbortWithError(http.StatusInternalServerError, util.Wrap("recoveryHandler", "recovered from panic", err))
}
