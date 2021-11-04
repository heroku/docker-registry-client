package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type TokenTransport struct {
	Transport http.RoundTripper
	Username  string
	Password  string
}

// CannotReplayRequestBody describes the HTTP error, when the server responded with a WWW-Authenticate header,
// but the body of the request (POST, PUT, PATCH) couldn't be replayed
type CannotReplayRequestBody struct {
	Err error
}

func (e CannotReplayRequestBody) Error() string {
	msg := "HTTP server responded with a WWW-Authenticate header, but the body of the request cannot be replayed"
	if e.Err != nil {
		msg += ": " + e.Err.Error()
	}
	return msg
}

var _ error = CannotReplayRequestBody{}

func (t *TokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.Transport.RoundTrip(req)
	if err != nil {
		return resp, err
	}
	if authService := isTokenDemand(resp); authService != nil {
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
	if req.Body != nil {
		// reinitialize Body reader if necessary
		if req.GetBody == nil {
			return nil, &CannotReplayRequestBody{fmt.Errorf("missing GetBody parameter")}
		}
		var err error
		req.Body, err = req.GetBody()
		if err != nil {
			return nil, &CannotReplayRequestBody{Err: err}
		}
	}
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
