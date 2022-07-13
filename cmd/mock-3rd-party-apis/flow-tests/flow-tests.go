package flowtests

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"user-manager/cmd/app/endpoints"
	"user-manager/cmd/mock-3rd-party-apis/config"
	"user-manager/util"
)

func TestRoleBeforeSignup(client *http.Client, config *config.Config) error {
	roleResponse, err := client.Get(config.AppUrl + "/api/role")
	if err != nil {
		return util.Wrap("status code mismatch", err)
	}
	if err = assertEq(roleResponse.StatusCode, 200); err != nil {
		return util.Wrap("status code mismatch", err)
	}
	var role endpoints.AuthRoleTO
	if err = readCloseAllInto(roleResponse.Body, role); err != nil {
		return util.Wrap("issue reading into RoleTO", err)
	}
	if err = assertEq(role, endpoints.AuthRoleTO{Role: ""}); err != nil {
		return util.Wrap("Auth role mismatch", err)
	}
	return nil
}

func assertEq(received interface{}, expected interface{}) error {
	if received != expected {
		return util.Errorf("expected %v got %v", expected, received)
	}
	return nil
}

func readCloseAllInto(reader io.ReadCloser, target interface{}) error {
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return util.Wrap("issue reading", err)
	}
	if err = reader.Close(); err != nil {
		return util.Wrap("issue closing", err)
	}
	if err = json.Unmarshal(body, target); err != nil {
		return util.Wrap("issue binding", err)
	}
	return nil
}
