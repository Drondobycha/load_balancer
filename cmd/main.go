package main

import (
	"context"
	"load_balancer/internal/balancer"
	"load_balancer/internal/config"
	"load_balancer/internal/ratelimiter"
	"load_balancer/internal/server"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	round_robin = "round_robin"
	least_conn  = "least_conn"
)

func main() {
	cfg := config.MustLoad()
	log.Print("starting application", slog.Any("config", cfg))
	limiter := ratelimiter.NewRateLimiter(cfg.RateLimit.DefaultRate, cfg.RateLimit.DefaultCapacity, nil)
	limiter.StartAutoRefill(5 * time.Second)
	// Initialize load balancer
	lb := balancer.NewLoadBalancer(cfg.Backends, cfg.HealsthCheckInterval)
	switch cfg.Strategy {
	case round_robin:
		lb.SetStrategy(&balancer.RoundRobinStrategy{})
		log.Printf("Using round robin strategy")
	case least_conn:
		leastConn := balancer.NewLeastConnStrategy()
		lb.SetStrategy(leastConn)
		log.Printf("Using least connection strategy")
	default:
		log.Fatalf("Unknown strategy: %s", cfg.Strategy)
	}

	// Start health checks
	go lb.StartHealthChecks()

	// Create HTTP server
	srv := server.NewServer(cfg.Port, lb, limiter)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.Start(); err != nil {
			log.Printf("Server error: %v", err)
			quit <- syscall.SIGTERM
		}
	}()

	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	// Stop rate limiter
	limiter.Stop()

	log.Println("Server exiting")
}
