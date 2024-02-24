package middleware

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	ginext "user-manager/cmd/app/gin-extensions"
)

func RegisterDatabaseMiddleware(group *gin.RouterGroup, database *mongo.Database) error {
	group.Use(func(ctx *gin.Context) {
		r := ginext.GetRequestContext(ctx)

		r.Database = database

		ctx.Next()
	})
	return nil
}
