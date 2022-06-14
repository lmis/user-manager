package ginext

import (
	"reflect"
	"user-manager/domainmodel"
	"user-manager/util"

	"database/sql"
	"fmt"

	"github.com/gin-gonic/gin"
)

type RequestContext struct {
	Authentication *domainmodel.Authentication
	Tx             *sql.Tx
	Log            util.Logger
	SecurityLog    util.Logger
}

func GetRequestContext(c *gin.Context) *RequestContext {
	val, ok := c.Get("ctx")
	if !ok {
		ctx := &RequestContext{}
		c.Set("ctx", ctx)
		return ctx
	}
	ctx, ok := val.(*RequestContext)
	if !ok {
		panic(fmt.Errorf("mistyped request context %s", reflect.TypeOf(val)))
	}
	return ctx
}

func LogAndAbortWithError(c *gin.Context, statusCode int, err error) {
	GetRequestContext(c).Log.Err(err)
	c.AbortWithError(statusCode, err)
}
