package service

import (
	"net/http"
	domain_model "user-manager/domain-model"
	"user-manager/util/errors"
	"user-manager/util/nullable"

	"github.com/gin-gonic/gin"
)

type SessionCookieService struct {
	ctx    *gin.Context
	config *domain_model.Config
}

func ProvideSessionCookieService(ctx *gin.Context, config *domain_model.Config) *SessionCookieService {
	return &SessionCookieService{ctx, config}
}

func (s *SessionCookieService) RemoveSessionCookie(sessionType domain_model.UserSessionType) {
	s.SetSessionCookie(nullable.Empty[string](), sessionType)
}

func (s *SessionCookieService) SetSessionCookie(sessionID nullable.Nullable[string], sessionType domain_model.UserSessionType) {
	ctx := s.ctx
	config := s.config

	maxAge := -1
	value := ""
	if sessionID.IsPresent {
		value = sessionID.OrPanic()
		maxAge = int(domain_model.LOGIN_SESSION_DURATION.Seconds())
	}
	secure := true
	if config.IsLocalEnv() {
		secure = false
	}
	ctx.SetCookie(s.getCookieName(sessionType), value, maxAge, "", "", secure, true)
	ctx.SetSameSite(http.SameSiteStrictMode)
}

func (s *SessionCookieService) GetSessionCookie(sessionType domain_model.UserSessionType) (nullable.Nullable[domain_model.UserSessionID], error) {
	ctx := s.ctx

	cookie, err := ctx.Request.Cookie(s.getCookieName(sessionType))
	if err != nil {
		if err == http.ErrNoCookie {
			return nullable.Empty[domain_model.UserSessionID](), nil
		}
		return nullable.Empty[domain_model.UserSessionID](), errors.Wrap("issue reading cookie", err)
	}
	return nullable.Of(domain_model.UserSessionID(cookie.Value)), nil
}

func (s *SessionCookieService) getCookieName(sessionType domain_model.UserSessionType) string {
	return sessionType.String() + "_TOKEN"
}
