package balancer

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AlexandrKudryavtsev/go-load-balancer/config"
)

func newTestLogger() *slog.Logger {
	return slog.New(
		slog.NewJSONHandler(io.Discard, nil),
	)
}

func TestNewPool(t *testing.T) {
	cfg := []config.BackendConfig{
		{
			URL:        "http://localhost:9001",
			HealthPath: "/health",
			Weight:     5,
		},
		{
			URL:        "http://localhost:9002",
			HealthPath: "/status",
			Weight:     3,
		},
	}

	pool, err := NewPool(cfg, newTestLogger())
	if err != nil {
		t.Fatalf("NewPool() unexpected error: %v", err)
	}
	if pool == nil {
		t.Fatalf("NewPool() = nil, want pool")
	}
	if len(pool.Backends) != 2 {
		t.Fatalf("len(Backends) = %d, want %d", len(pool.Backends), 2)
	}
	if pool.client == nil {
		t.Fatalf("Pool.client = nil, want client")
	}

	backend := pool.Backends[0]

	if backend.URL.String() != "http://localhost:9001" {
		t.Errorf("Backend.URL = %q, want %q", backend.URL.String(), "http://localhost:9001")
	}
	if backend.HealthPath != "/health" {
		t.Errorf("Backend.HealthPath = %q, want %q", backend.HealthPath, "/health")
	}
	if backend.Weight != 5 {
		t.Errorf("Backend.Weight = %d, want %d", backend.Weight, 5)
	}
	if backend.CurrentWeight != 0 {
		t.Errorf("Backend.CurrentWeight = %d, want %d", backend.CurrentWeight, 0)
	}
	if !backend.Alive.Load() {
		t.Errorf("Backend.Alive = false, want true")
	}
	if backend.Proxy == nil {
		t.Errorf("Backend.Proxy = nil, want proxy")
	}
}

func TestNewPoolReturnsErrorForMalformedURL(t *testing.T) {
	cfg := []config.BackendConfig{
		{
			URL:        "http://%zz",
			HealthPath: "/health",
			Weight:     1,
		},
	}

	pool, err := NewPool(cfg, newTestLogger())
	if err == nil {
		t.Fatalf("NewPool() expected error, got nil")
	}
	if pool != nil {
		t.Fatalf("NewPool() pool = %#v, want nil", pool)
	}
}

func TestPoolCheckHealth(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	cfg := []config.BackendConfig{
		{
			URL:        server.URL,
			HealthPath: "/health",
			Weight:     5,
		},
	}

	pool, err := NewPool(cfg, newTestLogger())
	if err != nil {
		t.Fatalf("NewPool() unexpected error: %v", err)
	}
	if len(pool.Backends) != 1 {
		t.Fatalf("len(Backends) = %d, want %d", len(pool.Backends), 1)
	}

	pool.Backends[0].Alive.Store(false)
	pool.check()

	if !pool.Backends[0].Alive.Load() {
		t.Fatalf("Backends.Alive = false, want true")
	}
}

func TestPoolCheckMarksBackendDeadOnServerError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	cfg := []config.BackendConfig{
		{
			URL:        server.URL,
			HealthPath: "/health",
			Weight:     5,
		},
	}

	pool, err := NewPool(cfg, newTestLogger())
	if err != nil {
		t.Fatalf("NewPool() unexpected error: %v", err)
	}
	if len(pool.Backends) != 1 {
		t.Fatalf("len(Backends) = %d, want %d", len(pool.Backends), 1)
	}

	pool.Backends[0].Alive.Store(true)
	pool.check()

	if pool.Backends[0].Alive.Load() {
		t.Fatalf("Backends.Alive = true, want false")
	}
}

func TestPoolCheckHealthWithClosedServer(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	server := httptest.NewServer(mux)
	server.Close()

	cfg := []config.BackendConfig{
		{
			URL:        server.URL,
			HealthPath: "/health",
			Weight:     5,
		},
	}

	pool, err := NewPool(cfg, newTestLogger())
	if err != nil {
		t.Fatalf("NewPool() unexpected error: %v", err)
	}
	if len(pool.Backends) != 1 {
		t.Fatalf("len(Backends) = %d, want %d", len(pool.Backends), 1)
	}

	pool.Backends[0].Alive.Store(true)
	pool.check()

	if pool.Backends[0].Alive.Load() {
		t.Fatalf("Backend.Alive = true, want false")
	}
}

func TestStartHealthCheckerChecksImmediatelyAndStopsOnContextCancel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := []config.BackendConfig{
		{
			URL:        server.URL,
			HealthPath: "/health",
			Weight:     1,
		},
	}

	pool, err := NewPool(cfg, newTestLogger())
	if err != nil {
		t.Fatalf("NewPool() unexpected error: %v", err)
	}
	if len(pool.Backends) != 1 {
		t.Fatalf("len(Backends) = %d, want %d", len(pool.Backends), 1)
	}

	pool.Backends[0].Alive.Store(false)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})

	go func() {
		defer close(done)
		pool.StartHealthChecker(ctx, 1*time.Hour)
	}()

	deadline := time.After(1 * time.Second)
	for !pool.Backends[0].Alive.Load() {
		select {
		case <-deadline:
			cancel()
			t.Fatalf("backend was not marked alive")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	cancel()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("StartHealthChecker() did not stop after context cancel")
	}
}
