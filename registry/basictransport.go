package registry

import (
	"net/http"
)

type BasicTransport struct {
	Transport http.RoundTripper
	Username  string
	Password  string
}

func (t *BasicTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.Username != "" || t.Password != "" {
		req.SetBasicAuth(t.Username, t.Password)
	}
	resp, err := t.Transport.RoundTrip(req)
	return resp, err
}
