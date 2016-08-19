package registry

type repositoriesResponse struct {
	Repositories []string `json:"repositories"`
}

func (registry *Registry) Repositories() ([]string, error) {
	url := registry.url("/v2/_catalog")
	registry.Logf("registry.repositories url=%s", url)

	var response repositoriesResponse
	if err := registry.getJson(url, &response); err != nil {
		return nil, err
	}

	return response.Repositories, nil
}
