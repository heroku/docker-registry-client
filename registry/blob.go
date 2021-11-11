package registry

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/docker/distribution"
	digest "github.com/opencontainers/go-digest"
	"github.com/sirupsen/logrus"
)

func (registry *Registry) DownloadBlob(repository string, digest digest.Digest) (io.ReadCloser, error) {
	url := registry.url("/v2/%s/blobs/%s", repository, digest)
	registry.Logf("registry.blob.download url=%s repository=%s digest=%s", url, repository, digest)

	resp, err := registry.Client.Get(url)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

// UploadBlob can be used to upload an FS layer or an image config file into the given repository.
// It uploads the bytes read from content. Digest must match with the hash of those bytes.
// In case of token authentication the HTTP request must be retried after a 401 Unauthorized response
// (see https://docs.docker.com/registry/spec/auth/token/). In this case the getBody function is called
// in order to retrieve a fresh instance of the content reader. This behaviour matches exactly of the
// GetBody parameter of http.Client. This also means that if content is of type *bytes.Buffer,
// *bytes.Reader or *strings.Reader, then GetBody is populated automatically (as explained in the
// documentation of http.NewRequest()), so nil can be passed as the getBody parameter.
func (registry *Registry) UploadBlob(repository string, digest digest.Digest, content io.Reader, getBody func() (io.ReadCloser, error)) error {
	uploadUrl, err := registry.initiateUpload(repository)
	if err != nil {
		return err
	}
	q := uploadUrl.Query()
	q.Set("digest", digest.String())
	uploadUrl.RawQuery = q.Encode()

	registry.Logf("registry.blob.upload url=%s repository=%s digest=%s", uploadUrl, repository, digest)

	upload, err := http.NewRequest("PUT", uploadUrl.String(), content)
	if err != nil {
		return err
	}
	upload.Header.Set("Content-Type", "application/octet-stream")
	if getBody != nil {
		upload.GetBody = getBody
	}

	resp, err := registry.Client.Do(upload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		registry.Logf("registry.blob.upload read body failed url=%s repository=%s error=%s", uploadUrl, repository, err.Error())
		return err
	}
	registry.Logf("registry.blob.upload read body failed url=%s repository=%s body=%s", uploadUrl, repository, body)
	return nil
}

func (registry *Registry) MonolithicUploadBlob(repository string, digest digest.Digest, content io.Reader, contentLength int, getBody func() (io.ReadCloser, error)) (err error) {
	initiateUrl := registry.url("/v2/%s/blobs/uploads/?digest=%s", repository, digest.String())

	registry.Logf("registry.blob.Monolithic-upload url=%s repository=%s", initiateUrl, repository)

	// 1. init
	init, err := http.NewRequest("POST", initiateUrl, nil)
	if err != nil {
		registry.Logf("registry.blob.Monolithic-upload init NewRequest failed url=%s repository=%s error=%s", initiateUrl, repository, err.Error())

		return err
	}
	init.Header.Set("Content-Type", "application/octet-stream")
	init.Header.Set("Content-Length", fmt.Sprint(contentLength))
	initResp, err := registry.Client.Do(init)
	if err != nil {
		registry.Logf("registry.blob.Monolithic-upload init Do request failed url=%s repository=%s error=%s", initiateUrl, repository, err.Error())

		return err
	}
	logrus.WithField("status", initResp.StatusCode).
		WithField("Docker-Upload-UUID", initResp.Header.Get("Docker-Upload-UUID")).
		WithField("Location", initResp.Header.Get("Location")).
		Infof("Monolithic upload initResp")

	// 2. upload
	// uploadURL := registry.url("/v2/%s/blobs/uploads/%s?digest=%s", repository, initResp.Header.Get("Docker-Upload-UUID"), digest.String())
	uploadURL := fmt.Sprintf("%s&digest=%s", initResp.Header.Get("Location"), digest.String())
	upload, err := http.NewRequest("PUT", uploadURL, content)
	if getBody != nil {
		upload.GetBody = getBody
	}
	init.Header.Set("Content-Type", "application/octet-stream")
	init.Header.Set("Content-Length", fmt.Sprint(contentLength))

	resp, err := registry.Client.Do(upload)
	if err != nil {
		registry.Logf("registry.blob.Monolithic-upload Do request failed url=%s repository=%s error=%s", uploadURL, repository, err.Error())

		return err
	}

	logrus.WithField("status", initResp.StatusCode).
		WithField("Docker-Upload-UUID", initResp.Header.Get("Docker-Upload-UUID")).
		WithField("Range", initResp.Header.Get("Range")).
		Infof("Monolithic upload resp")

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		registry.Logf("registry.blob.Monolithic-upload read body failed url=%s repository=%s error=%s", uploadURL, repository, err.Error())
		return err
	}
	registry.Logf("registry.blob.Monolithic-upload read body failed url=%s repository=%s body=%s", uploadURL, repository, body)

	return
}

func (registry *Registry) HasBlob(repository string, digest digest.Digest) (bool, error) {
	checkUrl := registry.url("/v2/%s/blobs/%s", repository, digest)
	registry.Logf("registry.blob.check url=%s repository=%s digest=%s", checkUrl, repository, digest)

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

func (registry *Registry) BlobMetadata(repository string, digest digest.Digest) (distribution.Descriptor, error) {
	checkUrl := registry.url("/v2/%s/blobs/%s", repository, digest)
	registry.Logf("registry.blob.check url=%s repository=%s digest=%s", checkUrl, repository, digest)

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
	registry.Logf("registry.blob.initiate-upload url=%s repository=%s", initiateUrl, repository)

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
