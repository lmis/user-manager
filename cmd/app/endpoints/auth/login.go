package auth

import (
	"database/sql"
	"net/http"
	"time"
	ginext "user-manager/cmd/app/gin-extensions"
	middleware "user-manager/cmd/app/middlewares"
	auth_service "user-manager/cmd/app/services/auth"
	session_service "user-manager/cmd/app/services/session"
	"user-manager/db"
	"user-manager/db/generated/models"
	"user-manager/util"

	"github.com/gin-gonic/gin"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type LoginResponseTO struct {
	LoggedIn                         bool      `json:"loggedIn"`
	ResubmitWithSecondFactorRequired bool      `json:"resubmitWithSecondFactorRequired"`
	TimeoutUntil                     time.Time `json:"timeoutUntil,omitempty"`
}

func PostLogin(c *gin.Context) {
	requestContext := ginext.GetRequestContext(c)
	tx := requestContext.Tx
	securityLog := requestContext.SecurityLog
	var credentialsTO middleware.LoginCredentialsTO
	if err := c.BindJSON(&credentialsTO); err != nil {
		c.AbortWithError(http.StatusBadRequest, util.Wrap("cannot bind to credentialsTO", err))
		return
	}

	var user *models.AppUser
	ctx, cancelTimeout := db.DefaultQueryContext()
	defer cancelTimeout()

	user, err := models.AppUsers(
		models.AppUserWhere.Email.EQ(credentialsTO.Email),
		qm.Load(models.AppUserRels.TwoFactorThrottling),
	).One(ctx, tx)
	if err != nil {
		if err == sql.ErrNoRows {
			// Avoid 401 etc, to keep browsers from throwing out basic auth
			securityLog.Info("Login attempt for non-existant user")
			c.JSON(http.StatusOK, LoginResponseTO{})
		} else {
			c.AbortWithError(http.StatusInternalServerError, util.Wrap("error finding user", err))
		}
		return
	}

	if err = auth_service.ValidateLogin(requestContext, user, credentialsTO.Password, credentialsTO.SecondFactor); err != nil {
		if validationError, ok := err.(*auth_service.ValidateLoginError); ok {
			securityLog.Info(validationError.Message)
			c.JSON(http.StatusOK, LoginResponseTO{
				ResubmitWithSecondFactorRequired: validationError.ResubmitWithSecondFactorRequired,
				TimeoutUntil:                     validationError.TimeoutUntil,
			})
		} else {
			c.AbortWithError(http.StatusInternalServerError, util.Wrap("issue validating login", err))
		}
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
	if err = session.Insert(ctx, tx, boil.Infer()); err != nil {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("cannot insert session", err))
		return
	}

	securityLog.Info("Login")
	session_service.SetSessionCookie(c, sessionID)
	c.JSON(http.StatusOK, LoginResponseTO{LoggedIn: true})
}
