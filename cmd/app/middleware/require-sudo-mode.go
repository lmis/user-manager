package middleware

import (
	"net/http"
	"time"
	domain_model "user-manager/domain-model"
	"user-manager/repository"
	"user-manager/service"
	"user-manager/util/errors"

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
	group.Use(func(ctx *gin.Context) { InitializeRequireSudoModeMiddleware(ctx).Handle() })
}

func (m *RequireSudoModeMiddleware) Handle() {
	c := m.c
	sessionCookieService := m.sessionCookieService
	sessionRepository := m.sessionRepository

	sudoSessionId, err := sessionCookieService.GetSessionCookie(domain_model.USER_SESSION_TYPE_SUDO)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, errors.Wrap("getting session cookie failed", err))
		return
	}

	if sudoSessionId.IsEmpty() {
		c.AbortWithError(http.StatusForbidden, errors.Error("sudo session cookie missing"))
		return
	}

	session, err := sessionRepository.GetSessionAndUser(sudoSessionId.OrPanic(), domain_model.USER_SESSION_TYPE_SUDO)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, errors.Wrap("getting sudo session failed", err))
		return
	}

	if session.IsEmpty() {
		c.AbortWithError(http.StatusForbidden, errors.Error("sudo session not found on db"))
		return
	}

	if err := sessionRepository.UpdateSessionTimeout(session.OrPanic().UserSessionID, time.Now().Add(domain_model.SUDO_SESSION_DURATION)); err != nil {
		c.AbortWithError(http.StatusInternalServerError, errors.Wrap("issue updating session timeout in db", err))
		return
	}
}
