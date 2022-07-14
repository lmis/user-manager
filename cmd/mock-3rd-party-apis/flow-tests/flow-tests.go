package flowtests

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"user-manager/cmd/app/endpoints"
	"user-manager/cmd/mock-3rd-party-apis/config"
	"user-manager/util"
)

func addCsrfHeaders(req *http.Request) {
	req.Header.Add("X-CSRF-Token", "abcdef")
	req.AddCookie(&http.Cookie{
		Name:  "CSRF-Token",
		Value: "abcdef",
	})
}

func addJsonPayload(req *http.Request, payload interface{}) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return util.Wrap("issue marshalling json", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Body = io.NopCloser(bytes.NewReader(b))

	return nil

}

func TestRoleBeforeSignup(config *config.Config) error {
	req, err := http.NewRequest("GET", config.AppUrl+"/api/role", nil)
	addCsrfHeaders(req)
	if err != nil {
		return util.Wrap("error building request", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return util.Wrap("error making request", err)
	}
	if err = assertResponseEq(200, endpoints.AuthRoleTO{Role: ""}, resp); err != nil {
		return util.Wrap("response mismatch", err)
	}
	return nil
}

func TestSignUp(config *config.Config) error {
	req, err := http.NewRequest("POST", config.AppUrl+"/api/sign-up", nil)
	if err != nil {
		return util.Wrap("error building request", err)
	}
	if err := addJsonPayload(req, endpoints.SignUpTO{
		UserName: "test-user",
		Language: "DE",
		Email:    "test-user-1@example.com",
		Password: []byte("hunter12"),
	}); err != nil {
		return util.Wrap("error adding payload to request", err)
	}

	addCsrfHeaders(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return util.Wrap("error making request", err)
	}
	if err = assertResponseEq(200, nil, resp); err != nil {
		return util.Wrap("response mismatch", err)
	}

	return nil
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
