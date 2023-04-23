package injector

import (
	"context"
	"database/sql"
	ginext "user-manager/cmd/app/gin-extensions"
	dm "user-manager/domain-model"

	"github.com/gin-gonic/gin"
)

func ProvideCtx(c *gin.Context) context.Context {
	return ginext.GetRequestContext(c).Ctx
}

func ProvideTx(c *gin.Context) *sql.Tx {
	return ginext.GetRequestContext(c).Tx
}

func ProvideLog(c *gin.Context) dm.Log {
	return ginext.GetRequestContext(c).Log
}

func ProvideSecurityLog(c *gin.Context) dm.SecurityLog {
	return ginext.GetRequestContext(c).SecurityLog
}

func ProvideUserSession(c *gin.Context) dm.UserSession {
	return ginext.GetRequestContext(c).UserSession
}
