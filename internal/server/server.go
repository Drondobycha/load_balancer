package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"load_balancer/internal/balancer"
	"load_balancer/internal/ratelimiter"
)

type Server struct {
	port       string
	balancer   *balancer.LoadBalancer
	limiter    *ratelimiter.RateLimiter
	httpServer *http.Server
}

func NewServer(port string, balancer *balancer.LoadBalancer, limiter *ratelimiter.RateLimiter) *Server {
	return &Server{
		port:     port,
		balancer: balancer,
		limiter:  limiter,
	}
}

func (s *Server) Start() error {
	router := http.NewServeMux()
	router.HandleFunc("/", s.handleRequest)

	s.httpServer = &http.Server{
		Addr:    ":" + s.port,
		Handler: router,
	}

	log.Printf("Starting server on port %s", s.port)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	clientIP := getClientIP(r)

	if !s.limiter.Allow(clientIP, r.URL.Path) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprintf(w, `{"code": 429, "message": "Rate limit exceeded"}`)
		return
	}

	log.Printf("Request from %s to %s", clientIP, r.URL.Path)
	s.balancer.ServeHTTP(w, r)
}

func getClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}
