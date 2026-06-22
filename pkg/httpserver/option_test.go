package httpserver

import (
	"testing"
	"time"
)

func TestOptions(t *testing.T) {
	tests := []struct {
		name  string
		apply Option
		check func(t *testing.T, srv *Server)
	}{
		{
			name:  "port",
			apply: Port(8080),
			check: func(t *testing.T, srv *Server) {
				t.Helper()

				if srv.server.Addr != ":8080" {
					t.Errorf("Server.Addr = %q, want %q", srv.server.Addr, ":8080")
				}
			},
		},
		{
			name:  "read timeout",
			apply: ReadTimeout(2 * time.Second),
			check: func(t *testing.T, srv *Server) {
				t.Helper()

				if srv.server.ReadTimeout != 2*time.Second {
					t.Errorf(
						"Server.ReadTimeout = %v, want %v",
						srv.server.ReadTimeout, 2*time.Second,
					)
				}
			},
		},
		{
			name:  "write timeout",
			apply: WriteTimeout(4 * time.Second),
			check: func(t *testing.T, srv *Server) {
				t.Helper()

				if srv.server.WriteTimeout != 4*time.Second {
					t.Errorf(
						"Server.WriteTimeout = %v, want %v",
						srv.server.WriteTimeout, 4*time.Second,
					)
				}
			},
		},
		{
			name:  "shutdown timeout",
			apply: ShutdownTimeout(3 * time.Second),
			check: func(t *testing.T, srv *Server) {
				t.Helper()

				if srv.shutdownTimeout != 3*time.Second {
					t.Errorf(
						"Server.ShutdownTimeout = %v, want %v",
						srv.shutdownTimeout, 3*time.Second,
					)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := newTestServer(t)

			tt.apply(srv)
			tt.check(t, srv)
		})
	}
}
