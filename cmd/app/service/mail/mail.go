package mail

import (
	"context"
	"embed"
	"go.mongodb.org/mongo-driver/mongo"
	"html/template"
	"strings"
	dm "user-manager/domain-model"
	"user-manager/util/db"
	"user-manager/util/errs"
)

type TemplateData struct {
	AppUrl      string
	ServiceName string
	Name        string
	NewEmail    string
	Token       string
}

const templatesPattern = "templates/*"

//go:embed templates/layout.tmpl
var layoutFS embed.FS
var layoutTemplate *template.Template

func init() {
	layoutTemplate = template.Must(template.New("layout.tmpl").ParseFS(layoutFS, templatesPattern))

	emailVerificationTemplate = template.Must(template.Must(layoutTemplate.Clone()).ParseFS(emailVerificationFS, templatesPattern))
	signUpAttemptedTemplate = template.Must(template.Must(layoutTemplate.Clone()).ParseFS(signUpAttemptedFS, templatesPattern))
	emailChangeVerificationTemplate = template.Must(template.Must(layoutTemplate.Clone()).ParseFS(emailChangeVerificationFS, templatesPattern))
	emailChangeNotificationTemplate = template.Must(template.Must(layoutTemplate.Clone()).ParseFS(emailChangeNotificationFS, templatesPattern))
	passwordResetTemplate = template.Must(template.Must(layoutTemplate.Clone()).ParseFS(passwordResetFS, templatesPattern))
}

//go:embed templates/email-verification.tmpl
var emailVerificationFS embed.FS
var emailVerificationTemplate *template.Template

func SendVerificationEmail(ctx context.Context, r *dm.RequestContext, email string, verificationToken string) error {
	config := r.Config

	if err := enqueueBasicEmail(
		ctx,
		r.Database,
		emailVerificationTemplate,
		TemplateData{
			AppUrl:      config.AppUrl,
			ServiceName: config.ServiceName,
			Name:        r.User.Name,
			Token:       verificationToken,
		},
		config.EmailFrom,
		email,
		dm.MailQueuePrioHigh,
	); err != nil {
		return errs.Wrap("error enqueuing basic email", err)
	}
	return nil
}

//go:embed templates/sign-up-attempted.tmpl
var signUpAttemptedFS embed.FS
var signUpAttemptedTemplate *template.Template

func SendSignUpAttemptEmail(ctx context.Context, r *dm.RequestContext, email string) error {
	config := r.Config

	if err := enqueueBasicEmail(
		ctx,
		r.Database,
		signUpAttemptedTemplate,
		TemplateData{
			AppUrl:      config.AppUrl,
			ServiceName: config.ServiceName,
			Name:        r.User.Name,
		},
		config.EmailFrom,
		email,
		dm.MailQueuePrioHigh,
	); err != nil {
		return errs.Wrap("error enqueuing basic email", err)
	}
	return nil
}

//go:embed templates/email-change-verification.tmpl
var emailChangeVerificationFS embed.FS
var emailChangeVerificationTemplate *template.Template

func SendChangeVerificationEmail(ctx context.Context, r *dm.RequestContext, newEmail string, verificationToken string) error {
	config := r.Config

	if err := enqueueBasicEmail(
		ctx,
		r.Database,
		emailChangeVerificationTemplate,
		TemplateData{
			AppUrl:      config.AppUrl,
			ServiceName: config.ServiceName,
			Name:        r.User.Name,
			Token:       verificationToken,
			NewEmail:    newEmail,
		},
		config.EmailFrom,
		newEmail,
		dm.MailQueuePrioHigh,
	); err != nil {
		return errs.Wrap("error enqueuing basic email", err)
	}
	return nil
}

//go:embed templates/email-change-notification.tmpl
var emailChangeNotificationFS embed.FS
var emailChangeNotificationTemplate *template.Template

func SendChangeNotificationEmail(ctx context.Context, r *dm.RequestContext, email string, newEmail string) error {
	config := r.Config

	if err := enqueueBasicEmail(
		ctx,
		r.Database,
		emailChangeNotificationTemplate,
		TemplateData{
			AppUrl:      config.AppUrl,
			ServiceName: config.ServiceName,
			Name:        r.User.Name,
			NewEmail:    newEmail,
		},
		config.EmailFrom,
		email,
		dm.MailQueuePrioHigh,
	); err != nil {
		return errs.Wrap("error enqueuing basic email", err)
	}
	return nil
}

//go:embed templates/password-reset.tmpl
var passwordResetFS embed.FS
var passwordResetTemplate *template.Template

func SendResetPasswordEmail(ctx context.Context, r *dm.RequestContext, email string, resetToken string) error {
	config := r.Config

	if err := enqueueBasicEmail(
		ctx,
		r.Database,
		passwordResetTemplate,
		TemplateData{
			AppUrl:      config.AppUrl,
			ServiceName: config.ServiceName,
			Name:        r.User.Name,
			Token:       resetToken,
		},
		config.EmailFrom,
		email,
		dm.MailQueuePrioHigh,
	); err != nil {
		return errs.Wrap("error enqueuing basic email", err)
	}
	return nil
}

func enqueueBasicEmail(
	ctx context.Context,
	database *mongo.Database,
	template *template.Template,
	data TemplateData,
	from string,
	to string,
	priority dm.MailQueuePriority,
) error {
	subjectWriter := &strings.Builder{}
	err := template.ExecuteTemplate(subjectWriter, "subject", data)
	if err != nil {
		return errs.Wrap("error executing subject template", err)
	}

	bodyWriter := &strings.Builder{}
	err = template.ExecuteTemplate(bodyWriter, "body", data)
	if err != nil {
		return errs.Wrap("error executing body template", err)
	}

	if err = insertPendingMail(ctx, database, dm.MailInsert{
		To:       to,
		From:     from,
		Subject:  subjectWriter.String(),
		Body:     bodyWriter.String(),
		Priority: priority,
	}); err != nil {
		return errs.Wrap("issue inserting pending email", err)
	}
	return nil
}

func insertPendingMail(
	ctx context.Context,
	database *mongo.Database,
	mail dm.MailInsert,
) error {
	queryCtx, cancel := db.DefaultQueryContext(ctx)
	defer cancel()

	_, err := database.Collection(dm.MailQueueCollectionName).InsertOne(queryCtx, dm.Mail{
		From:     mail.From,
		To:       mail.To,
		Subject:  mail.Subject,
		Body:     mail.Body,
		Priority: mail.Priority,
		Status:   dm.MailStatusPending,
	})

	if err != nil {
		return errs.Wrap("issue inserting email in db", err)
	}

	return nil
}
