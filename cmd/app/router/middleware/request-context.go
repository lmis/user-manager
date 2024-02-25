package middleware

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
	dm "user-manager/domain-model"
	"user-manager/util/errs"
)

const (
	RequestContextKey = "ctx"
)

func GetRequestContext(c *gin.Context) *dm.RequestContext {
	val, ok := c.Get(RequestContextKey)
	if !ok {
		panic(errs.Error("missing request context"))
	}
	ctx, ok := val.(*dm.RequestContext)
	if !ok {
		panic(errs.Errorf("mistyped request context %s", reflect.TypeOf(val)))
	}
	return ctx
}

func RegisterRequestContextMiddleware(app *gin.Engine, database *mongo.Database, config *dm.Config) error {
	if config == nil {
		return errs.Error("Invalid config: nil")
	}
	if database == nil {
		return errs.Error("Invalid database: nil")
	}

	app.Use(func(ctx *gin.Context) {
		ctx.Set(RequestContextKey, &dm.RequestContext{Config: config, Database: database})
	})
	return nil
}
