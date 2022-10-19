package functional_tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"user-manager/cmd/app/resource"
	"user-manager/cmd/mock-3rd-party-apis/config"
	mock_util "user-manager/cmd/mock-3rd-party-apis/util"
	"user-manager/db/generated/models"
	domain_model "user-manager/domain-model"
	"user-manager/util"
)

func TestCallWithMismatchingCsrfTokens(config *config.Config, emails mock_util.Emails, testUser *mock_util.TestUser) error {
	email := testUser.Email
	password := testUser.Password

	// Check user
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/%s", config.AppUrl, "user-info"), nil)
	if err != nil {
		return util.Wrap("error building user request", err)
	}

	req.Header.Add("X-CSRF-Token", "abcdef")
	req.AddCookie(&http.Cookie{
		Name:  "CSRF-Token",
		Value: "other",
	})
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return util.Wrap("error making user request", err)
	}
	if err = mock_util.AssertResponseEq(400, nil, resp); err != nil {
		return util.Wrap("auth user response mismatch", err)
	}

	// Login (with matching CSRF tokens)
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

	req, err = http.NewRequest("POST", fmt.Sprintf("%s/api/%s", config.AppUrl, "user/settings/language"), nil)
	if err != nil {
		return util.Wrap("error building language request", err)
	}

	b, err := json.Marshal(&resource.LanguageTO{Language: domain_model.UserLanguage(models.UserLanguageDE)})
	if err != nil {
		return util.Wrap("issue marshalling language json", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Body = io.NopCloser(bytes.NewReader(b))

	req.Header.Add("X-CSRF-Token", "abcdef")
	req.AddCookie(&http.Cookie{
		Name:  "CSRF-Token",
		Value: "other",
	})
	req.AddCookie(sessionCookie)

	resp, err = http.DefaultClient.Do(req)

	if err != nil {
		return util.Wrap("error making language request", err)
	}
	if err = mock_util.AssertResponseEq(400, nil, resp); err != nil {
		return util.Wrap("language response mismatch", err)
	}
	return nil
}
