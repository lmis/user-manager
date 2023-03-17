package ginext

import (
	"database/sql"
	"reflect"
	domain_model "user-manager/domain-model"
	"user-manager/util/errors"
	"user-manager/util/nullable"

	"github.com/gin-gonic/gin"
)

const (
	REQUEST_CONTEXT_KEY = "ctx"
)

type RequestContext struct {
	UserSession nullable.Nullable[domain_model.UserSession]
	Tx          *sql.Tx
	Log         domain_model.Log
	SecurityLog domain_model.SecurityLog
}

func GetRequestContext(c *gin.Context) *RequestContext {
	val, ok := c.Get(REQUEST_CONTEXT_KEY)
	if !ok {
		panic(errors.Error("missing request context"))
	}
	ctx, ok := val.(*RequestContext)
	if !ok {
		panic(errors.Errorf("mistyped request context %s", reflect.TypeOf(val)))
	}
	return ctx
}
