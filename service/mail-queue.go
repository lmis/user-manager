package service

import (
	"fmt"
	"strings"
	"text/template"
	dm "user-manager/domain-model"
	"user-manager/repository"
	"user-manager/util/errors"
)

type MailQueueService struct {
	mailQueueRepository *repository.MailQueueRepository
	config              *dm.Config
	translations        map[dm.UserLanguage]dm.Translations
	baseTemplate        *template.Template
}

func ProvideMailQueueService(mailQueueRepository *repository.MailQueueRepository, config *dm.Config, translations map[dm.UserLanguage]dm.Translations, baseTemplate *template.Template) *MailQueueService {
	return &MailQueueService{mailQueueRepository, config, translations, baseTemplate}
}

func (s *MailQueueService) SendVerificationEmail(language dm.UserLanguage, email string, verificationToken string) error {
	translation := s.translations[language]
	config := s.config

	data := map[string]string{
		"AppUrl":                 config.AppUrl,
		"EmailVerificationToken": verificationToken,
		"ServiceName":            config.ServiceName,
	}
	if err := s.enqueueBasicEmail(translation, translation.VerificationEmail, data, config.EmailFrom, email, dm.MailQueuePrioHigh); err != nil {
		return errors.Wrap("error enqueuing basic email", err)
	}
	return nil
}

func (s *MailQueueService) SendSignUpAttemptEmail(language dm.UserLanguage, email string) error {
	translation := s.translations[language]
	config := s.config

	data := map[string]string{
		"ServiceName": config.ServiceName,
	}
	if err := s.enqueueBasicEmail(translation, translation.SignUpAttemptedEmail, data, config.EmailFrom, email, dm.MailQueuePrioHigh); err != nil {
		return errors.Wrap("error enqueuing basic email", err)
	}
	return nil
}

func (s *MailQueueService) SendChangeVerificationEmail(language dm.UserLanguage, newEmail string, verificationToken string) error {
	translation := s.translations[language]
	config := s.config

	data := map[string]string{
		"AppUrl":                       config.AppUrl,
		"EmailChangeVerificationToken": verificationToken,
		"ServiceName":                  config.ServiceName,
		"NewEmail":                     newEmail,
	}
	if err := s.enqueueBasicEmail(translation, translation.ChangeVerificationEmail, data, config.EmailFrom, newEmail, dm.MailQueuePrioHigh); err != nil {
		return errors.Wrap("error enqueuing basic email", err)
	}
	return nil
}

func (s *MailQueueService) SendChangeNotificationEmail(language dm.UserLanguage, email string, newEmail string) error {
	translation := s.translations[language]
	config := s.config

	data := map[string]string{
		"ServiceName": config.ServiceName,
		"NewEmail":    newEmail,
	}
	if err := s.enqueueBasicEmail(translation, translation.ChangeNotificationEmail, data, config.EmailFrom, email, dm.MailQueuePrioHigh); err != nil {
		return errors.Wrap("error enqueuing basic email", err)
	}
	return nil
}

func (s *MailQueueService) SendResetPasswordEmail(language dm.UserLanguage, email string, resetToken string) error {
	translation := s.translations[language]
	config := s.config

	data := map[string]string{
		"ServiceName":        config.ServiceName,
		"AppUrl":             config.AppUrl,
		"PasswordResetToken": resetToken,
	}
	if err := s.enqueueBasicEmail(translation, translation.ResetPasswordEmail, data, config.EmailFrom, email, dm.MailQueuePrioHigh); err != nil {
		return errors.Wrap("error enqueuing basic email", err)
	}
	return nil
}

func (s *MailQueueService) executeTemplate(templateText string, data interface{}) (string, error) {
	t, err := template.New("base").Parse(templateText)
	if err != nil {
		return "", errors.Wrap("templateText cannot be parsed", err)
	}
	writer := &strings.Builder{}
	if err = t.Execute(writer, data); err != nil {
		return "", errors.Wrap("template execution error", err)
	}
	return writer.String(), nil
}

type baseTemplateData struct {
	Salutation string
	Paragraphs []string
	Footer     string
}

func (s *MailQueueService) enqueueBasicEmail(
	translation dm.Translations,
	templates []string,
	data map[string]string,
	from string,
	to string,
	priority dm.MailQueuePriority,
) error {
	baseTemplate := s.baseTemplate
	mailQueueRepository := s.mailQueueRepository

	var err error
	salutation := ""
	_, ok := data["UserName"]
	if ok {
		salutation, err = s.executeTemplate(translation.Salutation, data)
		if err != nil {
			return errors.Wrap("issue translating salutation", err)
		}
	} else {
		salutation = translation.SalutationAnonymous
	}
	subject := ""
	var paragraphs []string
	if len(templates) < 2 {
		return errors.Errorf("invalid template. %s need at least subject and paragraph", templates)
	}
	for i, t := range templates {
		p, err := s.executeTemplate(t, data)
		if err != nil {
			return errors.Wrap(fmt.Sprintf("issue translating template %d", i), err)
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
		return errors.Wrap("issue executing base template", err)
	}

	if err = mailQueueRepository.InsertPending(from, to, writer.String(), subject, priority); err != nil {
		return errors.Wrap("issue inserting pending email", err)
	}
	return nil
}
