package functional_tests

import (
	"strings"
	"time"
	"user-manager/cmd/app/resource"
	dm "user-manager/domain-model"
	"user-manager/functional-tests/helper"
	"user-manager/util/errs"
)

func TestSignUp(testUser *helper.TestUser) error {
	email := testUser.Email
	password := testUser.Password
	client := helper.NewRequestClient(testUser)

	// Check user before sign-up
	client.MakeApiRequest("GET", "user-info", nil)
	if err := client.AssertLastResponseEq(200, resource.UserInfoTO{Roles: nil, EmailVerified: false}); err != nil {
		return errs.Wrap("response mismatch", err)
	}

	// Sign-up
	client.MakeApiRequest("POST", "auth/sign-up", resource.SignUpTO{
		UserName: "test-user",
		Email:    email,
		Password: []byte(password),
	})
	if err := client.AssertLastResponseEq(204, nil); err != nil {
		return errs.Wrap("signup response mismatch", err)
	}

	// Login
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
	if err := client.AssertLastResponseEq(200, resource.UserInfoTO{Roles: []dm.UserRole{dm.UserRoleUser}, EmailVerified: false}); err != nil {
		return errs.Wrap("user response mismatch", err)
	}

	// Confirm with invalid token
	client.MakeApiRequest("POST", "user/confirm-email", resource.EmailConfirmationTO{
		Token: "invalid",
	})
	if err := client.AssertLastResponseEq(200, resource.EmailConfirmationResponseTO{Status: resource.EmailConfirmationResponseInvalidToken}); err != nil {
		return errs.Wrap("confirm email with invalid token response mismatch", err)
	}

	// Grab token from email
	token := ""
	for i := 0; token == "" && i < 10; i++ {
		emails := helper.GetSentEmails(testUser, email, "Email verification")
		if len(emails) > 1 {
			return errs.Error("too many email verification emails found")
		}
		if len(emails) == 1 {
			token = strings.TrimSpace(strings.Split(strings.Split(emails[0].Body, "email-verification?token=")[1], "\n")[0])
		}
		if token == "" {
			time.Sleep(500 * time.Millisecond)
		}
	}

	if token == "" {
		return errs.Error("token not found")
	}

	// Confirm with token
	client.MakeApiRequest("POST", "user/confirm-email", resource.EmailConfirmationTO{Token: token})
	if err := client.AssertLastResponseEq(200, resource.EmailConfirmationResponseTO{Status: resource.EmailConfirmationResponseNewlyConfirmed}); err != nil {
		return errs.Wrap("confirm email response mismatch", err)
	}

	// Confirm again
	client.MakeApiRequest("POST", "user/confirm-email", resource.EmailConfirmationTO{
		Token: token,
	})
	if err := client.AssertLastResponseEq(200, resource.EmailConfirmationResponseTO{Status: resource.EmailConfirmationResponseAlreadyConfirmed}); err != nil {
		return errs.Wrap("second confirm email response mismatch", err)
	}

	// Check user
	client.MakeApiRequest("GET", "user-info", nil)
	if err := client.AssertLastResponseEq(200, resource.UserInfoTO{Roles: []dm.UserRole{dm.UserRoleUser}, EmailVerified: true}); err != nil {
		return errs.Wrap("user after confirmation response mismatch", err)
	}
	testUser.EmailVerified = true

	// Logout
	client.MakeApiRequest("POST", "auth/logout", resource.LogoutTO{})
	if err := client.AssertLastResponseEq(204, nil); err != nil {
		return errs.Wrap("logout response mismatch", err)
	}

	// Check user
	client.MakeApiRequest("GET", "user-info", nil)
	if err := client.AssertLastResponseEq(200, resource.UserInfoTO{Roles: nil}); err != nil {
		return errs.Wrap("user after logout response mismatch", err)
	}

	// Signup again with same user
	client.MakeApiRequest("POST", "auth/sign-up", resource.SignUpTO{
		UserName: "same-email-different-user",
		Email:    email,
		Password: []byte("another-bad-password"),
	})
	if err := client.AssertLastResponseEq(204, nil); err != nil {
		return errs.Wrap("signup response mismatch", err)
	}

	// Grab email
	receivedNotificationEmail := false
	for i := 0; !receivedNotificationEmail && i < 10; i++ {
		emails := helper.GetSentEmails(testUser, email, "Sign up attempted")
		if len(emails) > 1 {
			return errs.Error("too many sign up attempted emails found")
		}
		if len(emails) == 1 {
			receivedNotificationEmail = true
			break
		}

		if !receivedNotificationEmail {
			time.Sleep(500 * time.Millisecond)
		}
	}

	if !receivedNotificationEmail {
		return errs.Error("notification email not received")
	}

	return nil
}
