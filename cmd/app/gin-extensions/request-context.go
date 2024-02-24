package ginext

import (
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
	dm "user-manager/domain-model"
	"user-manager/util/errors"

	"github.com/gin-gonic/gin"
)

const (
	RequestContextKey = "ctx"
)

type RequestContext struct {
	User        dm.User
	Database    *mongo.Database
	Log         dm.Log
	SecurityLog dm.SecurityLog
	Emailing    dm.Emailing
	Config      *dm.Config
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
