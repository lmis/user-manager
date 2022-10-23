package functional_tests

import (
	"strings"
	"time"
	"user-manager/cmd/app/resource"
	"user-manager/cmd/mock-3rd-party-apis/config"
	mock_util "user-manager/cmd/mock-3rd-party-apis/util"
	domain_model "user-manager/domain-model"
	"user-manager/util"
)

func TestPasswordReset(config *config.Config, emails mock_util.Emails, testUser *mock_util.TestUser) error {
	email := testUser.Email
	password := testUser.Password
	newPassword := []byte("hunter13")
	client := mock_util.NewRequestClient(config)

	// Trigger reset for non-existant email
	client.MakeApiRequest("POST", "auth/request-password-reset", resource.SignUpTO{
		Email: "does-not-exist",
	})
	if err := client.AssertLastResponseEq(204, nil); err != nil {
		return util.Wrap("first reset call response mismatch", err)
	}

	// Trigger reset
	client.MakeApiRequest("POST", "auth/request-password-reset", resource.SignUpTO{
		Email: email,
	})
	if err := client.AssertLastResponseEq(204, nil); err != nil {
		return util.Wrap("second reset call response mismatch", err)
	}

	// Login with old password should still work
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

	// Check user
	client.MakeApiRequest("GET", "user-info", nil)
	if err := client.AssertLastResponseEq(200, resource.UserInfoTO{Roles: []domain_model.UserRole{domain_model.USER_ROLE_USER}, EmailVerified: testUser.EmailVerified, Language: testUser.Language}); err != nil {
		return util.Wrap("auth role response mismatch", err)
	}

	// Logout
	client.MakeApiRequest("POST", "auth/logout", resource.LogoutTO{})
	if err := client.AssertLastResponseEq(204, nil); err != nil {
		return util.Wrap("logout response mismatch", err)
	}

	// Grab token from email
	token := ""
	for i := 0; token == "" && i < 10; i++ {
		for _, e := range emails[email] {
			// TODO: make independent of language
			if e.Subject == "Passwort zurücksetzen" {
				token = strings.TrimSpace(strings.Split(strings.Split(e.Body, "password-reset?token=")[1], " ")[0])
			}
		}
		if token == "" {
			time.Sleep(500 * time.Millisecond)
		}
	}

	if token == "" {
		return util.Error("token not found")
	}

	// Set new password with wrong token
	client.MakeApiRequest("POST", "auth/reset-password", resource.ResetPasswordTO{
		Token:       "not-correct",
		NewPassword: newPassword,
	})
	if err := client.AssertLastResponseEq(200, resource.ResetPasswordResponseTO{Status: resource.RESET_PASSWORD_RESPONSE_INVALID}); err != nil {
		return util.Wrap("wrong token reset password response mismatch", err)
	}

	// Login with new password should not yet work
	client.MakeApiRequest("POST", "auth/login", resource.LoginTO{
		Email:    email,
		Password: newPassword,
	})
	if err := client.AssertLastResponseEq(200, resource.LoginResponseTO{Status: resource.LOGIN_RESPONSE_INVALID_CREDENTIALS}); err != nil {
		return util.Wrap("login response mismatch", err)
	}

	// Set new password with right token
	client.MakeApiRequest("POST", "auth/reset-password", resource.ResetPasswordTO{
		Email:       email,
		Token:       token,
		NewPassword: newPassword,
	})
	if err := client.AssertLastResponseEq(200, resource.ResetPasswordResponseTO{Status: resource.RESET_PASSWORD_RESPONSE_SUCCESS}); err != nil {
		return util.Wrap("right token reset password response mismatch", err)
	}

	// Login with old password should no longer work
	client.MakeApiRequest("POST", "auth/login", resource.LoginTO{
		Email:    email,
		Password: password,
	})
	if err := client.AssertLastResponseEq(200, resource.LoginResponseTO{Status: resource.LOGIN_RESPONSE_INVALID_CREDENTIALS}); err != nil {
		return util.Wrap("login response mismatch", err)
	}

	// Login with new password should work now
	client.MakeApiRequest("POST", "auth/login", resource.LoginTO{
		Email:    email,
		Password: newPassword,
	})
	if err := client.AssertLastResponseEq(200, resource.LoginResponseTO{Status: resource.LOGIN_RESPONSE_LOGGED_IN}); err != nil {
		return util.Wrap("login response with new password mismatch", err)
	}
	if !client.HasSessionCookie() {
		return util.Error("session cookie not found")
	}

	// Check user
	client.MakeApiRequest("GET", "user-info", nil)
	if err := client.AssertLastResponseEq(200, resource.UserInfoTO{Roles: []domain_model.UserRole{domain_model.USER_ROLE_USER}, EmailVerified: testUser.EmailVerified, Language: testUser.Language}); err != nil {
		return util.Wrap("auth role response mismatch", err)
	}

	testUser.Password = newPassword
	return nil
}
