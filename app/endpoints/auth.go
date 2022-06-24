package endpoints

import (
	"fmt"
	"net/http"
	"time"
	"user-manager/db"
	"user-manager/db/generated/models"
	ginext "user-manager/gin-extensions"
	"user-manager/util"

	"github.com/gin-gonic/gin"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"golang.org/x/crypto/bcrypt"
)

type CredentialsTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponseTO struct {
	LoggedIn bool `json:"loggedIn"`
}

// TODO: Hanlde 2FA
func PostLogin(c *gin.Context) {
	requestContext := ginext.GetRequestContext(c)
	tx := requestContext.Tx
	securityLog := requestContext.SecurityLog
	loginResponseTO := LoginResponseTO{}
	var credentialsTO CredentialsTO
	err := c.BindJSON(&credentialsTO)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, util.Wrap("PostLogin", "cannot bind to credentialsTO", err))
		return
	}

	var user *models.AppUser
	ctx, cancelTimeout := db.DefaultQueryContext()
	defer cancelTimeout()

	user, err = models.AppUsers(models.AppUserWhere.Email.EQ(credentialsTO.Email)).One(ctx, tx)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("PostLogin", "user not found", err))
		return
	}
	if user == nil {
		// Avoid 401 etc, to keep browsers from throwing out basic auth
		securityLog.Info("Failed login attempt")
		c.JSON(http.StatusOK, loginResponseTO)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(credentialsTO.Password))
	if err != nil {
		securityLog.Info("Password mismatch")
		// Avoid 401 etc, to keep browsers from throwing out basic auth
		c.JSON(http.StatusOK, loginResponseTO)
		return
	}

	sessionID := util.MakeRandomURLSafeB64(21)

	session := models.UserSession{
		UserSessionID: sessionID,
		AppUserID:     user.AppUserID,
		TimeoutAt:     time.Now().Add(time.Minute * 60),
	}
	ctx, cancelTimeout = db.DefaultQueryContext()
	defer cancelTimeout()
	err = session.Insert(ctx, tx, boil.Infer())
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("PostLogin", "cannot insert session", err))
		return
	}

	setSessionCookie(c, sessionID)
	loginResponseTO.LoggedIn = true
	c.JSON(http.StatusOK, loginResponseTO)
}

func PostLogout(c *gin.Context) {
	setSessionCookie(c, "")
	requestContext := ginext.GetRequestContext(c)
	tx := requestContext.Tx
	authentication := requestContext.Authentication
	var userSession *models.UserSession
	if authentication != nil {
		userSession = authentication.UserSession
	}

	if userSession == nil {
		c.AbortWithError(http.StatusBadRequest, util.Error("PostLogout", "logout without session present"))
		return
	}
	ctx, cancelTimeout := db.DefaultQueryContext()
	defer cancelTimeout()
	rows, err := userSession.Delete(ctx, tx)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("PostLogout", "could not delete session", err))
		return
	}
	if rows != 1 {
		c.AbortWithError(http.StatusInternalServerError, util.Error("PostLogout", fmt.Sprintf("too many rows affected: %d", rows)))
		return
	}

	c.Status(http.StatusOK)
}

func setSessionCookie(c *gin.Context, sessionID string) {
	maxAge := 60 * 60
	if sessionID == "" {
		maxAge = -1
	}
	c.SetCookie("SESSION_ID", sessionID, maxAge, "", "", true, true)
	c.SetSameSite(http.SameSiteStrictMode)
}
