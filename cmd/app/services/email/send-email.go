package emailservice

import (
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/db/generated/models"
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
	if err := enqueueBasicEmail(r, lang, translation.VerificationEmail, data, config.EmailFrom, user.Email, PriorityHigh); err != nil {
		return util.Wrap("error enqueuing basic email", err)
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
	if err := enqueueBasicEmail(r, lang, translation.SignUpAttemptEmail, data, config.EmailFrom, user.Email, PriorityHigh); err != nil {
		return util.Wrap("error enqueuing basic email", err)
	}
	return nil
}
