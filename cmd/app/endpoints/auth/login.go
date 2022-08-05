package auth

import (
	"context"
	ginext "user-manager/cmd/app/gin-extensions"
	session_service "user-manager/cmd/app/services/session"
	"user-manager/db"
	"user-manager/db/generated/models"
	app_user "user-manager/domain-model/id/appUser"
	"user-manager/util"
	"user-manager/util/slices"

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

	user, err := db.Fetch(func(ctx context.Context) (*models.AppUser, error) {
		return models.AppUsers(
			models.AppUserWhere.Email.EQ(requestTO.Email),
			qm.Load(models.AppUserRels.AppUserRoles),
			qm.Load(models.AppUserRels.TwoFactorThrottling),
		).One(ctx, tx)
	})
	if err != nil {
		return nil, util.Wrap("error finding user", err)
	}
	if user == nil {
		securityLog.Info("Login attempt for non-existant user")
		return &LoginResponseTO{InvalidCredentials}, nil
	}

	hasNonUserRole := slices.Any(user.R.AppUserRoles, func(role *models.AppUserRole) bool { return role.Role != models.UserRoleUSER })
	if hasNonUserRole {
		securityLog.Info("Login attempt without second factor for non-user %d", user.AppUserID)
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
