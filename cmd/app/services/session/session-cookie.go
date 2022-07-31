package session

import (
	"net/http"
	"time"
	"user-manager/db/generated/models"
	"user-manager/util"

	"github.com/gin-gonic/gin"
)

const SudoSessionDuration = 10 * time.Minute
const LoginSessionDuriation = 60 * time.Minute
const DeviceSessionDuration = 30 * 24 * time.Hour

func RemoveSessionCookie(c *gin.Context, sessionType models.UserSessionType) {
	SetSessionCookie(c, "", sessionType)

}

func SetSessionCookie(c *gin.Context, sessionID string, sessionType models.UserSessionType) {
	maxAge := int(LoginSessionDuriation.Seconds())
	if sessionID == "" {
		maxAge = -1
	}
	c.SetCookie(getCookieName(sessionType), sessionID, maxAge, "", "", true, true)
	c.SetSameSite(http.SameSiteStrictMode)
}

func GetSessionCookie(c *gin.Context, sessionType models.UserSessionType) (string, error) {
	cookie, err := c.Request.Cookie(getCookieName(sessionType))
	if err != nil && err != http.ErrNoCookie {
		return "", util.Wrap("issue reading cookie", err)
	}
	return cookie.Value, nil
}

func getCookieName(sessionType models.UserSessionType) string {
	switch sessionType {
	case models.UserSessionTypeLOGIN:
		return "LOGIN_TOKEN"
	case models.UserSessionTypeREMEMBER_DEVICE:
		return "DEVICE_TOKEN"
	case models.UserSessionTypeSUDO:
		return "SUDO_TOKEN"
	}
	return ""
}
