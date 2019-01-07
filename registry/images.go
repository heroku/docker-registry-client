package registry

type imagesResponse struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

func (registry *Registry) Images(repository string) (images []imagesResponse, err error) {
	url := registry.url("/v2/%s/tags/list", repository)

	var response tagsResponse
	for {
		registry.Logf("registry.tags url=%s repository=%s", url, repository)
		url, err = registry.getPaginatedJson(url, &response)
		switch err {
		case ErrNoMorePages:
			tags = append(tags, response.Tags...)
			return tags, nil
		case nil:
			tags = append(tags, response.Tags...)
			continue
		default:
			return nil, err
		}
	}
}
