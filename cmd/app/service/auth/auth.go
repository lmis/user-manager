package auth

import (
	"crypto/rand"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	dm "user-manager/domain-model"
	"user-manager/util/errs"

	"golang.org/x/crypto/argon2"
)

func MakeCredentials(password []byte) (dm.UserCredentials, error) {
	if len(password) < 8 {
		return dm.UserCredentials{}, errs.Error("password too short")
	}
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return dm.UserCredentials{}, errs.Wrap("error generating salt", err)
	}

	// See https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#input-limits
	key := argon2.IDKey(password, salt, 1, 64*1024, 4, 32)

	return dm.UserCredentials{Key: key, Salt: salt}, nil
}

func VerifyCredentials(password []byte, credentials dm.UserCredentials) bool {
	key := argon2.IDKey(password, credentials.Salt, 1, 64*1024, 4, 32)
	return string(key) == string(credentials.Key)
}

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
