package middleware

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	ginext "user-manager/cmd/app/gin-extensions"
	dm "user-manager/domain-model"
	"user-manager/util/errs"

	"github.com/teris-io/shortid"
)

func RegisterRequestContextMiddleware(app *gin.Engine, database *mongo.Database, config *dm.Config) error {
	if config == nil {
		return errs.Error("Invalid config: nil")
	}
	if database == nil {
		return errs.Error("Invalid database: nil")
	}

	sid, err := shortid.New(1, shortid.DefaultABC, 2342)
	if err != nil {
		return errs.Wrap("error creating shortid generator", err)
	}

	app.Use(func(ctx *gin.Context) {
		ginext.SetRequestContext(ctx, &dm.RequestContext{Config: config, Database: database, RequestID: sid.MustGenerate()})
	})
	return nil
}
