package service

import (
	"net/http"
	domain_model "user-manager/domain-model"
	"user-manager/util/errors"

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
	s.SetSessionCookie("", sessionType)
}

func (s *SessionCookieService) SetSessionCookie(sessionID string, sessionType domain_model.UserSessionType) {
	ctx := s.ctx
	config := s.config

	maxAge := -1
	value := ""
	if sessionID != "" {
		value = sessionID
		maxAge = int(domain_model.LOGIN_SESSION_DURATION.Seconds())
	}
	secure := true
	if config.IsLocalEnv() {
		secure = false
	}
	ctx.SetCookie(s.getCookieName(sessionType), value, maxAge, "", "", secure, true)
	ctx.SetSameSite(http.SameSiteStrictMode)
}

func (s *SessionCookieService) GetSessionCookie(sessionType domain_model.UserSessionType) (domain_model.UserSessionID, error) {
	ctx := s.ctx

	cookie, err := ctx.Request.Cookie(s.getCookieName(sessionType))
	if err != nil {
		if err == http.ErrNoCookie {
			return "", nil
		}
		return "", errors.Wrap("issue reading cookie", err)
	}
	return domain_model.UserSessionID(cookie.Value), nil
}

func (s *SessionCookieService) getCookieName(sessionType domain_model.UserSessionType) string {
	return sessionType.String() + "_TOKEN"
}
