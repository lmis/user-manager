package functional_tests

import (
	"user-manager/cmd/app/resource"
	"user-manager/cmd/mock-3rd-party-apis/config"
	"user-manager/cmd/mock-3rd-party-apis/util"
	dm "user-manager/domain-model"
	"user-manager/util/errs"
)

func TestCallWithMismatchingCsrfTokens(config *config.Config, _ util.Emails, testUser *util.TestUser) error {
	email := testUser.Email
	password := testUser.Password
	client := util.NewRequestClient(config)
	client.SetCsrfTokens("some", "other")

	// Check user with mismatching CSRF tokens
	client.MakeApiRequest("GET", "user-info", nil)

	if err := client.AssertLastResponseEq(400, nil); err != nil {
		return errs.Wrap("auth user response mismatch", err)
	}

	// Login (with matching CSRF tokens)
	client.SetCsrfTokens("some", "some")
	client.MakeApiRequest("POST", "auth/login", resource.LoginTO{
		Email:    email,
		Password: password,
	})
	if err := client.AssertLastResponseEq(200, resource.LoginResponseTO{Status: resource.LoginResponseLoggedIn}); err != nil {
		return errs.Wrap("login response mismatch", err)
	}

	if !client.HasSessionCookie() {
		return errs.Error("session cookie not found")
	}

	// Change language settings (logged in, with wrong tokens)
	var otherLanguage dm.UserLanguage
	for _, lang := range dm.AllUserLanguages() {
		if lang != testUser.Language {
			otherLanguage = lang
		}
	}
	client.SetCsrfTokens("some", "other")
	client.MakeApiRequest("POST", "user/settings/language", &resource.LanguageTO{Language: otherLanguage})
	if err := client.AssertLastResponseEq(400, nil); err != nil {
		return errs.Wrap("language response mismatch", err)
	}
	return nil
}
