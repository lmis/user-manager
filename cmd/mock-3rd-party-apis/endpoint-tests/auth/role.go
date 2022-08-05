package auth_endpoint_test

import (
	api_endpoint "user-manager/cmd/app/endpoints"
	"user-manager/cmd/mock-3rd-party-apis/config"
	test_util "user-manager/cmd/mock-3rd-party-apis/endpoint-tests"
	"user-manager/util"
)

func TestRoleBeforeSignup(config *config.Config) error {
	resp, err := test_util.MakeApiRequest("GET", config, "role", nil, nil)
	if err != nil {
		return util.Wrap("error making role request", err)
	}
	if err = test_util.AssertResponseEq(200, api_endpoint.AuthRoleTO{Roles: nil, EmailVerified: false}, resp); err != nil {
		return util.Wrap("response mismatch", err)
	}
	return nil
}
