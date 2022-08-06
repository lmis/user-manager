package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"user-manager/cmd/mock-3rd-party-apis/config"
	"user-manager/util"
)

type FunctionalTest struct {
	Description string
	Test        func(*config.Config, Emails) error
}

func MakeApiRequest(method string, config *config.Config, subpath string, payload interface{}, sessionCookie *http.Cookie) (*http.Response, error) {
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

func AssertResponseEq(expectedStatusCode int, expectedPayload interface{}, resp *http.Response) error {
	if err := AssertEq(resp.StatusCode, expectedStatusCode); err != nil {
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
		if err := AssertEq(string(body), string(payloadAsJson)); err != nil {
			return util.Wrap("payload mismatch", err)
		}
	} else {
		if err := AssertEq(len(body), 0); err != nil {
			return util.Wrap("payload length mismatch", err)
		}

	}
	return nil
}

func AssertEq(received interface{}, expected interface{}) error {
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
func addCsrfHeaders(req *http.Request) {
	req.Header.Add("X-CSRF-Token", "abcdef")
	req.AddCookie(&http.Cookie{
		Name:  "CSRF-Token",
		Value: "abcdef",
	})
}
