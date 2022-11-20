package functional_tests

import (
	"user-manager/cmd/app/resource"
	"user-manager/cmd/mock-3rd-party-apis/config"
	mock_util "user-manager/cmd/mock-3rd-party-apis/util"
	domain_model "user-manager/domain-model"
	"user-manager/util"
)

func TestCallWithMismatchingCsrfTokens(config *config.Config, emails mock_util.Emails, testUser *mock_util.TestUser) error {
	email := testUser.Email
	password := testUser.Password
	client := mock_util.NewRequestClient(config)
	client.SetCsrfTokens("some", "other")

	// Check user with mismatching CSRF tokens
	client.MakeApiRequest("GET", "user-info", nil)

	if err := client.AssertLastResponseEq(400, nil); err != nil {
		return util.Wrap("auth user response mismatch", err)
	}

	// Login (with matching CSRF tokens)
	client.SetCsrfTokens("some", "some")
	client.MakeApiRequest("POST", "auth/login", resource.LoginTO{
		Email:    email,
		Password: password,
	})
	if err := client.AssertLastResponseEq(200, resource.LoginResponseTO{Status: resource.LOGIN_RESPONSE_LOGGED_IN}); err != nil {
		return util.Wrap("login response mismatch", err)
	}

	if !client.HasSessionCookie() {
		return util.Error("session cookie not found")
	}

	// Change language settings (logged in, with wrong tokens)
	var otherLanguage domain_model.UserLanguage
	for _, lang := range domain_model.AllUserLanguages() {
		if lang != testUser.Language {
			otherLanguage = lang
		}
	}
	client.SetCsrfTokens("some", "other")
	client.MakeApiRequest("POST", "user/settings/language", &resource.LanguageTO{Language: otherLanguage})
	if err := client.AssertLastResponseEq(400, nil); err != nil {
		return util.Wrap("language response mismatch", err)
	}
	return nil
}
