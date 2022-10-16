package functional_tests

import (
	"user-manager/cmd/app/resource"
	"user-manager/cmd/mock-3rd-party-apis/config"
	mock_util "user-manager/cmd/mock-3rd-party-apis/util"
	"user-manager/util"
)

func TestUserEndpointBeforeSignup(config *config.Config, _ mock_util.Emails, _ *mock_util.TestUser) error {
	resp, err := mock_util.MakeApiRequest("GET", config, "user", nil, nil)
	if err != nil {
		return util.Wrap("error making user request", err)
	}
	if err = mock_util.AssertResponseEq(200, resource.UserInfoTO{Roles: nil, EmailVerified: false}, resp); err != nil {
		return util.Wrap("response mismatch", err)
	}
	return nil
}
