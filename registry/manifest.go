package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	digest "github.com/opencontainers/go-digest"
)

// Manifest returns with the schema1 manifest addressed by repository/reference
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

// ManifestList returns with the ManifestList addressed by repository/reference
func (registry *Registry) ManifestList(repository, reference string) (*manifestlist.DeserializedManifestList, error) {
	url := registry.url("/v2/%s/manifests/%s", repository, reference)
	registry.Logf("registry.manifestlist.get url=%s repository=%s reference=%s", url, repository, reference)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", manifestlist.MediaTypeManifestList)
	resp, err := registry.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	ml := new(manifestlist.DeserializedManifestList)
	err = json.NewDecoder(resp.Body).Decode(ml)
	if err != nil {
		return nil, err
	}

	if ml.MediaType != manifestlist.MediaTypeManifestList {
		err = fmt.Errorf("mediaType in manifest list should be '%s' not '%s'",
			manifestlist.MediaTypeManifestList, ml.MediaType)
		return nil, err
	}
	return ml, nil
}

// ManifestV2 returns with the schema2 manifest addressed by repository/reference
// If reference is an image digest (sha256:...) that is the hash of a manifestlist,
// not a manifest, then the method will return the first manifest with amd64 architecture,
// or in the absence thereof, the first manifest in the list. (Rationale: the Docker
// Image Digests returned by `docker images --digests` often refer to manifestlists)
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

	actualMediaType := resp.Header.Get("Content-Type")
	switch actualMediaType {
	case schema2.MediaTypeManifest:
		deserialized := &schema2.DeserializedManifest{}
		err = deserialized.UnmarshalJSON(body)
		if err != nil {
			return nil, err
		}
		return deserialized, nil

	case manifestlist.MediaTypeManifestList:
		// if `reference` is an image digest, a manifest list may be received even though a schema2 manifest was requested
		// (since the image digest is the hash of the manifest list, not the manifest)
		// unwrap the referred manifest in this case
		ml := new(manifestlist.DeserializedManifestList)
		err := ml.UnmarshalJSON(body)
		if err != nil {
			return nil, err
		}

		if ml.MediaType != manifestlist.MediaTypeManifestList {
			err = fmt.Errorf("mediaType in manifest list should be '%s' not '%s'",
				manifestlist.MediaTypeManifestList, ml.MediaType)
			return nil, err
		}
		if len(ml.Manifests) == 0 {
			return nil, fmt.Errorf("empty manifest list was receceived: repository=%s reference=%s", repository, reference)
		}

		// use the amd64 manifest by default
		// TODO: query current platform architecture, OS and Variant and use those as selection criteria
		for _, m := range ml.Manifests {
			if m.Platform.Architecture == "amd64" {
				// address the manifest explicitly with its digest
				return registry.ManifestV2(repository, m.Digest.String())
			}
		}
		// fallback: use the first manifest in the list
		// NOTE: emptiness of the list was checked above
		return registry.ManifestV2(repository, ml.Manifests[0].Digest.String())

	default:
		return nil, fmt.Errorf("unexpected manifest schema was received from registry: mediatype should be %s, not %s (registry may not support schema2 manifests)", schema2.MediaTypeManifest, actualMediaType)
	}
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
