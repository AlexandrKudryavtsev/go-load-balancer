package balancer

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlexandrKudryavtsev/go-load-balancer/config"
)

func newTestLoadBalancerServer(t *testing.T, pool *Pool) *httptest.Server {
	t.Helper()

	balancer := New(pool)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		backend := balancer.Next()
		if backend == nil {
			http.Error(w, "no available backends", http.StatusServiceUnavailable)
			return
		}

		backend.Proxy.ServeHTTP(w, r)
	})

	return httptest.NewServer(handler)
}

func TestLoadBalancerProxiesRequestsUsingWeights(t *testing.T) {
	srv1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("backend-1"))
	}))
	defer srv1.Close()
	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("backend-2"))
	}))
	defer srv2.Close()

	cfg := []config.BackendConfig{
		{
			URL:        srv1.URL,
			HealthPath: "/health",
			Weight:     2,
		},
		{
			URL:        srv2.URL,
			HealthPath: "/health",
			Weight:     1,
		},
	}
	pool, err := NewPool(cfg, newTestLogger())
	if err != nil {
		t.Fatalf("NewPool() unexpected error: %v", err)
	}

	srvBalancer := newTestLoadBalancerServer(t, pool)
	defer srvBalancer.Close()

	counts := make(map[string]int, 2)

	for i := 0; i < 3; i++ {
		resp, err := http.Get(srvBalancer.URL)
		if err != nil {
			t.Fatalf("Get() unexpected error: %v", err)
		}

		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
		}
		if err != nil {
			t.Fatalf("ReadAll() unexpected error: %v", err)
		}

		got := string(body)
		if got != "backend-1" && got != "backend-2" {
			t.Fatalf("body = %q, want backend response", got)
		}
		counts[got]++
	}

	if counts["backend-1"] != 2 {
		t.Errorf("counts[backend-1] = %d, want %d", counts["backend-1"], 2)
	}
	if counts["backend-2"] != 1 {
		t.Errorf("counts[backend-2] = %d, want %d", counts["backend-2"], 1)
	}
}

func TestLoadBalancerReturnsBadGateway(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	srv.Close()
	pool, err := NewPool(
		[]config.BackendConfig{
			{
				URL:        srv.URL,
				HealthPath: "/health",
				Weight:     1,
			},
		},
		newTestLogger(),
	)
	if err != nil {
		t.Fatalf("NewPool() unexpected error: %v", err)
	}
	pool.Backends[0].Alive.Store(true)

	srvBalancer := newTestLoadBalancerServer(t, pool)
	defer srvBalancer.Close()

	resp, err := http.Get(srvBalancer.URL)
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusBadGateway)
	}

	if pool.Backends[0].Alive.Load() {
		t.Error("Backend.Alive = true, want false")
	}
}

func TestLoadBalancerReturnsServiceUnavailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("backend should not be called")
	}))
	defer srv.Close()
	pool, err := NewPool(
		[]config.BackendConfig{
			{
				URL:        srv.URL,
				HealthPath: "/health",
				Weight:     1,
			},
		},
		newTestLogger(),
	)
	if err != nil {
		t.Fatalf("NewPool() unexpected error: %v", err)
	}
	pool.Backends[0].Alive.Store(false)

	srvBalancer := newTestLoadBalancerServer(t, pool)
	defer srvBalancer.Close()

	resp, err := http.Get(srvBalancer.URL)
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusServiceUnavailable)
	}
}
