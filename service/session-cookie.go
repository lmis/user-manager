package service

import (
	"net/http"
	"user-manager/db/generated/models"
	domain_model "user-manager/domain-model"
	"user-manager/util"
	"user-manager/util/nullable"

	"github.com/gin-gonic/gin"
)

type SessionCookieService struct {
	ctx *gin.Context
}

func ProvideSessionCookieService(ctx *gin.Context) *SessionCookieService {
	return &SessionCookieService{ctx}
}

func (s *SessionCookieService) RemoveSessionCookie(sessionType domain_model.UserSessionType) {
	s.SetSessionCookie(nullable.Empty[string](), sessionType)
}

func (s *SessionCookieService) SetSessionCookie(sessionID nullable.Nullable[string], sessionType domain_model.UserSessionType) {
	maxAge := int(domain_model.LOGIN_SESSION_DURATION.Seconds())
	if sessionID.IsEmpty() {
		maxAge = -1
	}
	s.ctx.SetCookie(s.getCookieName(sessionType), sessionID.Val, maxAge, "", "", true, true)
	s.ctx.SetSameSite(http.SameSiteStrictMode)
}

func (s *SessionCookieService) GetSessionCookie(sessionType domain_model.UserSessionType) (nullable.Nullable[string], error) {
	cookie, err := s.ctx.Request.Cookie(s.getCookieName(sessionType))
	if err != nil {
		if err == http.ErrNoCookie {
			return nullable.Empty[string](), nil
		}
		return nullable.Empty[string](), util.Wrap("issue reading cookie", err)
	}
	return nullable.Of(cookie.Value), nil
}

func (s *SessionCookieService) getCookieName(sessionType domain_model.UserSessionType) string {
	switch models.UserSessionType(sessionType) {
	case models.UserSessionTypeLOGIN:
		return "LOGIN_TOKEN"
	case models.UserSessionTypeREMEMBER_DEVICE:
		return "DEVICE_TOKEN"
	case models.UserSessionTypeSUDO:
		return "SUDO_TOKEN"
	}
	panic(util.Errorf("Inexhaustive switch case for value %s", sessionType))
}
