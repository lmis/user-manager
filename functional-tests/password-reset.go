package functional_tests

import (
	"strings"
	"time"
	"user-manager/cmd/app/resource"
	dm "user-manager/domain-model"
	"user-manager/functional-tests/helper"
	"user-manager/util/errs"
)

func TestPasswordReset(testUser *helper.TestUser) error {
	email := testUser.Email
	password := testUser.Password
	newPassword := []byte("hunter13")
	client := helper.NewRequestClient(testUser)

	// Trigger reset for non-existent email
	client.MakeApiRequest("POST", "auth/request-password-reset", resource.SignUpTO{
		Email: "does-not-exist",
	})
	if err := client.AssertLastResponseEq(204, nil); err != nil {
		return errs.Wrap("first reset call response mismatch", err)
	}

	// Trigger reset
	client.MakeApiRequest("POST", "auth/request-password-reset", resource.SignUpTO{
		Email: email,
	})
	if err := client.AssertLastResponseEq(204, nil); err != nil {
		return errs.Wrap("second reset call response mismatch", err)
	}

	// Login with old password should still work
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

	// Check user
	client.MakeApiRequest("GET", "user-info", nil)
	if err := client.AssertLastResponseEq(200, resource.UserInfoTO{Roles: []dm.UserRole{dm.UserRoleUser}, EmailVerified: testUser.EmailVerified}); err != nil {
		return errs.Wrap("auth role response mismatch", err)
	}

	// Logout
	client.MakeApiRequest("POST", "auth/logout", resource.LogoutTO{})
	if err := client.AssertLastResponseEq(204, nil); err != nil {
		return errs.Wrap("logout response mismatch", err)
	}

	// Grab token from email
	token := ""
	for i := 0; token == "" && i < 10; i++ {
		emails := helper.GetSentEmails(testUser, email, "Password reset")
		if len(emails) > 1 {
			return errs.Error("too many password reset emails found")
		}
		if len(emails) == 1 {
			token = strings.TrimSpace(strings.Split(strings.Split(emails[0].Body, "password-reset?token=")[1], "\n")[0])
		}

		if token == "" {
			time.Sleep(500 * time.Millisecond)
		}
	}

	if token == "" {
		return errs.Error("token not found")
	}

	// Set new password with wrong token
	client.MakeApiRequest("POST", "auth/reset-password", resource.ResetPasswordTO{
		Token:       "not-correct",
		NewPassword: newPassword,
	})
	if err := client.AssertLastResponseEq(200, resource.ResetPasswordResponseTO{Status: resource.ResetPasswordResponseInvalid}); err != nil {
		return errs.Wrap("wrong token reset password response mismatch", err)
	}

	// Login with new password should not yet work
	client.MakeApiRequest("POST", "auth/login", resource.LoginTO{
		Email:    email,
		Password: string(newPassword),
	})
	if err := client.AssertLastResponseEq(200, resource.LoginResponseTO{Status: resource.LoginResponseInvalidCredentials}); err != nil {
		return errs.Wrap("login response mismatch", err)
	}

	// Set new password with right token
	client.MakeApiRequest("POST", "auth/reset-password", resource.ResetPasswordTO{
		Email:       email,
		Token:       token,
		NewPassword: newPassword,
	})
	if err := client.AssertLastResponseEq(200, resource.ResetPasswordResponseTO{Status: resource.ResetPasswordResponseSuccess}); err != nil {
		return errs.Wrap("right token reset password response mismatch", err)
	}

	// Login with old password should no longer work
	client.MakeApiRequest("POST", "auth/login", resource.LoginTO{
		Email:    email,
		Password: password,
	})
	if err := client.AssertLastResponseEq(200, resource.LoginResponseTO{Status: resource.LoginResponseInvalidCredentials}); err != nil {
		return errs.Wrap("login response mismatch", err)
	}

	// Login with new password should work now
	client.MakeApiRequest("POST", "auth/login", resource.LoginTO{
		Email:    email,
		Password: string(newPassword),
	})
	if err := client.AssertLastResponseEq(200, resource.LoginResponseTO{Status: resource.LoginResponseLoggedIn}); err != nil {
		return errs.Wrap("login response with new password mismatch", err)
	}
	if !client.HasSessionCookie() {
		return errs.Error("session cookie not found")
	}

	// Check user
	client.MakeApiRequest("GET", "user-info", nil)
	if err := client.AssertLastResponseEq(200, resource.UserInfoTO{Roles: []dm.UserRole{dm.UserRoleUser}, EmailVerified: testUser.EmailVerified}); err != nil {
		return errs.Wrap("auth role response mismatch", err)
	}

	testUser.Password = string(newPassword)
	return nil
}
