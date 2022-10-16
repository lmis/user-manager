package domain_model

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
