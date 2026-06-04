package balancer

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/AlexandrKudryavtsev/go-load-balancer/config"
)

type Pool struct {
	Backends []*Backend
	client   *http.Client
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

		backend := &Backend{
			URL:        backendUrl,
			HealthPath: backend.HealthPath,
			Alive:      atomic.Bool{},
			Proxy:      proxy,
		}

		backends = append(backends, backend)
	}

	return &Pool{
		Backends: backends,
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
	}, nil
}

func (p *Pool) StartHealthChecker(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	p.check()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fmt.Println("health checker")
			p.check()
		}
	}
}

func (p *Pool) check() {
	for _, backend := range p.Backends {
		resp, err := p.client.Get(backend.URL.String() + backend.HealthPath)
		if err != nil {
			backend.Alive.Store(false)
			continue
		}

		_ = resp.Body.Close()
		backend.Alive.Store(resp.StatusCode == http.StatusOK)
	}
}
