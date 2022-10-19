package functional_tests

import (
	"net/http"
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

	// Trigger reset for non-existant email
	resp, err := mock_util.MakeApiRequest("POST", config, "auth/request-password-reset", resource.SignUpTO{
		Email: "does-not-exist",
	}, nil)
	if err != nil {
		return util.Wrap("issue making first reset call", err)
	}
	if err = mock_util.AssertResponseEq(204, nil, resp); err != nil {
		return util.Wrap("first reset call response mismatch", err)
	}

	// Trigger reset
	resp, err = mock_util.MakeApiRequest("POST", config, "auth/request-password-reset", resource.SignUpTO{
		Email: email,
	}, nil)
	if err != nil {
		return util.Wrap("issue making second reset call", err)
	}
	if err = mock_util.AssertResponseEq(204, nil, resp); err != nil {
		return util.Wrap("second reset call response mismatch", err)
	}

	// Login with old password should still work
	resp, err = mock_util.MakeApiRequest("POST", config, "auth/login", resource.LoginTO{
		Email:    email,
		Password: password,
	}, nil)
	if err != nil {
		return util.Wrap("error making login request", err)
	}
	if err = mock_util.AssertResponseEq(200, resource.LoginResponseTO{Status: resource.LOGIN_RESPONSE_LOGGED_IN}, resp); err != nil {
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
	resp, err = mock_util.MakeApiRequest("GET", config, "user-info", nil, sessionCookie)
	if err != nil {
		return util.Wrap("error making user request", err)
	}
	if err = mock_util.AssertResponseEq(200, resource.UserInfoTO{Roles: []domain_model.UserRole{domain_model.USER_ROLE_USER}, Language: "DE"}, resp); err != nil {
		return util.Wrap("auth role response mismatch", err)
	}

	// Logout
	resp, err = mock_util.MakeApiRequest("POST", config, "auth/logout", resource.LogoutTO{}, sessionCookie)
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

	return nil
	// // Set new password with wrong token
	// newPassword := []byte("hunter3")
	// resp, err = mock_util.MakeApiRequest("POST", config, "user/confirm-email", resource.ResetPasswordTO{
	// 	Token:       "not-correct",
	// 	NewPassword: newPassword,
	// }, nil)
	// if err != nil {
	// 	return util.Wrap("error making first reset password call", err)
	// }
	// if err = mock_util.AssertResponseEq(200, resource.ResetPasswordResponseTO{Status: resource.InvalidToken}, resp); err != nil {
	// 	return util.Wrap("first reset password response mismatch", err)
	// }

	// // Login with new password should not yet work
	// resp, err = mock_util.MakeApiRequest("POST", config, "auth/login", resource.LoginTO{
	// 	Email:    email,
	// 	Password: newPassword,
	// }, nil)
	// if err != nil {
	// 	return util.Wrap("error making login request", err)
	// }
	// if err = mock_util.AssertResponseEq(200, resource.LoginResponseTO{Status: resource.InvalidCredentials}, resp); err != nil {
	// 	return util.Wrap("login response mismatch", err)
	// }

	// // Set new password with right token
	// resp, err = mock_util.MakeApiRequest("POST", config, "user/confirm-email", resource.ResetPasswordTO{
	// 	Token:       token,
	// 	NewPassword: newPassword,
	// }, nil)
	// if err != nil {
	// 	return util.Wrap("error making second reset password call", err)
	// }
	// if err = mock_util.AssertResponseEq(200, resource.ResetPasswordResponseTO{Status: resource.RESET_PASSWORD_SUCCESS}, resp); err != nil {
	// 	return util.Wrap("second reset password response mismatch", err)
	// }

	// // Login with old password should no longer work
	// resp, err = mock_util.MakeApiRequest("POST", config, "auth/login", resource.LoginTO{
	// 	Email:    email,
	// 	Password: password,
	// }, nil)
	// if err != nil {
	// 	return util.Wrap("error making login request", err)
	// }
	// if err = mock_util.AssertResponseEq(200, resource.LoginResponseTO{Status: resource.InvalidCredentials}, resp); err != nil {
	// 	return util.Wrap("login response mismatch", err)
	// }

	// // Login with new password should still work
	// resp, err = mock_util.MakeApiRequest("POST", config, "auth/login", resource.LoginTO{
	// 	Email:    email,
	// 	Password: []byte("hunter3"),
	// }, nil)
	// if err != nil {
	// 	return util.Wrap("error making login request with new password", err)
	// }
	// if err = mock_util.AssertResponseEq(200, resource.LoginResponseTO{Status: resource.LoggedIn}, resp); err != nil {
	// 	return util.Wrap("login response with new password mismatch", err)
	// }
	// for _, cookie := range resp.Cookies() {
	// 	if cookie.Name == "LOGIN_TOKEN" {
	// 		sessionCookie = cookie
	// 	}
	// }

	// if sessionCookie == nil {
	// 	return util.Error("session cookie not found")
	// }

	// // Check user
	// resp, err = mock_util.MakeApiRequest("GET", config, "user-info", nil, sessionCookie)
	// if err != nil {
	// 	return util.Wrap("error making user request", err)
	// }
	// if err = mock_util.AssertResponseEq(200, resource.UserInfoTO{Roles: []domain_model.UserRole{domain_model.USER_ROLE_USER}, Language: "DE"}, resp); err != nil {
	// 	return util.Wrap("auth role response mismatch", err)
	// }

	// testUser.Password = newPassword
	// return nil
}
