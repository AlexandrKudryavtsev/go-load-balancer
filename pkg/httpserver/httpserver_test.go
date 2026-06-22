package httpserver

import (
	"net/http"
	"testing"
	"time"
)

func newTestServer(t *testing.T, opts ...Option) *Server {
	t.Helper()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return New(handler, opts...)
}

func TestNew(t *testing.T) {
	srv := newTestServer(t)

	if srv == nil {
		t.Fatalf("New() = nil, want Server")
	}

	if srv.notify == nil {
		t.Fatal("Server.notify = nil, want channel")
	}

	if srv.server.Handler == nil {
		t.Fatal("Server.Handler = nil, want handler")
	}

	if srv.server.Addr != defaultAddress {
		t.Errorf(
			"Server.Addr = %q, want %q",
			srv.server.Addr,
			defaultAddress,
		)
	}

	if srv.server.ReadTimeout != defaultReadTimeout {
		t.Errorf(
			"Server.ReadTimeout = %v, want %v",
			srv.server.ReadTimeout,
			defaultReadTimeout,
		)
	}

	if srv.server.WriteTimeout != defaultWriteTimeout {
		t.Errorf(
			"Server.WriteTimeout = %v, want %v",
			srv.server.WriteTimeout,
			defaultWriteTimeout,
		)
	}

	if srv.shutdownTimeout != defaultShutdownTimeout {
		t.Errorf(
			"Server.ShutdownTimeout = %v, want %v",
			srv.shutdownTimeout,
			defaultShutdownTimeout,
		)
	}
}

func TestNewAppliesOptions(t *testing.T) {
	srv := newTestServer(t, Port(8080), ReadTimeout(2*time.Second))

	if srv == nil {
		t.Fatal("New() = nil, want server")
	}

	if srv.server.Addr != ":8080" {
		t.Errorf(
			"Server.Addr = %q, want %q",
			srv.server.Addr, ":8080",
		)
	}

	if srv.server.ReadTimeout != 2*time.Second {
		t.Errorf(
			"Server.ReadTimeout = %v, want %v",
			srv.server.ReadTimeout, 2*time.Second,
		)
	}
}

func TestNotifyReturnsNotifyChannel(t *testing.T) {
	srv := newTestServer(t)

	if srv == nil {
		t.Fatal("New() = nil, want server")
	}

	notifyCh := srv.Notify()

	if notifyCh != srv.notify {
		t.Fatal("Notify() did not return server notify channel")
	}
}

func TestStartAndShutdown(t *testing.T) {
	srv := newTestServer(t, Port(0), ShutdownTimeout(1*time.Second))

	srv.Start()

	if err := srv.Shutdown(); err != nil {
		t.Fatalf("Shutdown() unexpected error: %v", err)
	}

	select {
	case err, ok := <-srv.Notify():
		if ok && err != nil {
			t.Fatalf("Notify() error = %v, want nil", err)
		}
	case <-time.After(time.Second):
		t.Fatal("server did not stop")
	}
}
