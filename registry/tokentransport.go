package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type TokenTransport struct {
	Transport http.RoundTripper
	Username  string
	Password  string
	tokens    *TokenPool
}

func NewTokenTransport(transport http.RoundTripper, username string, password string) *TokenTransport {
	return &TokenTransport{
		Transport: transport,
		Username:  username,
		Password:  password,
		tokens:    NewTokenPool(),
	}
}

var scopeReg = regexp.MustCompile(`^/v2/([A-Za-z0-9/-]+)/\b(tags|manifests|blobs)\b`)

func (t *TokenTransport) GetScope(u string) string {
	prefix := "/v2/"
	sc := scopeReg.Find([]byte(u))
	if sc == nil {
		return ""
	} else {
		begin := strings.Index(string(sc), prefix) + len(prefix)
		end := strings.LastIndex(string(sc), "/")
		return string(sc[begin:end])
	}
}

func (t *TokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	sc := t.GetScope(req.URL.EscapedPath())
	token := t.tokens.GetToken(sc)
	if len(token) > 0 {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	resp, err := t.Transport.RoundTrip(req)
	if err != nil {
		return resp, err
	}
	if authService := isTokenDemand(resp); authService != nil {
		resp.Body.Close()
		if req.Body != nil {
			tmp, _ := ioutil.ReadAll(req.Body)
			var reader io.Reader
			reader = bytes.NewReader(tmp)
			req.Body, _ = reader.(io.ReadCloser)
		}

		resp, err = t.authAndRetry(authService, req)
	}
	return resp, err
}

type authToken struct {
	Token string `json:"token"`
}

func (t *TokenTransport) authAndRetry(authService *authService, req *http.Request) (*http.Response, error) {
	token, authResp, err := t.auth(authService)
	if err != nil {
		return authResp, err
	}

	t.tokens.SetToken(authService.Scope, token)
	retryResp, err := t.retry(req, token)
	return retryResp, err
}

func (t *TokenTransport) auth(authService *authService) (string, *http.Response, error) {
	authReq, err := authService.Request(t.Username, t.Password)
	if err != nil {
		return "", nil, err
	}

	client := http.Client{
		Transport: t.Transport,
	}

	response, err := client.Do(authReq)
	if err != nil {
		return "", nil, err
	}

	if response.StatusCode != http.StatusOK {
		return "", response, err
	}
	defer response.Body.Close()

	var authToken authToken
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&authToken)
	if err != nil {
		return "", nil, err
	}

	return authToken.Token, nil, nil
}

func (t *TokenTransport) retry(req *http.Request, token string) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := t.Transport.RoundTrip(req)
	return resp, err
}

type authService struct {
	Realm   string
	Service string
	Scope   string
}

func (authService *authService) Request(username, password string) (*http.Request, error) {
	url, err := url.Parse(authService.Realm)
	if err != nil {
		return nil, err
	}

	q := url.Query()
	q.Set("service", authService.Service)
	if authService.Scope != "" {
		q.Set("scope", authService.Scope)
	}
	url.RawQuery = q.Encode()

	request, err := http.NewRequest("GET", url.String(), nil)

	if username != "" || password != "" {
		request.SetBasicAuth(username, password)
	}

	return request, err
}

func isTokenDemand(resp *http.Response) *authService {
	if resp == nil {
		return nil
	}
	if resp.StatusCode != http.StatusUnauthorized {
		return nil
	}
	return parseOauthHeader(resp)
}

func parseOauthHeader(resp *http.Response) *authService {
	challenges := parseAuthHeader(resp.Header)
	for _, challenge := range challenges {
		if challenge.Scheme == "bearer" {
			return &authService{
				Realm:   challenge.Parameters["realm"],
				Service: challenge.Parameters["service"],
				Scope:   challenge.Parameters["scope"],
			}
		}
	}
	return nil
}
