package middleware

import (
	"net/http"
	dm "user-manager/domain-model"
	"user-manager/util/errors"

	"github.com/gin-gonic/gin"
)

type CsrfMiddleware struct {
	c      *gin.Context
	config *dm.Config
}

func ProvideCsrfMiddleware(c *gin.Context, config *dm.Config) *CsrfMiddleware {
	return &CsrfMiddleware{c, config}
}

func RegisterCsrfMiddleware(group *gin.RouterGroup) {
	group.Use(func(ctx *gin.Context) { InitializeCsrfMiddleware(ctx).Handle() })
}

func (m *CsrfMiddleware) Handle() {
	c := m.c
	config := m.config

	cookieName := "__Host-CSRF-Token"
	if config.IsLocalEnv() {
		cookieName = "CSRF-Token"
	}
	cookie, err := c.Cookie(cookieName)
	if err != nil && err != http.ErrNoCookie {
		_ = c.AbortWithError(http.StatusInternalServerError, errors.Wrap("getting CSRF cookie failed", err))
		return
	}
	header := c.GetHeader("X-CSRF-Token")
	if header == "" || cookie == "" {
		_ = c.AbortWithError(http.StatusBadRequest, errors.Error("missing tokens"))
		return
	}

	if header != cookie {
		_ = c.AbortWithError(http.StatusBadRequest, errors.Error("mismatching csrf tokens"))
		return
	}
}
