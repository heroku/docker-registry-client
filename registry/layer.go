package registry

import (
	"io"
	"fmt"
	"net/http"
	"net/url"

	"github.com/docker/distribution"
	digest "github.com/opencontainers/go-digest"
)

func (registry *Registry) DownloadLayer(repository string, digest digest.Digest) (io.ReadCloser, error) {
	url := registry.url("/v2/%s/blobs/%s", repository, digest)
	registry.Logf("registry.layer.download url=%s repository=%s digest=%s", url, repository, digest)

	resp, err := registry.Client.Get(url)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func (registry *Registry) UploadLayer(repository string, digest digest.Digest, content io.Reader) error {
	uploadUrl, err := registry.initiateUpload(repository)
	if err != nil {
		return err
	}
	q := uploadUrl.Query()
	q.Set("digest", digest.String())
	uploadUrl.RawQuery = q.Encode()

	registry.Logf("registry.layer.upload url=%s repository=%s digest=%s", uploadUrl, repository, digest)

	uploadStep1, err := http.NewRequest("PATCH", uploadUrl.String(), content)
	if err != nil {
		return err
	}
	uploadStep1.Header.Set("Content-Type", "application/octet-stream")
	resp1, err := registry.Client.Do(uploadStep1)
	if resp1 != nil {
		defer resp1.Body.Close()
	}
	// TODO: retry upload more than 0 bytes were successfully transferred
	// (HEAD upload UUID, adn check the Range header)
	if err != nil {
		if resp1 == nil {
			return fmt.Errorf("error while uploading layer to %s, digest: %s: %s", repository, digest, err)

		} else {
			return fmt.Errorf("error while uploading layer to %s: %v %v: digest: %s: %s", repository, resp1.StatusCode, resp1.Status, digest, err)
		}
	}
	if resp1.StatusCode != 202 {
		return fmt.Errorf("unexpected PATCH response while uploading layer to %s: %v %v: digest: %s", repository, resp1.StatusCode, resp1.Status, digest)
	}

	uploadStep2, err := http.NewRequest("PUT", uploadUrl.String(), nil)
	if err != nil {
		return err
	}
	uploadStep2.Header.Set("Content-Type", "application/octet-stream")

	_, err = registry.Client.Do(uploadStep2)
	return err
}

func (registry *Registry) HasLayer(repository string, digest digest.Digest) (bool, error) {
	checkUrl := registry.url("/v2/%s/blobs/%s", repository, digest)
	registry.Logf("registry.layer.check url=%s repository=%s digest=%s", checkUrl, repository, digest)

	resp, err := registry.Client.Head(checkUrl)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err == nil {
		return resp.StatusCode == http.StatusOK, nil
	}

	urlErr, ok := err.(*url.Error)
	if !ok {
		return false, err
	}
	httpErr, ok := urlErr.Err.(*HttpStatusError)
	if !ok {
		return false, err
	}
	if httpErr.Response.StatusCode == http.StatusNotFound {
		return false, nil
	}

	return false, err
}

func (registry *Registry) LayerMetadata(repository string, digest digest.Digest) (distribution.Descriptor, error) {
	checkUrl := registry.url("/v2/%s/blobs/%s", repository, digest)
	registry.Logf("registry.layer.check url=%s repository=%s digest=%s", checkUrl, repository, digest)

	resp, err := registry.Client.Head(checkUrl)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return distribution.Descriptor{}, err
	}

	return distribution.Descriptor{
		Digest: digest,
		Size:   resp.ContentLength,
	}, nil
}

func (registry *Registry) initiateUpload(repository string) (*url.URL, error) {
	initiateUrl := registry.url("/v2/%s/blobs/uploads/", repository)
	registry.Logf("registry.layer.initiate-upload url=%s repository=%s", initiateUrl, repository)

	resp, err := registry.Client.Post(initiateUrl, "application/octet-stream", nil)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	location := resp.Header.Get("Location")
	locationUrl, err := url.Parse(location)
	if err != nil {
		return nil, err
	}
	return locationUrl, nil
}
