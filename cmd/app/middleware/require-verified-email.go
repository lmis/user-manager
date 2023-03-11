package middleware

import (
	domain_model "user-manager/domain-model"
	"user-manager/util/errors"
	"user-manager/util/nullable"

	"net/http"

	"github.com/gin-gonic/gin"
)

type VerifiedEmailAuthorizationMiddleware struct {
	c           *gin.Context
	userSession nullable.Nullable[*domain_model.UserSession]
	securityLog domain_model.SecurityLog
}

func ProvideVerifiedEmailAuthorizationMiddleware(c *gin.Context, userSession nullable.Nullable[*domain_model.UserSession], securityLog domain_model.SecurityLog) *VerifiedEmailAuthorizationMiddleware {
	return &VerifiedEmailAuthorizationMiddleware{c, userSession, securityLog}
}

func RegisterVerifiedEmailAuthorizationMiddleware(group *gin.RouterGroup) {
	group.Use(func(ctx *gin.Context) { InitializeVerifiedEmailAuthorizationMiddleware(ctx).Handle() })
}

func (m *VerifiedEmailAuthorizationMiddleware) Handle() {
	c := m.c
	userSession := m.userSession
	securityLog := m.securityLog
	if userSession.IsEmpty() || !userSession.OrPanic().User.EmailVerified {
		securityLog.Info("Email not verified")
		c.AbortWithError(http.StatusForbidden, errors.Errorf("email not verified"))
		return
	}
}
