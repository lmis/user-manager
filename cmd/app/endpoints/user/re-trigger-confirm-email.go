package user

import (
	ginext "user-manager/cmd/app/gin-extensions"
	email_service "user-manager/cmd/app/services/email"
	user_service "user-manager/cmd/app/services/user"
	"user-manager/util"

	"github.com/gin-gonic/gin"
	"github.com/volatiletech/null/v8"
)

type RetriggerConfirmationEmailStatus string

type RetriggerConfirmationEmailResponseTO struct {
	Sent bool `json:"sent"`
}

func PostRetriggerConfirmationEmail(requestContext *ginext.RequestContext, _ *gin.Context) (*RetriggerConfirmationEmailResponseTO, error) {
	securityLog := requestContext.SecurityLog
	user := requestContext.Authentication.AppUser

	if user.EmailVerified {
		securityLog.Info("Email already verified")
		return &RetriggerConfirmationEmailResponseTO{Sent: false}, nil
	}

	user.EmailVerificationToken = null.StringFrom(util.MakeRandomURLSafeB64(21))

	if err := user_service.UpdateUser(requestContext, user); err != nil {
		return nil, util.Wrap("issue persisting user", err)
	}

	if err := email_service.SendVerificationEmail(requestContext, user); err != nil {
		return nil, util.Wrap("error sending verification email", err)
	}

	return &RetriggerConfirmationEmailResponseTO{Sent: true}, nil
}
