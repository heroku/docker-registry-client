package registry_test

import (
	"testing"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/opencontainers/go-digest"
)

func checkManifest(t *testing.T, tc *TestCase,
	wantDigest *string, wantMediaType string,
	getManifest func(t *testing.T) (distribution.Manifest, error)) {

	if *wantDigest == "" {
		return
	}

	t.Run(tc.Name(), func(t *testing.T) {
		got, err := getManifest(t)
		if err != nil {
			t.Error(err)
			return
		}
		mediaType, payload, err := got.Payload()
		if err != nil {
			t.Error("Payload() error:", err)
			return
		}
		d := digest.FromBytes(payload).String()

		if !*_testDataUpdate {
			// do actual testing of manifest
			if mediaType != wantMediaType {
				t.Errorf("MediaType = %v, want %v", mediaType, wantMediaType)
			}
			if d != *wantDigest {
				t.Errorf("digest = %v, want %v", d, *wantDigest)
			}
			if !blobSlicesAreEqual(got.References(), tc.Blobs) {
				t.Errorf("\nblobs:\n%v,\nwant:\n%v", got.References(), tc.Blobs)
			}
		} else {
			// update TestCase to reflect the result of the tested method
			*wantDigest = d
			tc.Blobs = got.References()
		}
	})
}

func TestRegistry_Manifest(t *testing.T) {
	for _, tc := range testCases(t) {
		checkManifest(t, tc, &tc.ManifestV1Digest, schema1.MediaTypeSignedManifest, func(t *testing.T) (distribution.Manifest, error) {
			return tc.Registry(t).Manifest(tc.Repository, tc.Reference)
		})
	}
	updateTestData(t)
}

func TestRegistry_ManifestV2(t *testing.T) {
	for _, tc := range testCases(t) {
		checkManifest(t, tc, &tc.ManifestV2Digest, schema2.MediaTypeManifest, func(t *testing.T) (distribution.Manifest, error) {
			return tc.Registry(t).ManifestV2(tc.Repository, tc.Reference)
		})
	}
	updateTestData(t)
}
