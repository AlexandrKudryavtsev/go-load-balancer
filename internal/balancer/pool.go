package balancer

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/AlexandrKudryavtsev/go-load-balancer/config"
)

type Pool struct {
	Backends []*Backend
	client   *http.Client
	log      *slog.Logger
}

func NewPool(cfg []config.BackendConfig, log *slog.Logger) (*Pool, error) {
	backends := make([]*Backend, 0, len(cfg))

	for _, backendCfg := range cfg {
		backendURL, err := url.Parse(backendCfg.URL)
		if err != nil {
			return nil, fmt.Errorf("failed parse backend url %q: %w", backendCfg.URL, err)
		}

		proxy := httputil.NewSingleHostReverseProxy(backendURL)

		backend := &Backend{
			URL:        backendURL,
			HealthPath: backendCfg.HealthPath,
			Proxy:      proxy,
		}

		backend.Alive.Store(true)

		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			backend.Alive.Store(false)

			log.Error(
				"proxy error",
				"backend", backend.URL.String(),
				"error", err,
			)

			http.Error(w, "bad gateway", http.StatusBadGateway)
		}

		backends = append(backends, backend)

		log.Info(
			"backend registered",
			"backend", backend.URL.String(),
			"health_path", backend.HealthPath,
		)
	}

	return &Pool{
		Backends: backends,
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
		log: log,
	}, nil
}

func (p *Pool) StartHealthChecker(ctx context.Context, interval time.Duration) {
	p.log.Info(
		"health checker started",
		"interval", interval.String(),
	)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	p.check()

	for {
		select {
		case <-ctx.Done():
			p.log.Info("health checker stopped")
			return

		case <-ticker.C:
			p.check()
		}
	}
}

func (p *Pool) check() {
	for _, backend := range p.Backends {
		wasAlive := backend.Alive.Load()

		resp, err := p.client.Get(backend.URL.String() + backend.HealthPath)
		if err != nil {
			if wasAlive {
				p.log.Warn(
					"backend marked dead",
					"backend", backend.URL.String(),
					"error", err,
				)
			}

			backend.Alive.Store(false)
			continue
		}

		_ = resp.Body.Close()

		alive := resp.StatusCode == http.StatusOK

		if wasAlive != alive {
			p.log.Info(
				"backend state changed",
				"backend", backend.URL.String(),
				"alive", alive,
				"status_code", resp.StatusCode,
			)
		}

		backend.Alive.Store(alive)
	}
}
