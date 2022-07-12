package emailservice

import (
	"user-manager/db/generated/models"
	ginext "user-manager/gin-extensions"
	"user-manager/util"
)

func SendVerificationEmail(r *ginext.RequestContext, user *models.AppUser) error {
	config := r.Config
	lang := user.Language
	translation := translations[lang]

	data := map[string]string{
		"AppUrl":                 config.AppUrl,
		"EmailVerificationToken": user.EmailVerificationToken.String,
		"ServiceName":            config.ServiceName,
	}
	err := enqueueBasicEmail(r, lang, translation.VerificationEmail, data, user.Email, PriorityHigh)
	if err != nil {
		return util.Wrap("SendVerificationEmail", "error enqueuing basic email", err)
	}
	return nil
}

func SendSignUpAttemptEmail(r *ginext.RequestContext, user *models.AppUser) error {
	config := r.Config
	lang := user.Language
	translation := translations[lang]

	data := map[string]string{
		"ServiceName": config.ServiceName,
	}
	err := enqueueBasicEmail(r, lang, translation.SignUpAttemptEmail, data, user.Email, PriorityHigh)
	if err != nil {
		return util.Wrap("SendSignUpAttemptEmail", "error enqueuing basic email", err)
	}
	return nil
}
