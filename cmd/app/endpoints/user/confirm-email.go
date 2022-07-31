package user

import (
	ginext "user-manager/cmd/app/gin-extensions"
	user_service "user-manager/cmd/app/services/user"
	"user-manager/util"

	"github.com/gin-gonic/gin"
	"github.com/volatiletech/null/v8"
)

type EmailConfirmationTO struct {
	Token string `json:"token"`
}

type EmailConfirmationStatus string

const (
	AlreadyConfirmed EmailConfirmationStatus = "already-confirmed"
	NewlyConfirmed   EmailConfirmationStatus = "newly-confirmed"
	InvalidToken     EmailConfirmationStatus = "invalid-token"
)

type EmailConfirmationResponseTO struct {
	Status EmailConfirmationStatus `json:"status"`
}

func PostConfirmEmail(requestContext *ginext.RequestContext, request *EmailConfirmationTO, _ *gin.Context) (*EmailConfirmationResponseTO, error) {
	securityLog := requestContext.SecurityLog
	user := requestContext.Authentication.AppUser

	if user.EmailVerified {
		securityLog.Info("Email already verified")
		return &EmailConfirmationResponseTO{
			Status: AlreadyConfirmed,
		}, nil
	}

	if !user.EmailVerificationToken.Valid {
		return nil, util.Error("no verification token present on database")
	}

	if request.Token != user.EmailVerificationToken.String {
		securityLog.Info("Invalid email verification token")
		return &EmailConfirmationResponseTO{
			Status: InvalidToken,
		}, nil
	}

	user.EmailVerificationToken = null.StringFromPtr(nil)
	user.EmailVerified = true

	if err := user_service.UpdateUser(requestContext, user); err != nil {
		return nil, util.Wrap("issue persisting user", err)
	}

	return &EmailConfirmationResponseTO{
		Status: NewlyConfirmed,
	}, nil
}
