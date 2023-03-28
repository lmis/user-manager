package middleware

import (
	"net/http"
	"time"
	ginext "user-manager/cmd/app/gin-extensions"
	dm "user-manager/domain-model"
	"user-manager/repository"
	"user-manager/service"
	"user-manager/util/errors"

	"github.com/gin-gonic/gin"
)

type ExtractLoginSessionMiddleware struct {
	c                    *gin.Context
	sessionRepository    *repository.SessionRepository
	sessionCookieService *service.SessionCookieService
}

func ProvideExtractLoginSessionMiddleware(c *gin.Context, sessionRepository *repository.SessionRepository, sessionCookieService *service.SessionCookieService) *ExtractLoginSessionMiddleware {
	return &ExtractLoginSessionMiddleware{c, sessionRepository, sessionCookieService}
}

func RegisterExtractLoginSessionMiddleware(group *gin.RouterGroup) {
	group.Use(func(ctx *gin.Context) { InitializeExtractLoginSessionMiddleware(ctx).Handle() })
}

func (m *ExtractLoginSessionMiddleware) Handle() {
	c := m.c
	sessionCookieService := m.sessionCookieService
	sessionRepository := m.sessionRepository

	sessionID, err := sessionCookieService.GetSessionCookie(dm.UserSessionTypeLogin)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, errors.Wrap("getting session cookie failed", err))
		return
	}

	if sessionID == "" {
		return
	}

	session, err := sessionRepository.GetSessionAndUser(sessionID, dm.UserSessionTypeLogin)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, errors.Wrap("fetching session failed", err))
		return
	}

	if session.UserSessionID != "" {

		ginext.GetRequestContext(c).UserSession = session

		if err := sessionRepository.UpdateSessionTimeout(session.UserSessionID, time.Now().Add(dm.LoginSessionDuration)); err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, errors.Wrap("issue updating session timeout in db", err))
			return
		}
	}
}
