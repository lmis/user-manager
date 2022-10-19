package middleware

import (
	ginext "user-manager/cmd/app/gin-extensions"
	domain_model "user-manager/domain-model"
	"user-manager/util"
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
	group.Use(ginext.WrapMiddleware(InitializeVerifiedEmailAuthorizationMiddleware))
}

func (m *VerifiedEmailAuthorizationMiddleware) Handle() {
	c := m.c
	userSession := m.userSession
	securityLog := m.securityLog
	if userSession.IsEmpty() || !userSession.OrPanic().User.EmailVerified {
		securityLog.Info("Email not verified")
		c.AbortWithError(http.StatusForbidden, util.Errorf("email not verified"))
		return
	}
}
