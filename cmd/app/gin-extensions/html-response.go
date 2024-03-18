package ginext

import (
	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"net/http"
	"user-manager/cmd/app/router/middleware"
	dm "user-manager/domain-model"
	"user-manager/util/errs"
)

type MakeComponent func(r *dm.RequestContext) (templ.Component, error)

func WrapTempl(makeComponent MakeComponent) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("Content-Type", "text/html")
		component, err := makeComponent(middleware.GetRequestContext(c))

		if err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, errs.Wrap("error making component", err))
			return
		}

		if err := component.Render(c, c.Writer); err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, errs.Wrap("error rendering component", err))
			return
		}

		if !c.IsAborted() {
			c.Status(200)
		}
	}
}
