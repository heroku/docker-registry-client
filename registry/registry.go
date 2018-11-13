package registry

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// Registry is the main Docker Registry API client type
type Registry struct {
	URL    string
	Client *http.Client
	Logf   LogfCallback
}

// LogfCallback is the prototype of the custom logging function used by Registry
type LogfCallback func(format string, args ...interface{})

// Quiet discards log messages silently.
func Quiet(format string, args ...interface{}) {
	/* discard logs */
}

// Log passes log messages along to Go's "log" module.
func Log(format string, args ...interface{}) {
	log.Printf(format, args...)
}

// Options stores optional parameters for constructing a new Registry
// See details in the docs of NewCustom()
type Options struct {
	Username         string       `json:"username,omitempty"`
	Password         string       `json:"password,omitempty"`
	Insecure         bool         `json:"insecure,omitempty"`
	Logf             LogfCallback `json:"-"`
	DoInitialPing    bool         `json:"do_initial_ping,omitempty"`
	DisableBasicAuth bool         `json:"disable_basicauth,omitempty"`
}

// NewCustom creates a new Registry with the given URL and optional parameters.
// The interpretation of the optional parameters:
//   Username, Password: credentials for the Docker Registry (default: anonymous access)
//   Insecure: disables TLS certificate verification (default: TLS certificate verification is enabled)
//   Logf: all log messages will be passed to this function (default: registry.Log)
//   DoInitialPing: if true, the registry will be Ping()ed during construction (defualt: false)
//					(note that some registries, e.g. quay.io, don't support anonymous Ping())
//   DisableBasicAuth: disable basicauth authentication (default: basicauth is enabled)
//                     (note that some registries, e.g. older versions of Artifactory, don't play well
//                     with both token and basic auth enabled)
func NewCustom(url string, opts Options) (*Registry, error) {
	url = strings.TrimSuffix(url, "/")
	var transport http.RoundTripper
	if opts.Insecure {
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	} else {
		transport = http.DefaultTransport
	}
	transport = WrapTransport(transport, url, opts)

	logf := opts.Logf
	if logf == nil {
		logf = Log
	}
	registry := &Registry{
		URL: url,
		Client: &http.Client{
			Transport: transport,
		},
		Logf: logf,
	}
	if opts.DoInitialPing {
		if err := registry.Ping(); err != nil {
			return nil, err
		}
	}
	return registry, nil

}

// New creates a new Registry with the given URL and credentials, then Ping()s it
// before returning it to verify that the registry is available.
// Be aware that this will print out log messages for the initial Ping()
// no matter what.
//
// This constructor is left here for backward compatitibiliy,
// use NewCustom() if you need more control over constructor parameters.
func New(url, username, password string) (*Registry, error) {
	return NewCustom(url, Options{
		Username:      username,
		Password:      password,
		Logf:          Log,
		DoInitialPing: true,
	})
}

// NewInsecure creates a new Registry, as with New, but using an http.Transport that disables
// SSL certificate verification.
// Be aware that this will print out log messages for the initial Ping()
// no matter what.
//
// This constructor is left here for backward compatitibiliy,
// use NewCustom() if you need more control over constructor parameters.
func NewInsecure(url, username, password string) (*Registry, error) {
	return NewCustom(url, Options{
		Username:      username,
		Password:      password,
		Insecure:      true,
		Logf:          Log,
		DoInitialPing: true,
	})
}

/*
 * WrapTransport takes an existing http.RoundTripper such as http.DefaultTransport,
 * and builds the transport stack necessary to authenticate to the Docker registry API.
 * This adds in support for OAuth bearer tokens and HTTP Basic auth, and sets up
 * error handling this library relies on.
 */
func WrapTransport(transport http.RoundTripper, url string, opts Options) http.RoundTripper {
	transport = &TokenTransport{
		Transport: transport,
		Username:  opts.Username,
		Password:  opts.Password,
	}
	if !opts.DisableBasicAuth {
		transport = &BasicTransport{
			Transport: transport,
			URL:       url,
			Username:  opts.Username,
			Password:  opts.Password,
		}
	}
	transport = &ErrorTransport{
		Transport: transport,
	}
	return transport
}

func (r *Registry) url(pathTemplate string, args ...interface{}) string {
	pathSuffix := fmt.Sprintf(pathTemplate, args...)
	url := fmt.Sprintf("%s%s", r.URL, pathSuffix)
	return url
}

func (r *Registry) Ping() error {
	url := r.url("/v2/")
	r.Logf("registry.ping url=%s", url)
	resp, err := r.Client.Get(url)
	if resp != nil {
		defer resp.Body.Close()
	}
	return err
}
