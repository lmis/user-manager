package auth

import (
	"database/sql"
	ginext "user-manager/cmd/app/gin-extensions"
	session_service "user-manager/cmd/app/services/session"
	"user-manager/db"
	"user-manager/db/generated/models"
	app_user "user-manager/domain-model/id/appUser"
	"user-manager/util"

	"github.com/gin-gonic/gin"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/crypto/bcrypt"
)

type LoginTO struct {
	Email    string `json:"email"`
	Password []byte `json:"password"`
}

type LoginResponseStatus string

const (
	LoggedIn             LoginResponseStatus = "logged-in"
	SecondFactorRequired LoginResponseStatus = "second-factor-allowed"
	InvalidCredentials   LoginResponseStatus = "invalid-credentials"
)

type LoginResponseTO struct {
	Status LoginResponseStatus `json:"status"`
}

func PostLogin(requestContext *ginext.RequestContext, requestTO *LoginTO, c *gin.Context) (*LoginResponseTO, error) {
	tx := requestContext.Tx
	securityLog := requestContext.SecurityLog

	var user *models.AppUser
	ctx, cancelTimeout := db.DefaultQueryContext()
	defer cancelTimeout()

	user, err := models.AppUsers(
		models.AppUserWhere.Email.EQ(requestTO.Email),
		qm.Load(models.AppUserRels.AppUserRoles),
		qm.Load(models.AppUserRels.TwoFactorThrottling),
	).One(ctx, tx)
	if err != nil {
		if err == sql.ErrNoRows {
			securityLog.Info("Login attempt for non-existant user")
			return &LoginResponseTO{InvalidCredentials}, nil
		}
		return nil, util.Wrap("error finding user", err)
	}
	hasOnlyUserRole := true
	for _, role := range user.R.AppUserRoles {
		if role.Role != models.UserRoleUSER {
			hasOnlyUserRole = false
		}
	}

	if !hasOnlyUserRole {
		securityLog.Info("Login attempt without second factor for non-user %s", user.AppUserID)
		return &LoginResponseTO{InvalidCredentials}, nil
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), requestTO.Password); err != nil {
		securityLog.Info("Password mismatch for user %s", user.AppUserID)
		return &LoginResponseTO{InvalidCredentials}, nil
	}

	if user.TwoFactorToken.Valid {
		return &LoginResponseTO{SecondFactorRequired}, nil
	}

	securityLog.Info("Login")
	sessionId, err := session_service.InsertSession(requestContext, models.UserSessionTypeLOGIN, app_user.ID(user.AppUserID), session_service.LoginSessionDuriation)
	if err != nil {
		return nil, util.Wrap("error inserting session", err)
	}

	session_service.SetSessionCookie(c, sessionId, models.UserSessionTypeLOGIN)
	return &LoginResponseTO{LoggedIn}, nil
}
