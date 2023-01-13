package functional_tests

import (
	"user-manager/cmd/app/resource"
	"user-manager/cmd/mock-3rd-party-apis/config"
	"user-manager/cmd/mock-3rd-party-apis/util"
	domain_model "user-manager/domain-model"
	"user-manager/util/errors"
)

func TestSimpleLogin(config *config.Config, emails util.Emails, testUser *util.TestUser) error {
	email := testUser.Email
	password := testUser.Password
	client := util.NewRequestClient(config)

	// Login with wrong email
	client.MakeApiRequest("POST", "auth/login", resource.LoginTO{
		Email:    "another-email",
		Password: password,
	})
	if err := client.AssertLastResponseEq(200, resource.LoginResponseTO{Status: resource.LOGIN_RESPONSE_INVALID_CREDENTIALS}); err != nil {
		return errors.Wrap("login with wrong email response mismatch", err)
	}
	if client.HasSessionCookie() {
		return errors.Error("unexpected session cookie returned despite wrong email")
	}

	// Login with wrong password
	client.MakeApiRequest("POST", "auth/login", resource.LoginTO{
		Email:    email,
		Password: []byte("not-the-password"),
	})
	if err := client.AssertLastResponseEq(200, resource.LoginResponseTO{Status: resource.LOGIN_RESPONSE_INVALID_CREDENTIALS}); err != nil {
		return errors.Wrap("login with wrong password response mismatch", err)
	}

	if client.HasSessionCookie() {
		return errors.Error("unexpected session cookie returned dspite wrong password")
	}

	// Get user info
	client.MakeApiRequest("GET", "user-info", nil)
	if err := client.AssertLastResponseEq(200, resource.UserInfoTO{Roles: nil, EmailVerified: false}); err != nil {
		return errors.Wrap("get user info response mismatch", err)
	}

	// Login with correct info
	client.MakeApiRequest("POST", "auth/login", resource.LoginTO{
		Email:    email,
		Password: password,
	})
	if err := client.AssertLastResponseEq(200, resource.LoginResponseTO{Status: resource.LOGIN_RESPONSE_LOGGED_IN}); err != nil {
		return errors.Wrap("login with correct info response mismatch", err)
	}
	if !client.HasSessionCookie() {
		return errors.Error("expected session cookie returned, got none")
	}

	// Get user info
	client.MakeApiRequest("GET", "user-info", nil)
	if err := client.AssertLastResponseEq(200, resource.UserInfoTO{Roles: []domain_model.UserRole{domain_model.USER_ROLE_USER}, EmailVerified: testUser.EmailVerified, Language: testUser.Language}); err != nil {
		return errors.Wrap("second get user info response mismatch", err)
	}
	return nil
}
