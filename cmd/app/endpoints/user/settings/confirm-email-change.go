package settings

import (
	ginext "user-manager/cmd/app/gin-extensions"
	user_service "user-manager/cmd/app/services/user"
	"user-manager/util"

	"github.com/gin-gonic/gin"
	"github.com/volatiletech/null/v8"
)

type EmailChangeConfirmationTO struct {
	Token string `json:"token"`
}

type EmailChangeStatus string

const (
	NoEmailChangeInProgress EmailChangeStatus = "no-change-in-progress"
	InvalidToken            EmailChangeStatus = "invalid-token"
	NewEmailConfirmed       EmailChangeStatus = "new-email-confirmed"
)

type EmailChangeConfirmationResponseTO struct {
	Status EmailChangeStatus `json:"status"`
	Email  string            `json:"email"`
}

func PostConfirmEmailChange(requestContext *ginext.RequestContext, request *EmailChangeConfirmationTO, _ *gin.Context) (*EmailChangeConfirmationResponseTO, error) {
	securityLog := requestContext.SecurityLog
	user := requestContext.Authentication.AppUser

	if !user.NewEmail.Valid {
		return &EmailChangeConfirmationResponseTO{
			Status: NoEmailChangeInProgress,
			Email:  user.Email,
		}, nil
	}

	if !user.EmailVerificationToken.Valid {
		return nil, util.Error("no verification token present on database")
	}

	if request.Token != user.EmailVerificationToken.String {
		securityLog.Info("Invalid email verification token")
		return &EmailChangeConfirmationResponseTO{
			Status: InvalidToken,
			Email:  user.Email,
		}, nil
	}

	user.EmailVerificationToken = null.StringFromPtr(nil)
	user.Email = user.NewEmail.String
	user.NewEmail = null.StringFromPtr(nil)

	if err := user_service.UpdateUser(requestContext, user); err != nil {
		return nil, util.Wrap("issue persisting user", err)
	}

	return &EmailChangeConfirmationResponseTO{
		Status: NewEmailConfirmed,
		Email:  user.Email,
	}, nil
}
