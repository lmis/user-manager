package ginext

import (
	"context"
	"database/sql"
	"reflect"
	dm "user-manager/domain-model"
	"user-manager/util/errors"

	"github.com/gin-gonic/gin"
)

const (
	RequestContextKey = "ctx"
)

type RequestContext struct {
	Ctx         context.Context
	UserSession dm.UserSession
	Tx          *sql.Tx
	Log         dm.Log
	SecurityLog dm.SecurityLog
}

func GetRequestContext(c *gin.Context) *RequestContext {
	val, ok := c.Get(RequestContextKey)
	if !ok {
		panic(errors.Error("missing request context"))
	}
	ctx, ok := val.(*RequestContext)
	if !ok {
		panic(errors.Errorf("mistyped request context %s", reflect.TypeOf(val)))
	}
	return ctx
}
