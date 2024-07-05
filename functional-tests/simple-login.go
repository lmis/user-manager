package functional_tests

import (
	"user-manager/cmd/app/resource"
	dm "user-manager/domain-model"
	"user-manager/functional-tests/helper"
	"user-manager/util/errs"
)

func TestSimpleLogin(testUser *helper.TestUser) error {
	email := testUser.Email
	password := testUser.Password
	client := helper.NewRequestClient(testUser)

	// Login with wrong email
	client.MakeApiRequest("POST", "auth/login", resource.LoginTO{
		Email:    "another-email",
		Password: password,
	})
	if err := client.AssertLastResponseEq(200, resource.LoginResponseTO{Status: resource.LoginResponseInvalidCredentials}); err != nil {
		return errs.Wrap("login with wrong email response mismatch", err)
	}
	if client.HasSessionCookie() {
		return errs.Error("unexpected session cookie returned despite wrong email")
	}

	// Login with wrong password
	client.MakeApiRequest("POST", "auth/login", resource.LoginTO{
		Email:    email,
		Password: "not-the-password",
	})
	if err := client.AssertLastResponseEq(200, resource.LoginResponseTO{Status: resource.LoginResponseInvalidCredentials}); err != nil {
		return errs.Wrap("login with wrong password response mismatch", err)
	}

	if client.HasSessionCookie() {
		return errs.Error("unexpected session cookie returned despite wrong password")
	}

	// Get user info
	client.MakeApiRequest("GET", "user-info", nil)
	if err := client.AssertLastResponseEq(200, resource.UserInfoTO{Roles: nil, EmailVerified: false}); err != nil {
		return errs.Wrap("get user info response mismatch", err)
	}

	// Login with correct info
	client.MakeApiRequest("POST", "auth/login", resource.LoginTO{
		Email:    email,
		Password: password,
	})
	if err := client.AssertLastResponseEq(200, resource.LoginResponseTO{Status: resource.LoginResponseLoggedIn}); err != nil {
		return errs.Wrap("login with correct info response mismatch", err)
	}
	if !client.HasSessionCookie() {
		return errs.Error("expected session cookie returned, got none")
	}

	// Get user info
	client.MakeApiRequest("GET", "user-info", nil)
	if err := client.AssertLastResponseEq(200, resource.UserInfoTO{Roles: []dm.UserRole{dm.UserRoleUser}, EmailVerified: testUser.EmailVerified}); err != nil {
		return errs.Wrap("second get user info response mismatch", err)
	}
	return nil
}
