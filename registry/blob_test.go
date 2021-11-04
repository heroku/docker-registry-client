package registry

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"

	digest "github.com/opencontainers/go-digest"
)

const dockerhubUrl = "https://registry-1.docker.io"

func dockerhubTestCredentials(t *testing.T) (username, password string) {
	const usernameEnv, passwordEnv = "DRC_TEST_DOCKERHUB_USERNAME", "DRC_TEST_DOCKERHUB_PASSWORD"
	username = os.Getenv(usernameEnv)
	password = os.Getenv(passwordEnv)
	if username == "" || password == "" {
		t.Skipf("DockerHub test credentials aren't specified in environment variables %s and %s", usernameEnv, passwordEnv)
	}
	return
}

func TestRegistry_UploadBlob(t *testing.T) {
	username, password := dockerhubTestCredentials(t)
	repository := username + "/docker-registry-client-test"
	registry, err := New(dockerhubUrl, username, password)
	if err != nil {
		t.Fatal("couldn't connect to registry:", err)
	}

	blobData := []byte("This is a test blob.")
	digest := digest.FromBytes(blobData)
	content := bytes.NewBuffer(blobData)
	err = registry.UploadBlob(repository, digest, content, nil)
	if err != nil {
		t.Error("couldn't upload blob:", err)
	}

}

func TestRegistry_UploadBlobFromFile(t *testing.T) {
	username, password := dockerhubTestCredentials(t)
	repository := username + "/docker-registry-client-test"
	registry, err := New(dockerhubUrl, username, password)
	if err != nil {
		t.Fatal("couldn't create registry client:", err)
	}

	// create blob file
	blobData := []byte("This is a test blob.")
	tmpfile, err := ioutil.TempFile("", "testblob")
	if err != nil {
		t.Fatal(err)
	}
	filename := tmpfile.Name()
	defer os.Remove(filename) // error deliberately ignored
	if _, err := tmpfile.Write(blobData); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// prepare UploadBlob() parameters
	digest := digest.FromBytes(blobData)
	body := func() (io.ReadCloser, error) {
		// NOTE: the file will be closed by UploadBlob() (actually the http.Client)
		return os.Open(filename)
	}
	blobReader, err := body()
	if err != nil {
		t.Fatal(err)
	}

	// call UploadBlob()
	err = registry.UploadBlob(repository, digest, blobReader, body)
	if err != nil {
		t.Error("UploadBlob() failed:", err)
	}
}
