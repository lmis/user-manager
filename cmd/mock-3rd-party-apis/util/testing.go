package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"user-manager/cmd/mock-3rd-party-apis/config"
	domain_model "user-manager/domain-model"
	"user-manager/util"
)

type TestUser struct {
	Email         string
	EmailVerified bool
	Password      []byte
	Language      domain_model.UserLanguage
}

type FunctionalTest struct {
	Description string
	Test        func(*config.Config, Emails, *TestUser) error
}

type RequestClient struct {
	config       *config.Config
	cookies      map[string]*http.Cookie
	headers      map[string]string
	lastResponse *http.Response
}

func NewRequestClient(config *config.Config) *RequestClient {
	client := &RequestClient{
		config:  config,
		cookies: map[string]*http.Cookie{},
		headers: map[string]string{},
	}
	client.SetCsrfTokens("abcdef", "abcdef")
	return client
}

func (r *RequestClient) SetCsrfTokens(header string, cookie string) {
	r.headers["X-CSRF-Token"] = header
	r.cookies["CSRF-Token"] = &http.Cookie{
		Name:  "CSRF-Token",
		Value: cookie,
	}
}

func (r *RequestClient) MakeApiRequest(method string, subpath string, payload interface{}) {
	config := r.config
	req, err := http.NewRequest(method, fmt.Sprintf("%s/api/%s", config.AppUrl, subpath), nil)
	if err != nil {
		panic(util.Wrap("error building request", err))
	}
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			panic(util.Wrap("issue marshalling json", err))
		}
		req.Header.Add("Content-Type", "application/json")
		req.Body = io.NopCloser(bytes.NewReader(b))
	}

	for _, val := range r.cookies {
		req.AddCookie(val)
	}
	for name, val := range r.headers {
		req.Header.Add(name, val)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(util.Wrap("error making request", err))
	}

	for _, cookie := range resp.Cookies() {
		r.cookies[cookie.Name] = cookie
	}

	r.lastResponse = resp
}

func (r *RequestClient) HasSessionCookie() bool {
	if val, ok := r.cookies["LOGIN_TOKEN"]; ok {
		return val.Value != ""
	}
	return false
}

func (r *RequestClient) AssertLastResponseEq(expectedStatusCode int, expectedPayload interface{}) error {
	resp := r.lastResponse
	if err := AssertEq(resp.StatusCode, expectedStatusCode); err != nil {
		return util.Wrap("status code mismatch", err)
	}
	body := readAllClose(resp.Body)
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

func readAllClose(reader io.ReadCloser) []byte {
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		panic(util.Wrap("issue reading", err))
	}
	if err = reader.Close(); err != nil {
		panic(util.Wrap("issue closing", err))
	}

	return body
}
