package middleware

import (
	"net/http"
	"time"
	ginext "user-manager/cmd/app/gin-extensions"
	domain_model "user-manager/domain-model"
	"user-manager/repository"
	"user-manager/service"
	"user-manager/util"

	"github.com/gin-gonic/gin"
)

type RequireSudoModeMiddleware struct {
	c                    *gin.Context
	sessionCookieService *service.SessionCookieService
	sessionRepository    *repository.SessionRepository
}

func ProvideRequireSudoModeMiddleware(c *gin.Context, sessionCookieService *service.SessionCookieService, sessionRepository *repository.SessionRepository) *RequireSudoModeMiddleware {
	return &RequireSudoModeMiddleware{c, sessionCookieService, sessionRepository}
}

func RegisterRequireSudoModeMiddleware(group *gin.RouterGroup) {
	group.Use(ginext.WrapMiddleware(InitializeRequireSudoModeMiddleware))
}

func (m *RequireSudoModeMiddleware) Handle() {
	c := m.c

	sudoSessionId, err := m.sessionCookieService.GetSessionCookie(domain_model.USER_SESSION_TYPE_SUDO)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("getting session cookie failed", err))
		return
	}

	if sudoSessionId.IsEmpty() {
		c.AbortWithError(http.StatusForbidden, util.Error("sudo session cookie missing"))
		return
	}

	session, err := m.sessionRepository.GetSessionAndUser(sudoSessionId.Val, domain_model.USER_SESSION_TYPE_SUDO)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("getting sudo session failed", err))
		return
	}

	if session.IsEmpty() {
		c.AbortWithError(http.StatusForbidden, util.Error("sudo session not found on db"))
		return
	}

	if err := m.sessionRepository.UpdateSessionTimeout(session.Val.UserSessionID, time.Now().Add(domain_model.SUDO_SESSION_DURATION)); err != nil {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("issue updating session timeout in db", err))
		return
	}
}
