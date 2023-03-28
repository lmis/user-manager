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

func TestSignUp(config *config.Config, emails util.Emails, testUser *util.TestUser) error {
	email := testUser.Email
	password := testUser.Password
	language := testUser.Language
	client := util.NewRequestClient(config)

	// Check user before sign-up
	client.MakeApiRequest("GET", "user-info", nil)
	if err := client.AssertLastResponseEq(200, resource.UserInfoTO{Roles: nil, EmailVerified: false, Language: ""}); err != nil {
		return errors.Wrap("response mismatch", err)
	}

	// Sign-up
	client.MakeApiRequest("POST", "auth/sign-up", resource.SignUpTO{
		UserName: "test-user",
		Language: string(language),
		Email:    email,
		Password: password,
	})
	if err := client.AssertLastResponseEq(204, nil); err != nil {
		return errors.Wrap("signup response mismatch", err)
	}

	// Login
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
	if err := client.AssertLastResponseEq(200, resource.UserInfoTO{Roles: []dm.UserRole{dm.UserRoleUser}, EmailVerified: false, Language: language}); err != nil {
		return errors.Wrap("user response mismatch", err)
	}

	// Confirm with invalid token
	client.MakeApiRequest("POST", "user/confirm-email", resource.EmailConfirmationTO{
		Token: "invalid",
	})
	if err := client.AssertLastResponseEq(200, resource.EmailConfirmationResponseTO{Status: resource.EmailConfirmationResponseInvalidToken}); err != nil {
		return errors.Wrap("confirm email with invalid token response mismatch", err)
	}

	// Grab token from email
	token := ""
	for i := 0; token == "" && i < 10; i++ {
		for _, e := range emails[email] {
			if e.Subject == "Email Bestätigung" {
				token = strings.TrimSpace(strings.Split(strings.Split(e.Body, "email-verification?token=")[1], " ")[0])
			}
		}
		if token == "" {
			time.Sleep(500 * time.Millisecond)
		}
	}

	if token == "" {
		return errors.Error("token not found")
	}

	// Confirm with token
	client.MakeApiRequest("POST", "user/confirm-email", resource.EmailConfirmationTO{Token: token})
	if err := client.AssertLastResponseEq(200, resource.EmailConfirmationResponseTO{Status: resource.EmailConfirmationResponseNewlyConfirmed}); err != nil {
		return errors.Wrap("confirm email response mismatch", err)
	}

	// Confirm again
	client.MakeApiRequest("POST", "user/confirm-email", resource.EmailConfirmationTO{
		Token: token,
	})
	if err := client.AssertLastResponseEq(200, resource.EmailConfirmationResponseTO{Status: resource.EmailConfirmationResponseAlreadyConfirmed}); err != nil {
		return errors.Wrap("second confirm email response mismatch", err)
	}

	// Check user
	client.MakeApiRequest("GET", "user-info", nil)
	if err := client.AssertLastResponseEq(200, resource.UserInfoTO{Roles: []dm.UserRole{dm.UserRoleUser}, EmailVerified: true, Language: language}); err != nil {
		return errors.Wrap("user after confirmation response mismatch", err)
	}
	testUser.EmailVerified = true

	// Logout
	client.MakeApiRequest("POST", "auth/logout", resource.LogoutTO{})
	if err := client.AssertLastResponseEq(204, nil); err != nil {
		return errors.Wrap("logout response mismatch", err)
	}

	// Check user
	client.MakeApiRequest("GET", "user-info", nil)
	if err := client.AssertLastResponseEq(200, resource.UserInfoTO{Roles: nil}); err != nil {
		return errors.Wrap("user after logout response mismatch", err)
	}

	// Signup again with same user
	client.MakeApiRequest("POST", "auth/sign-up", resource.SignUpTO{
		UserName: "same-email-different-user",
		Language: "EN",
		Email:    email,
		Password: []byte("another-bad-password"),
	})
	if err := client.AssertLastResponseEq(204, nil); err != nil {
		return errors.Wrap("signup response mismatch", err)
	}

	// Grab email
	receivedNotificationEmail := false
	for i := 0; !receivedNotificationEmail && i < 10; i++ {
		// TODO: make independent of language
		for _, email := range emails[email] {
			if email.Subject == "Anmeldeversuch" {
				receivedNotificationEmail = true
				break
			}
		}
		if !receivedNotificationEmail {
			time.Sleep(500 * time.Millisecond)
		}
	}

	if !receivedNotificationEmail {
		return errors.Error("notification email not received")
	}

	return nil
}
