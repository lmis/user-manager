package domain_model

import "text/template"

type Translations struct {
	Salutation              string   `yaml:"salutation"`
	SalutationAnonymous     string   `yaml:"salutationAnonymous"`
	VerificationEmail       []string `yaml:"verificationEmail"`
	SignUpAttemptedEmail    []string `yaml:"signUpAttemptedEmail"`
	ChangeVerificationEmail []string `yaml:"changeVerificationEmail"`
	ChangeNotificationEmail []string `yaml:"changeNotificationEmail"`
	ResetPasswordEmail      []string `yaml:"resetPasswordEmail"`
	Footer                  string   `yaml:"footer"`
}

type TranslationsByLang map[UserLanguage]Translations

type Emailing struct {
	BaseTemplate *template.Template
	Translations TranslationsByLang
}
