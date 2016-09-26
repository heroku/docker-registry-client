package registry

type catalogResponse struct {
	Repositories []string `json:"repositories"`
}

func (registry *Registry) Catalog() ([]string, error) {
	url := registry.url("/v2/_catalog")
	registry.Logf("registry.repositories url=%s", url)

	var response catalogResponse
	if err := registry.getJson(url, &response); err != nil {
		return nil, err
	}

	return response.Repositories, nil
}
