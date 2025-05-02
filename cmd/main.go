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

	// Инициализация балансировщика
	lb := balancer.NewLoadBalancer(cfg.Backends, cfg.HealsthCheckInterval)
	if cfg.Strategy == round_robin {
		lb.SetStrategy(&balancer.RoundRobinStrategy{})
		log.Printf("Using round robin strategy")
	} else if cfg.Strategy == least_conn {
		leastConn := balancer.NewLeastConnStrategy()
		lb.SetStrategy(leastConn)
		log.Printf("Using least connection strategy")
	} else {
		log.Fatalf("Unknown strategy: %s", cfg.Strategy)
	}
	// Запуск health checks
	go lb.StartHealthChecks()

	// Создание HTTP сервера
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

	log.Println("Server exiting")
}
