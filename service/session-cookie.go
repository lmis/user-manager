package service

import (
	"errors"
	"net/http"
	dm "user-manager/domain-model"
	"user-manager/util/errs"

	"github.com/gin-gonic/gin"
)

func RemoveSessionCookie(ctx *gin.Context, config *dm.Config, sessionType dm.UserSessionType) {
	SetSessionCookie(ctx, config, "", sessionType)
}

func SetSessionCookie(ctx *gin.Context, config *dm.Config, sessionID string, sessionType dm.UserSessionType) {

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
	ctx.SetCookie(getCookieName(sessionType), value, maxAge, "", "", secure, true)
	ctx.SetSameSite(http.SameSiteStrictMode)
}

func GetSessionCookie(ctx *gin.Context, sessionType dm.UserSessionType) (dm.UserSessionToken, error) {
	cookie, err := ctx.Request.Cookie(getCookieName(sessionType))
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return "", nil
		}
		return "", errs.Wrap("issue reading cookie", err)
	}
	return dm.UserSessionToken(cookie.Value), nil
}

func getCookieName(sessionType dm.UserSessionType) string {
	return string(sessionType) + "_TOKEN"
}
