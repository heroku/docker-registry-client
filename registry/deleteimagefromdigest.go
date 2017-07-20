package registry

import (
	"net/http"
)

func (registry *Registry) DeleteImageFromDigest(repository string, digest string) error {
	url := registry.url("/v2/%s/manifests/%s", repository, digest)
	registry.Logf("registry.image.delete url=%s repository=%s reference=%s", url, repository, digest)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	resp, err := registry.Client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}

	return nil
}