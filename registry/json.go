package registry

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
)

var (
	ErrNoMorePages = errors.New("No more pages")
)

func (registry *Registry) getJson(url string, response interface{}) error {
	resp, err := registry.Client.Get(url)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(response)
	if err != nil {
		return err
	}

	return nil
}

// getPaginatedJson accepts a string and a pointer, and returns the
// next page URL while updating pointed-to variable with a parsed JSON
// value. When there are no more pages it returns `ErrNoMorePages`.
func (registry *Registry) getPaginatedJson(url string, response interface{}) (string, error) {
	resp, err := registry.Client.Get(url)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return "", err
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(response)
	if err != nil {
		return "", err
	}
	return getNextLink(resp)
}

var linkRE *regexp.Regexp = regexp.MustCompile(`^ *<?([^;>]+)>?(?:; *([a-z]+)="?([^";]*)"?)*$`)

func getNextLink(resp *http.Response) (string, error) {
	for _, link := range resp.Header[http.CanonicalHeaderKey("Link")] {
		parts := linkRE.FindStringSubmatch(link)
		if len(parts) < 4 {
			continue
		}
		// We have a structure like []string{(whole match), URL, key, value, ...}
		for i := 2; i < len(parts); i += 2 {
			if parts[i] == "rel" && parts[i+1] == "next" {
				return parts[1], nil
			}
		}
	}
	return "", ErrNoMorePages
}
