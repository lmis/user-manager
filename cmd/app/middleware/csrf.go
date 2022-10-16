package middleware

import (
	"net/http"
	ginext "user-manager/cmd/app/gin-extensions"
	domain_model "user-manager/domain-model"
	"user-manager/util"

	"github.com/gin-gonic/gin"
)

type CsrfMiddleware struct {
	c      *gin.Context
	config *domain_model.Config
}

func ProvideCsrfMiddleware(c *gin.Context, config *domain_model.Config) *CsrfMiddleware {
	return &CsrfMiddleware{c, config}
}

func RegisterCsrfMiddleware(group *gin.RouterGroup) {
	group.Use(ginext.WrapMiddleware(InitializeCsrfMiddleware))
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
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("getting CSRF cookie failed", err))
		return
	}
	header := c.GetHeader("X-CSRF-Token")
	if header == "" || cookie == "" {
		c.AbortWithError(http.StatusBadRequest, util.Error("missing tokens"))
		return
	}

	if header != cookie {
		c.AbortWithError(http.StatusBadRequest, util.Error("mismatching csrf tokens"))
		return
	}
}
