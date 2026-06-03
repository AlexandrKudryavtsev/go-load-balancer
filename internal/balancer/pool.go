package balancer

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/AlexandrKudryavtsev/go-load-balancer/config"
)

type Pool struct {
	Backends []*Backend
}

func NewPool(cfg []config.BackendConfig) (*Pool, error) {
	backends := []*Backend{}

	for _, backend := range cfg {
		backendUrl, err := url.Parse(backend.URL)

		if err != nil {
			return nil, fmt.Errorf("failed create pool: %w", err)
		}

		proxy := httputil.NewSingleHostReverseProxy(backendUrl)
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, "bad gateway", http.StatusBadGateway)
		}

		backends = append(backends, &Backend{
			URL:   backendUrl,
			Alive: true,
			Proxy: proxy,
		})
	}

	return &Pool{
		Backends: backends,
	}, nil
}
