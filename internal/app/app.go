package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/AlexandrKudryavtsev/go-load-balancer/config"
	"github.com/AlexandrKudryavtsev/go-load-balancer/internal/balancer"
	"github.com/AlexandrKudryavtsev/go-load-balancer/pkg/httpserver"
	"github.com/AlexandrKudryavtsev/go-load-balancer/pkg/logger"
)

func Run(cfg *config.Config) {
	mux := http.NewServeMux()

	log := logger.New(cfg.Logger)

	pool, err := balancer.NewPool(cfg.Backends, log)
	if err != nil {
		log.Error("failed create new pool", "error", err)
		return
	}

	b := balancer.New(pool)

	healthCheckerContext, cancelHealthCheckerContext := context.WithCancel(context.Background())
	defer cancelHealthCheckerContext()
	go pool.StartHealthChecker(healthCheckerContext, cfg.HealthCheck.Interval.Duration)

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
		httpserver.ReadTimeout(cfg.Server.ReadTimeout.Duration),
		httpserver.WriteTimeout(cfg.Server.WriteTimeout.Duration),
		httpserver.ShutdownTimeout(cfg.Server.ShutdownTimeout.Duration),
	)

	httpServer.Start()

	log.Info(
		"load balancer started",
		"port", cfg.Server.Port,
		"backends", len(cfg.Backends),
	)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	select {
	case err, ok := <-httpServer.Notify():
		if ok && err != nil {
			log.Error("http server", "error", err)
		}

	case sig := <-quit:
		log.Info("received signal", "signal", sig)
		cancelHealthCheckerContext()

		if err := httpServer.Shutdown(); err != nil {
			log.Error("shutdown error", "error", err)
			return
		}

		log.Info("http server stopped")
	}
}
