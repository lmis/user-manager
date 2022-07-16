package flowtests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	"user-manager/cmd/app/endpoints"
	authendpoints "user-manager/cmd/app/endpoints/auth"
	userendpoints "user-manager/cmd/app/endpoints/user"
	"user-manager/cmd/mock-3rd-party-apis/config"
	emailapi "user-manager/third-party-models/email-api"
	"user-manager/util"
)

func addCsrfHeaders(req *http.Request) {
	req.Header.Add("X-CSRF-Token", "abcdef")
	req.AddCookie(&http.Cookie{
		Name:  "CSRF-Token",
		Value: "abcdef",
	})
}

func makeApiRequest(method string, config *config.Config, subpath string, payload interface{}, sessionCookie *http.Cookie) (*http.Response, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("%s/api/%s", config.AppUrl, subpath), nil)
	if err != nil {
		return nil, util.Wrap("error building request", err)
	}
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, util.Wrap("issue marshalling json", err)
		}
		req.Header.Add("Content-Type", "application/json")
		req.Body = io.NopCloser(bytes.NewReader(b))
	}

	addCsrfHeaders(req)
	if sessionCookie != nil {
		req.AddCookie(sessionCookie)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, util.Wrap("error making signup request", err)
	}
	return resp, nil

}

func TestRoleBeforeSignup(config *config.Config) error {
	resp, err := makeApiRequest("GET", config, "role", nil, nil)
	if err != nil {
		return util.Wrap("error making role request", err)
	}
	if err = assertResponseEq(200, endpoints.AuthRoleTO{Role: "", EmailVerified: false}, resp); err != nil {
		return util.Wrap("response mismatch", err)
	}
	return nil
}

func TestSignUp(config *config.Config, emails map[string][]emailapi.EmailTO) error {
	email := "test-user-1@example.com"
	password := []byte("hunter12")
	// Signup
	resp, err := makeApiRequest("POST", config, "sign-up", endpoints.SignUpTO{
		UserName: "test-user",
		Language: "DE",
		Email:    email,
		Password: password,
	}, nil)
	if err = assertResponseEq(200, nil, resp); err != nil {
		return util.Wrap("signup response mismatch", err)
	}

	// Login
	resp, err = makeApiRequest("POST", config, "auth/login", authendpoints.CredentialsTO{
		Email:    email,
		Password: password,
	}, nil)
	if err != nil {
		return util.Wrap("error making login request", err)
	}
	if err = assertResponseEq(200, authendpoints.LoginResponseTO{LoggedIn: true}, resp); err != nil {
		return util.Wrap("login response mismatch", err)
	}
	var sessionCookie *http.Cookie
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "SESSION_ID" {
			sessionCookie = cookie
		}
	}

	if sessionCookie == nil {
		return util.Error("session cookie not found")
	}

	// Check role
	resp, err = makeApiRequest("GET", config, "role", authendpoints.CredentialsTO{
		Email:    email,
		Password: password,
	}, sessionCookie)
	if err != nil {
		return util.Wrap("error making auth role request", err)
	}
	if err = assertResponseEq(200, endpoints.AuthRoleTO{Role: "USER"}, resp); err != nil {
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
	resp, err = makeApiRequest("POST", config, "user/confirm-email", userendpoints.EmailConfirmationTO{
		Token: token,
	}, sessionCookie)
	if err != nil {
		return util.Wrap("error making confirm email call", err)
	}
	if err = assertResponseEq(200, userendpoints.EmailConfirmationResponseTO{Status: "newly-confirmed"}, resp); err != nil {
		return util.Wrap("confirm email response mismatch", err)
	}

	// Confirm again
	resp, err = makeApiRequest("POST", config, "user/confirm-email", userendpoints.EmailConfirmationTO{
		Token: token,
	}, sessionCookie)

	if err != nil {
		return util.Wrap("error making second confirm email call", err)
	}
	if err = assertResponseEq(200, userendpoints.EmailConfirmationResponseTO{Status: "already-confirmed"}, resp); err != nil {
		return util.Wrap("second confirm email response mismatch", err)
	}

	// Check role
	resp, err = makeApiRequest("GET", config, "role", authendpoints.CredentialsTO{
		Email:    email,
		Password: password,
	}, sessionCookie)
	if err != nil {
		return util.Wrap("error making auth role after confirmation request", err)
	}
	if err = assertResponseEq(200, endpoints.AuthRoleTO{Role: "USER", EmailVerified: true}, resp); err != nil {
		return util.Wrap("auth role after confirmation response mismatch", err)
	}

	// Logout
	resp, err = makeApiRequest("POST", config, "auth/logout", nil, sessionCookie)
	if err != nil {
		return util.Wrap("error making logout call", err)
	}
	if err = assertResponseEq(200, nil, resp); err != nil {
		return util.Wrap("logout response mismatch", err)
	}

	// Check role
	resp, err = makeApiRequest("GET", config, "role", authendpoints.CredentialsTO{
		Email:    email,
		Password: password,
	}, sessionCookie)
	if err != nil {
		return util.Wrap("error making auth role after logout request", err)
	}
	if err = assertResponseEq(200, endpoints.AuthRoleTO{Role: ""}, resp); err != nil {
		return util.Wrap("auth role after logout response mismatch", err)
	}

	return nil
}

func TestCsrf(config *config.Config) error {
	// TODO
	return nil
}
func TestBadPassword(config *config.Config) error {
	// TODO: Sign up, login with wrong password
	return nil
}
func Test2FA(config *config.Config) error {
	// TODO
	return nil
}
func Test2FAThrottling(config *config.Config) error {
	// TODO
	return nil
}

func TestRetriggerConfirmationEmail(config *config.Config) error {
	// TODO
	return nil
}

func TestChangeEmail() {
	// TODO
}

func TestchangePassword() {
	// TODO
}

func assertResponseEq(expectedStatusCode int, expectedPayload interface{}, resp *http.Response) error {
	if err := assertEq(resp.StatusCode, expectedStatusCode); err != nil {
		return util.Wrap("status code mismatch", err)
	}
	body, err := readAllClose(resp.Body)
	if err != nil {
		return util.Wrap("issue reading body target", err)
	}
	if expectedPayload != nil {
		payloadAsJson, err := json.Marshal(expectedPayload)
		if err != nil {
			return util.Wrap("issue serializing expected payload", err)
		}
		if err := assertEq(string(body), string(payloadAsJson)); err != nil {
			return util.Wrap("payload mismatch", err)
		}
	} else {
		if err := assertEq(len(body), 0); err != nil {
			return util.Wrap("payload length mismatch", err)
		}

	}
	return nil
}

func assertEq(received interface{}, expected interface{}) error {
	if received != expected {
		return util.Errorf("expected %v got %v", expected, received)
	}
	return nil

}

func readAllClose(reader io.ReadCloser) ([]byte, error) {
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, util.Wrap("issue reading", err)
	}
	if err = reader.Close(); err != nil {
		return nil, util.Wrap("issue closing", err)
	}

	return body, nil
}

func readCloseAllInto(reader io.ReadCloser, target interface{}) error {
	body, err := readAllClose(reader)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(body, target); err != nil {
		return util.Wrap("issue binding", err)
	}
	return nil
}
