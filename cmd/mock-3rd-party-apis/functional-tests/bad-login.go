package functional_tests

import (
	"user-manager/cmd/app/resource"
	"user-manager/cmd/mock-3rd-party-apis/config"
	mock_util "user-manager/cmd/mock-3rd-party-apis/util"
	"user-manager/util"
)

func TestBadLogin(config *config.Config, emails mock_util.Emails, testUser *mock_util.TestUser) error {
	email := testUser.Email
	password := testUser.Password

	// Login with wrong email
	resp, err := mock_util.MakeApiRequest("POST", config, "auth/login", resource.LoginTO{
		Email:    "another-email",
		Password: password,
	}, nil)
	if err != nil {
		return util.Wrap("error making login with wrong email request", err)
	}
	if err = mock_util.AssertResponseEq(200, resource.LoginResponseTO{Status: resource.LOGIN_RESPONSE_INVALID_CREDENTIALS}, resp); err != nil {
		return util.Wrap("login with wrong email response mismatch", err)
	}
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "LOGIN_TOKEN" {
			return util.Error("unexpected session cookie returned despite wrong email")
		}
	}

	// Login with wrong password
	resp, err = mock_util.MakeApiRequest("POST", config, "auth/login", resource.LoginTO{
		Email:    email,
		Password: []byte("not-the-password"),
	}, nil)
	if err != nil {
		return util.Wrap("error making login with wrong password request", err)
	}
	if err = mock_util.AssertResponseEq(200, resource.LoginResponseTO{Status: resource.LOGIN_RESPONSE_INVALID_CREDENTIALS}, resp); err != nil {
		return util.Wrap("login with wrong password response mismatch", err)
	}
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "LOGIN_TOKEN" {
			return util.Error("unexpected session cookie returned dspite wrong password")
		}
	}

	return nil
}
