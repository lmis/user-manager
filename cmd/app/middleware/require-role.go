package middleware

import (
	domain_model "user-manager/domain-model"
	"user-manager/util/errors"
	"user-manager/util/slices"

	"net/http"

	"github.com/gin-gonic/gin"
)

type RequireRoleMiddleware struct {
	c           *gin.Context
	securityLog domain_model.SecurityLog
	userSession domain_model.UserSession
}

func ProvideRequireRoleMiddleware(c *gin.Context, securityLog domain_model.SecurityLog, userSession domain_model.UserSession) *RequireRoleMiddleware {
	return &RequireRoleMiddleware{c, securityLog, userSession}
}

func RegisterRequireRoleMiddlware(group *gin.RouterGroup, requiredRole domain_model.UserRole) {
	group.Use(func(ctx *gin.Context) { InitializeRequireRoleMiddleware(ctx).Handle(requiredRole) })
}

func (m *RequireRoleMiddleware) Handle(requiredRole domain_model.UserRole) {
	c := m.c
	securityLog := m.securityLog
	userSession := m.userSession
	if userSession.UserSessionID == "" {
		securityLog.Info("Not a %s: unauthenticated", requiredRole)
		c.AbortWithError(http.StatusUnauthorized, errors.Error("not authenticated"))
		return
	}

	receivedRoles := userSession.User.UserRoles
	if !slices.Contains(receivedRoles, requiredRole) {
		securityLog.Info("Not a %s: wrong role (%v)", requiredRole, receivedRoles)
		c.AbortWithError(http.StatusUnauthorized, errors.Errorf("wrong role. required %s, received %v", requiredRole, receivedRoles))
		return
	}
}
