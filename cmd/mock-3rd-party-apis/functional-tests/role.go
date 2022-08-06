package functional_tests

import (
	api_endpoint "user-manager/cmd/app/endpoints"
	"user-manager/cmd/mock-3rd-party-apis/config"
	mock_util "user-manager/cmd/mock-3rd-party-apis/util"
	"user-manager/util"
)

func TestRoleBeforeSignup(config *config.Config, _ mock_util.Emails) error {
	resp, err := mock_util.MakeApiRequest("GET", config, "role", nil, nil)
	if err != nil {
		return util.Wrap("error making role request", err)
	}
	if err = mock_util.AssertResponseEq(200, api_endpoint.AuthRoleTO{Roles: nil, EmailVerified: false}, resp); err != nil {
		return util.Wrap("response mismatch", err)
	}
	return nil
}
