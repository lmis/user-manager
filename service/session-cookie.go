package service

import (
	"net/http"
	dm "user-manager/domain-model"
	"user-manager/util/errors"

	"github.com/gin-gonic/gin"
)

type SessionCookieService struct {
	ctx    *gin.Context
	config *dm.Config
}

func ProvideSessionCookieService(ctx *gin.Context, config *dm.Config) *SessionCookieService {
	return &SessionCookieService{ctx, config}
}

func (s *SessionCookieService) RemoveSessionCookie(sessionType dm.UserSessionType) {
	s.SetSessionCookie("", sessionType)
}

func (s *SessionCookieService) SetSessionCookie(sessionID string, sessionType dm.UserSessionType) {
	ctx := s.ctx
	config := s.config

	maxAge := -1
	value := ""
	if sessionID != "" {
		value = sessionID
		maxAge = int(dm.LoginSessionDuration.Seconds())
	}
	secure := true
	if config.IsLocalEnv() {
		secure = false
	}
	ctx.SetCookie(s.getCookieName(sessionType), value, maxAge, "", "", secure, true)
	ctx.SetSameSite(http.SameSiteStrictMode)
}

func (s *SessionCookieService) GetSessionCookie(sessionType dm.UserSessionType) (dm.UserSessionID, error) {
	ctx := s.ctx

	cookie, err := ctx.Request.Cookie(s.getCookieName(sessionType))
	if err != nil {
		if err == http.ErrNoCookie {
			return "", nil
		}
		return "", errors.Wrap("issue reading cookie", err)
	}
	return dm.UserSessionID(cookie.Value), nil
}

func (s *SessionCookieService) getCookieName(sessionType dm.UserSessionType) string {
	return sessionType.String() + "_TOKEN"
}
