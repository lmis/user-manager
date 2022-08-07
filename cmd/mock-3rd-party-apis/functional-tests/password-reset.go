package functional_tests

import (
	"net/http"
	"strings"
	"time"
	api_endpoint "user-manager/cmd/app/endpoints"
	auth_endpoint "user-manager/cmd/app/endpoints/auth"
	"user-manager/cmd/mock-3rd-party-apis/config"
	mock_util "user-manager/cmd/mock-3rd-party-apis/util"
	"user-manager/db/generated/models"
	"user-manager/util"
)

func TestPasswordReset(config *config.Config, emails mock_util.Emails, testUser *mock_util.TestUser) error {
	email := testUser.Email
	password := testUser.Password

	// Trigger reset for non-existant email
	resp, err := mock_util.MakeApiRequest("POST", config, "auth/request-password-reset", auth_endpoint.SignUpTO{
		Email: "does-not-exist",
	}, nil)
	if err != nil {
		return util.Wrap("issue making first reset call", err)
	}
	if err = mock_util.AssertResponseEq(204, nil, resp); err != nil {
		return util.Wrap("first reset call response mismatch", err)
	}

	// Trigger reset
	resp, err = mock_util.MakeApiRequest("POST", config, "auth/request-password-reset", auth_endpoint.SignUpTO{
		Email: email,
	}, nil)
	if err != nil {
		return util.Wrap("issue making second reset call", err)
	}
	if err = mock_util.AssertResponseEq(204, nil, resp); err != nil {
		return util.Wrap("second reset call response mismatch", err)
	}

	// Login with old password should still work
	resp, err = mock_util.MakeApiRequest("POST", config, "auth/login", auth_endpoint.LoginTO{
		Email:    email,
		Password: password,
	}, nil)
	if err != nil {
		return util.Wrap("error making login request", err)
	}
	if err = mock_util.AssertResponseEq(200, auth_endpoint.LoginResponseTO{Status: auth_endpoint.LoggedIn}, resp); err != nil {
		return util.Wrap("login response mismatch", err)
	}
	var sessionCookie *http.Cookie
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "LOGIN_TOKEN" {
			sessionCookie = cookie
		}
	}

	if sessionCookie == nil {
		return util.Error("session cookie not found")
	}

	// Check user
	resp, err = mock_util.MakeApiRequest("GET", config, "user", nil, sessionCookie)
	if err != nil {
		return util.Wrap("error making user request", err)
	}
	if err = mock_util.AssertResponseEq(200, api_endpoint.UserTO{Roles: []models.UserRole{"USER"}, Language: "DE"}, resp); err != nil {
		return util.Wrap("auth role response mismatch", err)
	}

	// Logout
	resp, err = mock_util.MakeApiRequest("POST", config, "auth/logout", auth_endpoint.LogoutTO{}, sessionCookie)
	if err != nil {
		return util.Wrap("error making logout call", err)
	}
	if err = mock_util.AssertResponseEq(204, nil, resp); err != nil {
		return util.Wrap("logout response mismatch", err)
	}

	// Grab token from email
	token := ""
	for i := 0; token == "" && i < 10; i++ {
		for _, e := range emails[email] {
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
	newPassword := []byte("hunter3")
	resp, err = mock_util.MakeApiRequest("POST", config, "user/confirm-email", auth_endpoint.ResetPasswordTO{
		Token:       "not-correct",
		NewPassword: newPassword,
	}, nil)
	if err != nil {
		return util.Wrap("error making first reset password call", err)
	}
	if err = mock_util.AssertResponseEq(200, auth_endpoint.ResetPasswordResponseTO{Status: auth_endpoint.InvalidToken}, resp); err != nil {
		return util.Wrap("first reset password response mismatch", err)
	}

	// Login with new password should not yet work
	resp, err = mock_util.MakeApiRequest("POST", config, "auth/login", auth_endpoint.LoginTO{
		Email:    email,
		Password: newPassword,
	}, nil)
	if err != nil {
		return util.Wrap("error making login request", err)
	}
	if err = mock_util.AssertResponseEq(200, auth_endpoint.LoginResponseTO{Status: auth_endpoint.InvalidCredentials}, resp); err != nil {
		return util.Wrap("login response mismatch", err)
	}

	// Set new password with right token
	resp, err = mock_util.MakeApiRequest("POST", config, "user/confirm-email", auth_endpoint.ResetPasswordTO{
		Token:       token,
		NewPassword: newPassword,
	}, nil)
	if err != nil {
		return util.Wrap("error making second reset password call", err)
	}
	if err = mock_util.AssertResponseEq(200, auth_endpoint.ResetPasswordResponseTO{Status: auth_endpoint.Success}, resp); err != nil {
		return util.Wrap("second reset password response mismatch", err)
	}

	// Login with old password should no longer work
	resp, err = mock_util.MakeApiRequest("POST", config, "auth/login", auth_endpoint.LoginTO{
		Email:    email,
		Password: password,
	}, nil)
	if err != nil {
		return util.Wrap("error making login request", err)
	}
	if err = mock_util.AssertResponseEq(200, auth_endpoint.LoginResponseTO{Status: auth_endpoint.InvalidCredentials}, resp); err != nil {
		return util.Wrap("login response mismatch", err)
	}

	// Login with new password should still work
	resp, err = mock_util.MakeApiRequest("POST", config, "auth/login", auth_endpoint.LoginTO{
		Email:    email,
		Password: []byte("hunter3"),
	}, nil)
	if err != nil {
		return util.Wrap("error making login request with new password", err)
	}
	if err = mock_util.AssertResponseEq(200, auth_endpoint.LoginResponseTO{Status: auth_endpoint.LoggedIn}, resp); err != nil {
		return util.Wrap("login response with new password mismatch", err)
	}
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "LOGIN_TOKEN" {
			sessionCookie = cookie
		}
	}

	if sessionCookie == nil {
		return util.Error("session cookie not found")
	}

	// Check user
	resp, err = mock_util.MakeApiRequest("GET", config, "user", nil, sessionCookie)
	if err != nil {
		return util.Wrap("error making user request", err)
	}
	if err = mock_util.AssertResponseEq(200, api_endpoint.UserTO{Roles: []models.UserRole{"USER"}, Language: "DE"}, resp); err != nil {
		return util.Wrap("auth role response mismatch", err)
	}

	testUser.Password = newPassword
	return nil
}
