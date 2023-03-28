package functional_tests

import (
	"strings"
	"time"
	"user-manager/cmd/app/resource"
	"user-manager/cmd/mock-3rd-party-apis/config"
	"user-manager/cmd/mock-3rd-party-apis/util"
	dm "user-manager/domain-model"
	"user-manager/util/errors"
)

func TestPasswordReset(config *config.Config, emails util.Emails, testUser *util.TestUser) error {
	email := testUser.Email
	password := testUser.Password
	newPassword := []byte("hunter13")
	client := util.NewRequestClient(config)

	// Trigger reset for non-existent email
	client.MakeApiRequest("POST", "auth/request-password-reset", resource.SignUpTO{
		Email: "does-not-exist",
	})
	if err := client.AssertLastResponseEq(204, nil); err != nil {
		return errors.Wrap("first reset call response mismatch", err)
	}

	// Trigger reset
	client.MakeApiRequest("POST", "auth/request-password-reset", resource.SignUpTO{
		Email: email,
	})
	if err := client.AssertLastResponseEq(204, nil); err != nil {
		return errors.Wrap("second reset call response mismatch", err)
	}

	// Login with old password should still work
	client.MakeApiRequest("POST", "auth/login", resource.LoginTO{
		Email:    email,
		Password: password,
	})
	if err := client.AssertLastResponseEq(200, resource.LoginResponseTO{Status: resource.LoginResponseLoggedIn}); err != nil {
		return errors.Wrap("login response mismatch", err)
	}

	if !client.HasSessionCookie() {
		return errors.Error("session cookie not found")
	}

	// Check user
	client.MakeApiRequest("GET", "user-info", nil)
	if err := client.AssertLastResponseEq(200, resource.UserInfoTO{Roles: []dm.UserRole{dm.UserRoleUser}, EmailVerified: testUser.EmailVerified, Language: testUser.Language}); err != nil {
		return errors.Wrap("auth role response mismatch", err)
	}

	// Logout
	client.MakeApiRequest("POST", "auth/logout", resource.LogoutTO{})
	if err := client.AssertLastResponseEq(204, nil); err != nil {
		return errors.Wrap("logout response mismatch", err)
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
		return errors.Error("token not found")
	}

	// Set new password with wrong token
	client.MakeApiRequest("POST", "auth/reset-password", resource.ResetPasswordTO{
		Token:       "not-correct",
		NewPassword: newPassword,
	})
	if err := client.AssertLastResponseEq(200, resource.ResetPasswordResponseTO{Status: resource.ResetPasswordResponseInvalid}); err != nil {
		return errors.Wrap("wrong token reset password response mismatch", err)
	}

	// Login with new password should not yet work
	client.MakeApiRequest("POST", "auth/login", resource.LoginTO{
		Email:    email,
		Password: newPassword,
	})
	if err := client.AssertLastResponseEq(200, resource.LoginResponseTO{Status: resource.LoginResponseInvalidCredentials}); err != nil {
		return errors.Wrap("login response mismatch", err)
	}

	// Set new password with right token
	client.MakeApiRequest("POST", "auth/reset-password", resource.ResetPasswordTO{
		Email:       email,
		Token:       token,
		NewPassword: newPassword,
	})
	if err := client.AssertLastResponseEq(200, resource.ResetPasswordResponseTO{Status: resource.ResetPasswordResponseSuccess}); err != nil {
		return errors.Wrap("right token reset password response mismatch", err)
	}

	// Login with old password should no longer work
	client.MakeApiRequest("POST", "auth/login", resource.LoginTO{
		Email:    email,
		Password: password,
	})
	if err := client.AssertLastResponseEq(200, resource.LoginResponseTO{Status: resource.LoginResponseInvalidCredentials}); err != nil {
		return errors.Wrap("login response mismatch", err)
	}

	// Login with new password should work now
	client.MakeApiRequest("POST", "auth/login", resource.LoginTO{
		Email:    email,
		Password: newPassword,
	})
	if err := client.AssertLastResponseEq(200, resource.LoginResponseTO{Status: resource.LoginResponseLoggedIn}); err != nil {
		return errors.Wrap("login response with new password mismatch", err)
	}
	if !client.HasSessionCookie() {
		return errors.Error("session cookie not found")
	}

	// Check user
	client.MakeApiRequest("GET", "user-info", nil)
	if err := client.AssertLastResponseEq(200, resource.UserInfoTO{Roles: []dm.UserRole{dm.UserRoleUser}, EmailVerified: testUser.EmailVerified, Language: testUser.Language}); err != nil {
		return errors.Wrap("auth role response mismatch", err)
	}

	testUser.Password = newPassword
	return nil
}
