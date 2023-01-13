package middleware

import (
	"net/http"
	"time"
	ginext "user-manager/cmd/app/gin-extensions"
	domain_model "user-manager/domain-model"
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
	group.Use(ginext.WrapMiddleware(InitializeExtractLoginSessionMiddleware))
}

func (m *ExtractLoginSessionMiddleware) Handle() {
	c := m.c
	sessionCookieService := m.sessionCookieService
	sessionRepository := m.sessionRepository

	sessionId, err := sessionCookieService.GetSessionCookie(domain_model.USER_SESSION_TYPE_LOGIN)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, errors.Wrap("getting session cookie failed", err))
		return
	}

	if sessionId.IsEmpty() {
		return
	}

	session, err := sessionRepository.GetSessionAndUser(sessionId.OrPanic(), domain_model.USER_SESSION_TYPE_LOGIN)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, errors.Wrap("fetching session failed", err))
		return
	}

	if session.IsPresent {

		ginext.GetRequestContext(c).UserSession = session

		if err := sessionRepository.UpdateSessionTimeout(session.OrPanic().UserSessionID, time.Now().Add(domain_model.LOGIN_SESSION_DURATION)); err != nil {
			c.AbortWithError(http.StatusInternalServerError, errors.Wrap("issue updating session timeout in db", err))
			return
		}
	}
}
