package functional_tests

import (
	"user-manager/cmd/app/resource"
	"user-manager/cmd/mock-3rd-party-apis/config"
	"user-manager/cmd/mock-3rd-party-apis/util"
	dm "user-manager/domain-model"
	"user-manager/util/errs"
)

func TestChangePassword(config *config.Config, _ util.Emails, testUser *util.TestUser) error {
	email := testUser.Email
	password := testUser.Password
	newPassword := []byte("changed-password-via-settings")
	client := util.NewRequestClient(config)

	// Login with correct info
	client.MakeApiRequest("POST", "auth/login", resource.LoginTO{
		Email:    email,
		Password: password,
	})

	// Attempt to change without sudo mode
	client.MakeApiRequest("POST", "user/settings/sensitive-settings/change-password", resource.ChangePasswordTO{
		OldPassword: password,
		NewPassword: newPassword,
	})
	if err := client.AssertLastResponseEq(403, nil); err != nil {
		return errs.Wrap("change without sudo mode response mismatch", err)
	}

	// Sudo login
	client.MakeApiRequest("POST", "auth/login", resource.LoginTO{
		Email:    email,
		Password: password,
		Sudo:     true,
	})

	// Attempt to change with wrong old password
	client.MakeApiRequest("POST", "user/settings/sensitive-settings/change-password", resource.ChangePasswordTO{
		OldPassword: []byte("not the password"),
		NewPassword: newPassword,
	})
	if err := client.AssertLastResponseEq(400, nil); err != nil {
		return errs.Wrap("change with wrong old password response mismatch", err)
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

	// Change with correct password
	client.MakeApiRequest("POST", "user/settings/sensitive-settings/change-password", resource.ChangePasswordTO{
		OldPassword: testUser.Password,
		NewPassword: newPassword,
	})
	if err := client.AssertLastResponseEq(204, nil); err != nil {
		return errs.Wrap("change with correct password response mismatch", err)
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
		Password: newPassword,
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

	testUser.Password = newPassword
	return nil
}
