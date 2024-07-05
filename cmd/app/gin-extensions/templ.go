package ginext

import (
	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"net/http"
	dm "user-manager/domain-model"
	"user-manager/util/errs"
)

type HandlerReturningComponent[requestTO interface{}] func(*gin.Context, *dm.RequestContext, requestTO) (templ.Component, error)
type HandlerReturningComponentWithoutPayload func(*gin.Context, *dm.RequestContext) (templ.Component, error)

func WrapTempl[requestTO interface{}](handler HandlerReturningComponent[requestTO]) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request requestTO
		if err := c.Bind(&request); err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, errs.Wrap("cannot bind to request TO", err))
			return
		}

		component, err := handler(c, GetRequestContext(c), request)

		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, errs.Wrap("error making component", err))
			return
		}

		if component != nil {
			c.Set("Content-Type", "text/html")
			if err := component.Render(c, c.Writer); err != nil {
				_ = c.AbortWithError(http.StatusInternalServerError, errs.Wrap("error rendering component", err))
				return
			}
		}

		if shouldWriteStatus(c) {
			status := 204
			if component != nil {
				status = 200
			}
			c.Status(status)
		}
	}
}
func WrapTemplWithoutPayload(handler HandlerReturningComponentWithoutPayload) gin.HandlerFunc {
	return func(c *gin.Context) {
		component, err := handler(c, GetRequestContext(c))
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, errs.Wrap("error making component", err))
			return
		}

		if component != nil {
			c.Set("Content-Type", "text/html")
			if err := component.Render(c, c.Writer); err != nil {
				_ = c.AbortWithError(http.StatusInternalServerError, errs.Wrap("error rendering component", err))
				return
			}
		}

		if shouldWriteStatus(c) {
			status := 204
			if component != nil {
				status = 200
			}
			c.Status(status)
		}
	}
}

func shouldWriteStatus(c *gin.Context) bool {
	return !c.IsAborted() && c.Writer.Status() == 0
}
