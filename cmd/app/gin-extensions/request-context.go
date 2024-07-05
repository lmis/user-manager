package ginext

import (
	"github.com/gin-gonic/gin"
	"reflect"
	dm "user-manager/domain-model"
	"user-manager/util/errs"
)

const (
	requestContextKey = "REQ-CTX"
)

func GetRequestContext(c *gin.Context) *dm.RequestContext {
	val := c.MustGet(requestContextKey)
	ctx, ok := val.(*dm.RequestContext)
	if !ok {
		panic(errs.Errorf("mistyped request context %s", reflect.TypeOf(val)))
	}
	return ctx
}

func SetRequestContext(c *gin.Context, r *dm.RequestContext) {
	c.Set(requestContextKey, r)
}
