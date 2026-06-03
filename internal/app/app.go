package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AlexandrKudryavtsev/go-load-balancer/config"
	"github.com/AlexandrKudryavtsev/go-load-balancer/internal/balancer"
	"github.com/AlexandrKudryavtsev/go-load-balancer/pkg/httpserver"
)

func Run(cfg *config.Config) {
	mux := http.NewServeMux()

	pool, err := balancer.NewPool(cfg.Backends)
	if err != nil {
		fmt.Printf("failed create new pool: %v\n", err)
		return
	}

	d, err := time.ParseDuration(cfg.Server.HealthCheckInterval)
	if err != nil {
		fmt.Printf("failed parse duration: %v\n", err)
		return
	}
	b := balancer.New(pool)

	healthCheckerContext, cancelHealthCheckerContext := context.WithCancel(context.Background())
	defer cancelHealthCheckerContext()
	go pool.StartHealthChecker(healthCheckerContext, d)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		backend := b.Next()
		if backend == nil {
			http.Error(w, "no available backends", http.StatusServiceUnavailable)
			return
		}

		backend.Proxy.ServeHTTP(w, r)
	})

	httpServer := httpserver.New(
		mux,
		httpserver.Port(cfg.Server.Port),
	)

	httpServer.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	select {
	case err, ok := <-httpServer.Notify():
		if ok && err != nil {
			fmt.Printf("http server error: %v\n", err)
		}

	case sig := <-quit:
		fmt.Printf("received signal: %s\n", sig)
		cancelHealthCheckerContext()

		if err := httpServer.Shutdown(); err != nil {
			fmt.Printf("shutdown error: %v\n", err)
			return
		}

		fmt.Println("http server stopped")
	}
}
