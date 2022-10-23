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

func (s *MailQueueService) SendVerificationEmail(language domain_model.UserLanguage, email string, verificationToken string) error {
	translation := s.translations[language]
	config := s.config

	data := map[string]string{
		"AppUrl":                 config.AppUrl,
		"EmailVerificationToken": verificationToken,
		"ServiceName":            config.ServiceName,
	}
	if err := s.enqueueBasicEmail(translation, translation.VerificationEmail, data, config.EmailFrom, email, domain_model.MAIL_QUEUE_PRIO_HIGH); err != nil {
		return util.Wrap("error enqueuing basic email", err)
	}
	return nil
}

func (s *MailQueueService) SendSignUpAttemptEmail(language domain_model.UserLanguage, email string) error {
	translation := s.translations[language]
	config := s.config

	data := map[string]string{
		"ServiceName": config.ServiceName,
	}
	if err := s.enqueueBasicEmail(translation, translation.SignUpAttemptedEmail, data, config.EmailFrom, email, domain_model.MAIL_QUEUE_PRIO_HIGH); err != nil {
		return util.Wrap("error enqueuing basic email", err)
	}
	return nil
}

func (s *MailQueueService) SendChangeVerificationEmail(language domain_model.UserLanguage, newEmail string, verificationToken string) error {
	translation := s.translations[language]
	config := s.config

	data := map[string]string{
		"AppUrl":                       config.AppUrl,
		"EmailChangeVerificationToken": verificationToken,
		"ServiceName":                  config.ServiceName,
		"NewEmail":                     newEmail,
	}
	if err := s.enqueueBasicEmail(translation, translation.ChangeVerificationEmail, data, config.EmailFrom, newEmail, domain_model.MAIL_QUEUE_PRIO_HIGH); err != nil {
		return util.Wrap("error enqueuing basic email", err)
	}
	return nil
}

func (s *MailQueueService) SendChangeNotificationEmail(language domain_model.UserLanguage, email string, newEmail string) error {
	translation := s.translations[language]
	config := s.config

	data := map[string]string{
		"ServiceName": config.ServiceName,
		"NewEmail":    newEmail,
	}
	if err := s.enqueueBasicEmail(translation, translation.ChangeNotificationEmail, data, config.EmailFrom, email, domain_model.MAIL_QUEUE_PRIO_HIGH); err != nil {
		return util.Wrap("error enqueuing basic email", err)
	}
	return nil
}

func (s *MailQueueService) SendResetPasswordEmail(language domain_model.UserLanguage, email string, resetToken string) error {
	translation := s.translations[language]
	config := s.config

	data := map[string]string{
		"ServiceName":        config.ServiceName,
		"AppUrl":             config.AppUrl,
		"PasswordResetToken": resetToken,
	}
	if err := s.enqueueBasicEmail(translation, translation.ResetPasswordEmail, data, config.EmailFrom, email, domain_model.MAIL_QUEUE_PRIO_HIGH); err != nil {
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
	baseTemplate := s.baseTemplate
	mailQueueRepository := s.mailQueueRepository

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
		return util.Errorf("invalid template. %s need at least subject and paragraph", templates)
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
	if err = baseTemplate.Execute(writer, baseTemplateData{
		Salutation: salutation,
		Paragraphs: paragraphs,
		Footer:     translation.Footer,
	}); err != nil {
		return util.Wrap("issue executing base template", err)
	}

	if err = mailQueueRepository.InsertPending(from, to, writer.String(), subject, priority); err != nil {
		return util.Wrap("issue inserting pending email", err)
	}
	return nil
}
