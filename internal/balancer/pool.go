package balancer

import (
	"net/url"

	"github.com/AlexandrKudryavtsev/go-load-balancer/config"
)

type Pool struct {
	Backends []*Backend
}

func NewPool(cfg []config.BackendConfig) *Pool {
	backends := []*Backend{}

	for _, backend := range cfg {
		url, _ := url.Parse(backend.URL)

		backends = append(backends, &Backend{
			URL:   url,
			Alive: true,
		})
	}

	return &Pool{
		Backends: backends,
	}
}
