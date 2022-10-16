package functional_tests

import (
	"user-manager/cmd/mock-3rd-party-apis/config"
	mock_util "user-manager/cmd/mock-3rd-party-apis/util"
)

func TestSignUp(config *config.Config, emails mock_util.Emails, testUser *mock_util.TestUser) error {
	panic("todo")
	// email := testUser.Email
	// password := testUser.Password
	// // Signup
	// resp, err := mock_util.MakeApiRequest("POST", config, "auth/sign-up", resource.SignUpTO{
	// 	UserName: "test-user",
	// 	Language: "DE",
	// 	Email:    email,
	// 	Password: password,
	// }, nil)
	// if err != nil {
	// 	return util.Wrap("issue making auth/sign-up call", err)
	// }
	// if err = mock_util.AssertResponseEq(204, nil, resp); err != nil {
	// 	return util.Wrap("signup response mismatch", err)
	// }

	// // Login
	// resp, err = mock_util.MakeApiRequest("POST", config, "auth/login", resource.LoginTO{
	// 	Email:    email,
	// 	Password: password,
	// }, nil)
	// if err != nil {
	// 	return util.Wrap("error making login request", err)
	// }
	// if err = mock_util.AssertResponseEq(200, resource.LoginResponseTO{Status: resource.LoggedIn}, resp); err != nil {
	// 	return util.Wrap("login response mismatch", err)
	// }
	// var sessionCookie *http.Cookie
	// for _, cookie := range resp.Cookies() {
	// 	if cookie.Name == "LOGIN_TOKEN" {
	// 		sessionCookie = cookie
	// 	}
	// }

	// if sessionCookie == nil {
	// 	return util.Error("session cookie not found")
	// }

	// // Check user
	// resp, err = mock_util.MakeApiRequest("GET", config, "user", nil, sessionCookie)
	// if err != nil {
	// 	return util.Wrap("error making user request", err)
	// }
	// if err = mock_util.AssertResponseEq(200, resource.UserTO{Roles: []domain_model.UserRole{"USER"}, Language: "DE"}, resp); err != nil {
	// 	return util.Wrap("user resonse mismatch", err)
	// }

	// // Confirm with invalid token
	// resp, err = mock_util.MakeApiRequest("POST", config, "user/confirm-email", resource.EmailConfirmationTO{
	// 	Token: "invalid",
	// }, sessionCookie)
	// if err != nil {
	// 	return util.Wrap("error making confirm email with invalid token call", err)
	// }
	// if err = mock_util.AssertResponseEq(200, resource.EmailConfirmationResponseTO{Status: resource.InvalidToken}, resp); err != nil {
	// 	return util.Wrap("confirm email with invalid token response mismatch", err)
	// }

	// // Grab token from email
	// token := ""
	// for i := 0; token == "" && i < 10; i++ {
	// 	for _, e := range emails[email] {
	// 		if e.Subject == "Email Bestätigung" {
	// 			token = strings.TrimSpace(strings.Split(strings.Split(e.Body, "email-verification?token=")[1], " ")[0])
	// 		}
	// 	}
	// 	if token == "" {
	// 		time.Sleep(500 * time.Millisecond)
	// 	}
	// }

	// if token == "" {
	// 	return util.Error("token not found")
	// }

	// // Confirm with token
	// resp, err = mock_util.MakeApiRequest("POST", config, "user/confirm-email", resource.EmailConfirmationTO{
	// 	Token: token,
	// }, sessionCookie)
	// if err != nil {
	// 	return util.Wrap("error making confirm email call", err)
	// }
	// if err = mock_util.AssertResponseEq(200, resource.EmailConfirmationResponseTO{Status: "newly-confirmed"}, resp); err != nil {
	// 	return util.Wrap("confirm email response mismatch", err)
	// }

	// // Confirm again
	// resp, err = mock_util.MakeApiRequest("POST", config, "user/confirm-email", resource.EmailConfirmationTO{
	// 	Token: token,
	// }, sessionCookie)

	// if err != nil {
	// 	return util.Wrap("error making second confirm email call", err)
	// }
	// if err = mock_util.AssertResponseEq(200, resource.EmailConfirmationResponseTO{Status: "already-confirmed"}, resp); err != nil {
	// 	return util.Wrap("second confirm email response mismatch", err)
	// }

	// // Check user
	// resp, err = mock_util.MakeApiRequest("GET", config, "user", nil, sessionCookie)
	// if err != nil {
	// 	return util.Wrap("error making user after confirmation request", err)
	// }
	// if err = mock_util.AssertResponseEq(200, resource.UserTO{Roles: []domain_model.UserRole{"USER"}, EmailVerified: true, Language: "DE"}, resp); err != nil {
	// 	return util.Wrap("user after confirmation response mismatch", err)
	// }

	// // Logout
	// resp, err = mock_util.MakeApiRequest("POST", config, "auth/logout", resource.LogoutTO{}, sessionCookie)
	// if err != nil {
	// 	return util.Wrap("error making logout call", err)
	// }
	// if err = mock_util.AssertResponseEq(204, nil, resp); err != nil {
	// 	return util.Wrap("logout response mismatch", err)
	// }

	// // Check user
	// resp, err = mock_util.MakeApiRequest("GET", config, "user", nil, sessionCookie)
	// if err != nil {
	// 	return util.Wrap("error making user after logout request", err)
	// }
	// if err = mock_util.AssertResponseEq(200, resource.UserTO{Roles: nil}, resp); err != nil {
	// 	return util.Wrap("user after logout response mismatch", err)
	// }

	// // Signup again with same user
	// resp, err = mock_util.MakeApiRequest("POST", config, "auth/sign-up", resource.SignUpTO{
	// 	UserName: "same-email-different-user",
	// 	Language: "EN",
	// 	Email:    email,
	// 	Password: []byte("another-bad-password"),
	// }, nil)
	// if err != nil {
	// 	return util.Wrap("issue making second auth/sign-up call", err)
	// }
	// if err = mock_util.AssertResponseEq(204, nil, resp); err != nil {
	// 	return util.Wrap("signup response mismatch", err)
	// }

	// // Grab email
	// receivedNotificationEmail := false
	// for i := 0; !receivedNotificationEmail && i < 10; i++ {
	// 	if slices.Any(emails[email], func(email *email_api.EmailTO) bool { return email.Subject == "Anmeldeversuch" }) {
	// 		receivedNotificationEmail = true
	// 	} else {
	// 		time.Sleep(500 * time.Millisecond)
	// 	}
	// }

	// if !receivedNotificationEmail {
	// 	return util.Error("notification email not received")
	// }

	// return nil
}
