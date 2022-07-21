package sessionservice

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RemoveSessionCookie(c *gin.Context) {
	SetSessionCookie(c, "")

}

func SetSessionCookie(c *gin.Context, sessionID string) {
	maxAge := 60 * 60
	if sessionID == "" {
		maxAge = -1
	}
	c.SetCookie("SESSION_ID", sessionID, maxAge, "", "", true, true)
	c.SetSameSite(http.SameSiteStrictMode)
}
