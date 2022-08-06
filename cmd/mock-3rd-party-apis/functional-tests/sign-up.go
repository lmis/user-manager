package functional_tests

import (
	"net/http"
	"strings"
	"time"
	api_endpoint "user-manager/cmd/app/endpoints"
	auth_endpoint "user-manager/cmd/app/endpoints/auth"
	user_endpoint "user-manager/cmd/app/endpoints/user"
	"user-manager/cmd/mock-3rd-party-apis/config"
	mock_util "user-manager/cmd/mock-3rd-party-apis/util"
	"user-manager/db/generated/models"
	"user-manager/util"
)

func TestSignUp(config *config.Config, emails mock_util.Emails) error {
	email := "test-user-" + util.MakeRandomURLSafeB64(3) + "@example.com"
	password := []byte("hunter12")
	// Signup
	resp, err := mock_util.MakeApiRequest("POST", config, "auth/sign-up", auth_endpoint.SignUpTO{
		UserName: "test-user",
		Language: "DE",
		Email:    email,
		Password: password,
	}, nil)
	if err != nil {
		return util.Wrap("issue making auth/sign-up call", err)
	}
	if err = mock_util.AssertResponseEq(204, nil, resp); err != nil {
		return util.Wrap("signup response mismatch", err)
	}

	// Login
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

	// Check role
	resp, err = mock_util.MakeApiRequest("GET", config, "role", auth_endpoint.LoginTO{
		Email:    email,
		Password: password,
	}, sessionCookie)
	if err != nil {
		return util.Wrap("error making auth role request", err)
	}
	if err = mock_util.AssertResponseEq(200, api_endpoint.AuthRoleTO{Roles: []models.UserRole{"USER"}}, resp); err != nil {
		return util.Wrap("auth role response mismatch", err)
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
		return util.Error("token not found")
	}

	// Confirm with token
	resp, err = mock_util.MakeApiRequest("POST", config, "user/confirm-email", user_endpoint.EmailConfirmationTO{
		Token: token,
	}, sessionCookie)
	if err != nil {
		return util.Wrap("error making confirm email call", err)
	}
	if err = mock_util.AssertResponseEq(200, user_endpoint.EmailConfirmationResponseTO{Status: "newly-confirmed"}, resp); err != nil {
		return util.Wrap("confirm email response mismatch", err)
	}

	// Confirm again
	resp, err = mock_util.MakeApiRequest("POST", config, "user/confirm-email", user_endpoint.EmailConfirmationTO{
		Token: token,
	}, sessionCookie)

	if err != nil {
		return util.Wrap("error making second confirm email call", err)
	}
	if err = mock_util.AssertResponseEq(200, user_endpoint.EmailConfirmationResponseTO{Status: "already-confirmed"}, resp); err != nil {
		return util.Wrap("second confirm email response mismatch", err)
	}

	// Check role
	resp, err = mock_util.MakeApiRequest("GET", config, "role", auth_endpoint.LoginTO{
		Email:    email,
		Password: password,
	}, sessionCookie)
	if err != nil {
		return util.Wrap("error making auth role after confirmation request", err)
	}
	if err = mock_util.AssertResponseEq(200, api_endpoint.AuthRoleTO{Roles: []models.UserRole{"USER"}, EmailVerified: true}, resp); err != nil {
		return util.Wrap("auth role after confirmation response mismatch", err)
	}

	// Logout
	resp, err = mock_util.MakeApiRequest("POST", config, "auth/logout", auth_endpoint.LogoutTO{}, sessionCookie)
	if err != nil {
		return util.Wrap("error making logout call", err)
	}
	if err = mock_util.AssertResponseEq(204, nil, resp); err != nil {
		return util.Wrap("logout response mismatch", err)
	}

	// Check role
	resp, err = mock_util.MakeApiRequest("GET", config, "role", auth_endpoint.LoginTO{
		Email:    email,
		Password: password,
	}, sessionCookie)
	if err != nil {
		return util.Wrap("error making auth role after logout request", err)
	}
	if err = mock_util.AssertResponseEq(200, api_endpoint.AuthRoleTO{Roles: nil}, resp); err != nil {
		return util.Wrap("auth role after logout response mismatch", err)
	}

	return nil
}
