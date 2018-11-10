package registry

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	digest "github.com/opencontainers/go-digest"
)

func (registry *Registry) Manifest(repository, reference string) (*schema1.SignedManifest, error) {
	url := registry.url("/v2/%s/manifests/%s", repository, reference)
	registry.Logf("registry.manifest.get url=%s repository=%s reference=%s", url, repository, reference)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", schema1.MediaTypeManifest)
	resp, err := registry.Client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	signedManifest := &schema1.SignedManifest{}
	err = signedManifest.UnmarshalJSON(body)
	if err != nil {
		return nil, err
	}

	return signedManifest, nil
}

func (registry *Registry) ManifestV2(repository, reference string) (*schema2.DeserializedManifest, error) {
	url := registry.url("/v2/%s/manifests/%s", repository, reference)
	registry.Logf("registry.manifest.get url=%s repository=%s reference=%s", url, repository, reference)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", schema2.MediaTypeManifest)
	resp, err := registry.Client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	deserialized := &schema2.DeserializedManifest{}
	err = deserialized.UnmarshalJSON(body)
	if err != nil {
		return nil, err
	}
	return deserialized, nil
}

func (registry *Registry) ManifestDigest(repository, reference string) (digest.Digest, error) {
	url := registry.url("/v2/%s/manifests/%s", repository, reference)
	registry.Logf("registry.manifest.head url=%s repository=%s reference=%s", url, repository, reference)

	resp, err := registry.Client.Head(url)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return "", err
	}
	return digest.Parse(resp.Header.Get("Docker-Content-Digest"))
}

func (registry *Registry) DeleteManifest(repository string, digest digest.Digest) error {
	url := registry.url("/v2/%s/manifests/%s", repository, digest)
	registry.Logf("registry.manifest.delete url=%s repository=%s reference=%s", url, repository, digest)

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

// PutManifest uploads manifest to the given repository/reference.
// Manifest is typically either of type schema2.DeserializedManifest or schema1.SignedManifest
func (registry *Registry) PutManifest(repository, reference string, manifest distribution.Manifest) error {
	url := registry.url("/v2/%s/manifests/%s", repository, reference)
	registry.Logf("registry.manifest.put url=%s repository=%s reference=%s", url, repository, reference)

	mediaType, payload, err := manifest.Payload()
	if err != nil {
		return err
	}

	buffer := bytes.NewBuffer(payload)
	req, err := http.NewRequest("PUT", url, buffer)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", mediaType)
	resp, err := registry.Client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	return err
}

// PutManifestV2 uploads the given schama2.Manifest to the given repository/reference and returns with its digest.
// If you want to upload a schema2.DeserializedManifest, please use the generic PutManifest().
func (registry *Registry) PutManifestV2(repository, reference string, manifest *schema2.Manifest) (digest.Digest, error) {
	deserializedManifest, err := schema2.FromStruct(*manifest)
	if err != nil {
		return "", err
	}
	_, canonical, err := deserializedManifest.Payload()
	if err != nil {
		return "", err
	}
	digest := digest.FromBytes(canonical)
	err = registry.PutManifest(repository, reference, deserializedManifest)
	return digest, err
}
