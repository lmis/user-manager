package middleware

import (
	dm "user-manager/domain-model"
	"user-manager/util/errors"

	"net/http"

	"github.com/gin-gonic/gin"
)

type VerifiedEmailAuthorizationMiddleware struct {
	c           *gin.Context
	userSession dm.UserSession
	securityLog dm.SecurityLog
}

func ProvideVerifiedEmailAuthorizationMiddleware(c *gin.Context, userSession dm.UserSession, securityLog dm.SecurityLog) *VerifiedEmailAuthorizationMiddleware {
	return &VerifiedEmailAuthorizationMiddleware{c, userSession, securityLog}
}

func RegisterVerifiedEmailAuthorizationMiddleware(group *gin.RouterGroup) {
	group.Use(func(ctx *gin.Context) { InitializeVerifiedEmailAuthorizationMiddleware(ctx).Handle() })
}

func (m *VerifiedEmailAuthorizationMiddleware) Handle() {
	c := m.c
	userSession := m.userSession
	securityLog := m.securityLog
	if userSession.UserSessionID == "" || !userSession.User.EmailVerified {
		securityLog.Info("Email not verified")
		_ = c.AbortWithError(http.StatusForbidden, errors.Error("email not verified"))
		return
	}
}
