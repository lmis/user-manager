package ginext

import (
	"embed"
	"reflect"
	appconfig "user-manager/cmd/app/config"
	domainmodel "user-manager/domain-model"
	"user-manager/util"

	"database/sql"

	"github.com/gin-gonic/gin"
)

type RequestContext struct {
	Config         *appconfig.Config
	TranslationsFS embed.FS
	Authentication *domainmodel.Authentication
	Tx             *sql.Tx
	Log            util.Logger
	SecurityLog    util.Logger
}

func GetRequestContext(c *gin.Context) *RequestContext {
	val, ok := c.Get("ctx")
	if !ok {
		panic(util.Error("missing request context"))
	}
	ctx, ok := val.(*RequestContext)
	if !ok {
		panic(util.Errorf("mistyped request context %s", reflect.TypeOf(val)))
	}
	return ctx
}
