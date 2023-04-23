package middleware

import (
	ginext "user-manager/cmd/app/gin-extensions"

	"github.com/gin-gonic/gin"
)

func RegisterRequestContextMiddleware(app *gin.Engine) {
	app.Use(RequestContextMiddleware)
}

func RequestContextMiddleware(c *gin.Context) {
	c.Set(ginext.RequestContextKey, &ginext.RequestContext{Ctx: c.Request.Context()})
}
