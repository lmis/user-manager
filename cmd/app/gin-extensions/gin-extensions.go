package ginext

import (
	"reflect"
	appconfig "user-manager/cmd/app/config"
	domainmodel "user-manager/domain-model"
	"user-manager/util"

	"database/sql"

	"github.com/gin-gonic/gin"
)

const (
	RequestContextKey = "ctx"
)

type RequestContext struct {
	Config         *appconfig.Config
	Authentication *domainmodel.Authentication
	Tx             *sql.Tx
	Log            util.Logger
	SecurityLog    util.Logger
}

func GetRequestContext(c *gin.Context) *RequestContext {
	val, ok := c.Get(RequestContextKey)
	if !ok {
		panic(util.Error("missing request context"))
	}
	ctx, ok := val.(*RequestContext)
	if !ok {
		panic(util.Errorf("mistyped request context %s", reflect.TypeOf(val)))
	}
	return ctx
}
