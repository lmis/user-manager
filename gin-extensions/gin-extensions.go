package ginext

import (
	"embed"
	"reflect"
	"user-manager/config"
	domainmodel "user-manager/domain-model"
	"user-manager/util"

	"database/sql"

	"github.com/gin-gonic/gin"
)

type RequestContext struct {
	Config         *config.Config
	TranslationsFS embed.FS
	Authentication *domainmodel.Authentication
	Tx             *sql.Tx
	Log            util.Logger
	SecurityLog    util.Logger
}

func GetRequestContext(c *gin.Context) *RequestContext {
	val, ok := c.Get("ctx")
	if !ok {
		panic(util.Error("GetRequestContext", "missing request context"))
	}
	ctx, ok := val.(*RequestContext)
	if !ok {
		panic(util.Errorf("GetRequestContext", "mistyped request context %s", reflect.TypeOf(val)))
	}
	return ctx
}
