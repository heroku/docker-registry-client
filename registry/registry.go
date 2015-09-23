package registry

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

type Registry struct {
	URL      string
	Client   *http.Client
}

func New(registryUrl, username, password string) (*Registry, error) {
	url := strings.TrimSuffix(registryUrl, "/")

	transport := http.DefaultTransport
	transport = &TokenTransport{
		Transport: transport,
		Username:  username,
		Password:  password,
	}
	transport = &BasicTransport{
		Transport: transport,
		Username:  username,
		Password:  password,
	}
	transport = &ErrorTransport{
		Transport: transport,
	}

	registry := &Registry{
		URL:    url,
		Client: &http.Client{
			Transport: transport,
		},
	}

	if err := registry.Ping(); err != nil {
		return nil, err
	}

	return registry, nil
}

func (r *Registry) url(pathTemplate string, args ...interface{}) string {
	pathSuffix := fmt.Sprintf(pathTemplate, args...)
	url := fmt.Sprintf("%s%s", r.URL, pathSuffix)
	return url
}

func (r *Registry) Ping() error {
	url := r.url("/v2/")
	log.Printf("registry.ping url=%s", url)
	resp, err := r.Client.Get(url)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		
	}
	return err
}
