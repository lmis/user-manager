package service

import (
	"fmt"
	"strings"
	"text/template"
	domain_model "user-manager/domain-model"
	"user-manager/repository"
	"user-manager/util"
)

type MailQueueService struct {
	mailQueueRepository *repository.MailQueueRepository
	config              *domain_model.Config
	translations        map[domain_model.UserLanguage]domain_model.Translations
	baseTemplate        *template.Template
}

func ProvideMailQueueService(mailQueueRepository *repository.MailQueueRepository, config *domain_model.Config, translations map[domain_model.UserLanguage]domain_model.Translations, baseTemplate *template.Template) *MailQueueService {
	return &MailQueueService{mailQueueRepository, config, translations, baseTemplate}
}

func (s *MailQueueService) SendVerificationEmail(user *domain_model.AppUser) error {
	lang := user.Language
	translation := s.translations[lang]

	data := map[string]string{
		"AppUrl":                 s.config.AppUrl,
		"EmailVerificationToken": user.EmailVerificationToken.OrPanic(),
		"ServiceName":            s.config.ServiceName,
	}
	if err := s.enqueueBasicEmail(translation, translation.VerificationEmail, data, s.config.EmailFrom, user.Email, domain_model.MAIL_QUEUE_PRIO_HIGH); err != nil {
		return util.Wrap("error enqueuing basic email", err)
	}
	return nil
}

func (s *MailQueueService) SendSignUpAttemptEmail(user *domain_model.AppUser) error {
	lang := user.Language
	translation := s.translations[lang]

	data := map[string]string{
		"ServiceName": s.config.ServiceName,
	}
	if err := s.enqueueBasicEmail(translation, translation.SignUpAttemptedEmail, data, s.config.EmailFrom, user.Email, domain_model.MAIL_QUEUE_PRIO_HIGH); err != nil {
		return util.Wrap("error enqueuing basic email", err)
	}
	return nil
}

func (s *MailQueueService) SendChangeVerificationEmail(user *domain_model.AppUser) error {
	lang := user.Language
	translation := s.translations[lang]

	data := map[string]string{
		"AppUrl":                       s.config.AppUrl,
		"EmailChangeVerificationToken": user.EmailVerificationToken.OrPanic(),
		"ServiceName":                  s.config.ServiceName,
		"NewEmail":                     user.NewEmail.OrPanic(),
	}
	if err := s.enqueueBasicEmail(translation, translation.ChangeVerificationEmail, data, s.config.EmailFrom, user.NewEmail.OrPanic(), domain_model.MAIL_QUEUE_PRIO_HIGH); err != nil {
		return util.Wrap("error enqueuing basic email", err)
	}
	return nil
}

func (s *MailQueueService) SendChangeNotificationEmail(user *domain_model.AppUser) error {
	lang := user.Language
	translation := s.translations[lang]

	data := map[string]string{
		"ServiceName": s.config.ServiceName,
		"NewEmail":    user.NewEmail.OrPanic(),
	}
	if err := s.enqueueBasicEmail(translation, translation.ChangeNotificationEmail, data, s.config.EmailFrom, user.Email, domain_model.MAIL_QUEUE_PRIO_HIGH); err != nil {
		return util.Wrap("error enqueuing basic email", err)
	}
	return nil
}

func (s *MailQueueService) SendResetPasswordEmail(user *domain_model.AppUser) error {
	lang := user.Language
	translation := s.translations[lang]

	data := map[string]string{
		"ServiceName":        s.config.ServiceName,
		"AppUrl":             s.config.AppUrl,
		"PasswordResetToken": user.PasswordResetToken.OrPanic(),
	}
	if err := s.enqueueBasicEmail(translation, translation.ChangeNotificationEmail, data, s.config.EmailFrom, user.Email, domain_model.MAIL_QUEUE_PRIO_HIGH); err != nil {
		return util.Wrap("error enqueuing basic email", err)
	}
	return nil
}

func (r *MailQueueService) executeTemplate(templateText string, data interface{}) (string, error) {
	t, err := template.New("base").Parse(templateText)
	if err != nil {
		return "", util.Wrap("templateText canot be parsed", err)
	}
	writer := &strings.Builder{}
	if err = t.Execute(writer, data); err != nil {
		return "", util.Wrap("template execution error", err)
	}
	return writer.String(), nil
}

type baseTemplateData struct {
	Salutation string
	Paragraphs []string
	Footer     string
}

func (s *MailQueueService) enqueueBasicEmail(
	translation domain_model.Translations,
	templates []string,
	data map[string]string,
	from string,
	to string,
	priority domain_model.MailQueuePriority,
) error {
	var err error
	salutation := ""
	_, ok := data["UserName"]
	if ok {
		salutation, err = s.executeTemplate(translation.Salutation, data)
		if err != nil {
			return util.Wrap("issue translating salutation", err)
		}
	} else {
		salutation = translation.SalutationAnonymous
	}
	subject := ""
	paragraphs := []string{}
	if len(templates) < 2 {
		return util.Errorf("invalid template")
	}
	for i, t := range templates {
		p, err := s.executeTemplate(t, data)
		if err != nil {
			return util.Wrap(fmt.Sprintf("issue translating template %d", i), err)
		}
		if i == 0 {
			subject = p
		} else {
			paragraphs = append(paragraphs, p)
		}
	}
	writer := &strings.Builder{}
	if err = s.baseTemplate.Execute(writer, baseTemplateData{
		Salutation: salutation,
		Paragraphs: paragraphs,
		Footer:     translation.Footer,
	}); err != nil {
		return util.Wrap("issue executing base template", err)
	}

	if err = s.mailQueueRepository.InsertPending(from, to, writer.String(), subject, priority); err != nil {
		return util.Wrap("issue inserting pending email", err)
	}
	return nil
}
