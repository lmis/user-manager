package injector

import (
	"database/sql"
	ginext "user-manager/cmd/app/gin-extensions"
	domain_model "user-manager/domain-model"
	"user-manager/util/nullable"

	"github.com/gin-gonic/gin"
)

func ProvideTx(c *gin.Context) *sql.Tx {
	return ginext.GetRequestContext(c).Tx
}

func ProvideLog(c *gin.Context) domain_model.Log {
	return ginext.GetRequestContext(c).Log
}

func ProvideSecurityLog(c *gin.Context) domain_model.SecurityLog {
	return ginext.GetRequestContext(c).SecurityLog
}

func ProvideUserSession(c *gin.Context) nullable.Nullable[*domain_model.UserSession] {
	return ginext.GetRequestContext(c).UserSession
}
