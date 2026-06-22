package balancer

import (
	"net/url"
	"testing"
)

func newTestBackend(t *testing.T, rawURL string, weight int, alive bool) *Backend {
	t.Helper()

	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("url.Parse() error: %v", err)
	}

	backend := &Backend{
		URL:    u,
		Weight: weight,
	}
	backend.Alive.Store(alive)

	return backend
}

func TestNextOneLiveBackend(t *testing.T) {
	pool := &Pool{
		Backends: []*Backend{
			newTestBackend(t, "http://backend-1", 5, true),
			newTestBackend(t, "http://backend-2", 3, false),
			newTestBackend(t, "http://backend-3", 1, false),
		},
	}

	b := New(pool)
	backend := b.Next()
	want := pool.Backends[0]

	if backend == nil {
		t.Fatalf("Next() = nil, want backend")
	}
	if backend != want {
		t.Fatalf(
			"Next() = %s, want %s",
			backend.URL.String(),
			want.URL.String(),
		)
	}
}

func TestNextWithNoAliveBackends(t *testing.T) {
	pool := &Pool{
		Backends: []*Backend{
			newTestBackend(t, "http://backend-1", 5, false),
			newTestBackend(t, "http://backend-2", 3, false),
			newTestBackend(t, "http://backend-3", 1, false),
		},
	}

	b := New(pool)
	backend := b.Next()

	if backend != nil {
		t.Fatalf("Next() = %v, want nil", backend)
	}
}

func TestNextUsesWeightedRoundRobin(t *testing.T) {
	pool := &Pool{
		Backends: []*Backend{
			newTestBackend(t, "http://backend-1", 5, true),
			newTestBackend(t, "http://backend-2", 3, true),
			newTestBackend(t, "http://backend-3", 1, true),
		},
	}

	b := New(pool)

	counts := make(map[*Backend]int, 3)

	for i := 0; i < 9; i++ {
		backend := b.Next()
		if backend == nil {
			t.Fatal("Next() = nil, want backend")
		}

		counts[backend]++
	}

	if counts[pool.Backends[0]] != 5 {
		t.Errorf("backend-1 count = %d, want %d", counts[pool.Backends[0]], 5)
	}
	if counts[pool.Backends[1]] != 3 {
		t.Errorf("backend-2 count = %d, want %d", counts[pool.Backends[1]], 3)
	}
	if counts[pool.Backends[2]] != 1 {
		t.Errorf("backend-3 count = %d, want %d", counts[pool.Backends[2]], 1)
	}
}

func TestNextSkipsDeadBackends(t *testing.T) {
	pool := &Pool{
		Backends: []*Backend{
			newTestBackend(t, "http://backend-1", 100, false),
			newTestBackend(t, "http://backend-2", 1, true),
		},
	}

	b := New(pool)

	backend := b.Next()
	if backend == nil {
		t.Fatal("Next() = nil, want backend")
	}
	if backend != pool.Backends[1] {
		t.Fatalf(
			"Next() = %s, want %s",
			backend.URL.String(),
			pool.Backends[1].URL.String(),
		)
	}
}
