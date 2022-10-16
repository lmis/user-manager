package ginext

import (
	"github.com/gin-gonic/gin"
)

type Middleware interface {
	Handle()
}

type MiddlewareWithArg[T interface{}] interface {
	Handle(T)
}

func WrapMiddleware[T Middleware](makeMiddleware func(*gin.Context) T) gin.HandlerFunc {
	return func(c *gin.Context) {
		makeMiddleware(c).Handle()
	}
}

func WrapMiddlewareWithArg[A interface{}, T MiddlewareWithArg[A]](makeMiddleware func(*gin.Context) T, arg A) gin.HandlerFunc {
	return func(c *gin.Context) {
		makeMiddleware(c).Handle(arg)
	}
}
