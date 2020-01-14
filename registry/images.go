package registry

type imageResponse struct {
	Manifest map[string]interface{} `json:manifest`
}

func (registry *Registry) Images(repository string) (images []string, err error) {
	url := registry.url("/v2/%s/tags/list", repository)

	var response imageResponse
	for {
		registry.Logf("registry.images url=%s repository=%s", url, repository)
		url, err = registry.getPaginatedJson(url, &response)

		switch err {
		case ErrNoMorePages:
			for imageDigest := range response.Manifest {
				images = append(images, imageDigest)
			}
			return images, nil
		case nil:
			for imageDigest := range response.Manifest {
				images = append(images, imageDigest)
			}
			continue
		default:
			return nil, err
		}
	}
}
