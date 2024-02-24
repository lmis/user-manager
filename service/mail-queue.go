package service

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"strings"
	"text/template"
	ginext "user-manager/cmd/app/gin-extensions"
	dm "user-manager/domain-model"
	"user-manager/repository"
	"user-manager/util/errs"
)

func SendVerificationEmail(ctx context.Context, r *ginext.RequestContext, language dm.UserLanguage, email string, verificationToken string) error {
	config := r.Config
	translation := r.Emailing.Translations[language]

	data := map[string]string{
		"AppUrl":                 config.AppUrl,
		"EmailVerificationToken": verificationToken,
		"ServiceName":            config.ServiceName,
	}
	if err := enqueueBasicEmail(ctx, r.Database, r.Emailing.BaseTemplate, translation, translation.VerificationEmail, data, config.EmailFrom, email, dm.MailQueuePrioHigh); err != nil {
		return errs.Wrap("error enqueuing basic email", err)
	}
	return nil
}

func SendSignUpAttemptEmail(ctx context.Context, r *ginext.RequestContext, language dm.UserLanguage, email string) error {
	config := r.Config
	translation := r.Emailing.Translations[language]

	data := map[string]string{
		"ServiceName": config.ServiceName,
	}
	if err := enqueueBasicEmail(ctx, r.Database, r.Emailing.BaseTemplate, translation, translation.SignUpAttemptedEmail, data, config.EmailFrom, email, dm.MailQueuePrioHigh); err != nil {
		return errs.Wrap("error enqueuing basic email", err)
	}
	return nil
}

func SendChangeVerificationEmail(ctx context.Context, r *ginext.RequestContext, language dm.UserLanguage, newEmail string, verificationToken string) error {
	config := r.Config
	translation := r.Emailing.Translations[language]

	data := map[string]string{
		"AppUrl":                       config.AppUrl,
		"EmailChangeVerificationToken": verificationToken,
		"ServiceName":                  config.ServiceName,
		"NewEmail":                     newEmail,
	}
	if err := enqueueBasicEmail(ctx, r.Database, r.Emailing.BaseTemplate, translation, translation.ChangeVerificationEmail, data, config.EmailFrom, newEmail, dm.MailQueuePrioHigh); err != nil {
		return errs.Wrap("error enqueuing basic email", err)
	}
	return nil
}

func SendChangeNotificationEmail(ctx context.Context, r *ginext.RequestContext, language dm.UserLanguage, email string, newEmail string) error {
	config := r.Config
	translation := r.Emailing.Translations[language]

	data := map[string]string{
		"ServiceName": config.ServiceName,
		"NewEmail":    newEmail,
	}
	if err := enqueueBasicEmail(ctx, r.Database, r.Emailing.BaseTemplate, translation, translation.ChangeNotificationEmail, data, config.EmailFrom, email, dm.MailQueuePrioHigh); err != nil {
		return errs.Wrap("error enqueuing basic email", err)
	}
	return nil
}

func SendResetPasswordEmail(ctx context.Context, r *ginext.RequestContext, language dm.UserLanguage, email string, resetToken string) error {
	config := r.Config
	translation := r.Emailing.Translations[language]

	data := map[string]string{
		"ServiceName":        config.ServiceName,
		"AppUrl":             config.AppUrl,
		"PasswordResetToken": resetToken,
	}
	if err := enqueueBasicEmail(ctx, r.Database, r.Emailing.BaseTemplate, translation, translation.ResetPasswordEmail, data, config.EmailFrom, email, dm.MailQueuePrioHigh); err != nil {
		return errs.Wrap("error enqueuing basic email", err)
	}
	return nil
}

func executeTemplate(templateText string, data interface{}) (string, error) {
	t, err := template.New("base").Parse(templateText)
	if err != nil {
		return "", errs.Wrap("templateText cannot be parsed", err)
	}
	writer := &strings.Builder{}
	if err = t.Execute(writer, data); err != nil {
		return "", errs.Wrap("template execution error", err)
	}
	return writer.String(), nil
}

type baseTemplateData struct {
	Salutation string
	Paragraphs []string
	Footer     string
}

func enqueueBasicEmail(
	ctx context.Context,
	database *mongo.Database,
	baseTemplate *template.Template,
	translation dm.Translations,
	templates []string,
	data map[string]string,
	from string,
	to string,
	priority dm.MailQueuePriority,
) error {

	var err error
	salutation := ""
	_, ok := data["UserName"]
	if ok {
		salutation, err = executeTemplate(translation.Salutation, data)
		if err != nil {
			return errs.Wrap("issue translating salutation", err)
		}
	} else {
		salutation = translation.SalutationAnonymous
	}
	subject := ""
	var paragraphs []string
	if len(templates) < 2 {
		return errs.Errorf("invalid template. %s need at least subject and paragraph", templates)
	}
	for i, t := range templates {
		p, err := executeTemplate(t, data)
		if err != nil {
			return errs.Wrap(fmt.Sprintf("issue translating template %d", i), err)
		}
		if i == 0 {
			subject = p
		} else {
			paragraphs = append(paragraphs, p)
		}
	}
	writer := &strings.Builder{}
	if err = baseTemplate.Execute(writer, baseTemplateData{
		Salutation: salutation,
		Paragraphs: paragraphs,
		Footer:     translation.Footer,
	}); err != nil {
		return errs.Wrap("issue executing base template", err)
	}

	if err = repository.InsertPendingMail(ctx, database, dm.MailInsert{
		To:       to,
		From:     from,
		Subject:  subject,
		Content:  writer.String(),
		Priority: priority,
	}); err != nil {
		return errs.Wrap("issue inserting pending email", err)
	}
	return nil
}
