package balancer

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

type LoadBalancer struct {
	backends            []*Backend
	mu                  sync.RWMutex
	strategy            Strategy
	healthCheckInterval time.Duration
}

type Backend struct {
	URL          *url.URL
	Alive        bool
	ReverseProxy *httputil.ReverseProxy
	mu           sync.RWMutex
}

func NewLoadBalancer(backendURLs []string, healthCheckInterval time.Duration) *LoadBalancer {
	var backends []*Backend

	for _, backendURL := range backendURLs {
		parsedURL, err := url.Parse(backendURL)
		if err != nil {
			log.Printf("Failed to parse backend URL %s: %v", backendURL, err)
		} else {
			proxy := httputil.NewSingleHostReverseProxy(parsedURL)
			proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, err error) {
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusBadGateway)
				fmt.Fprintf(writer, `{"error":"bad gateway","message":"%v"}`, err)
			}

			backends = append(backends, &Backend{
				URL:          parsedURL,
				Alive:        true,
				ReverseProxy: proxy,
			})
		}
	}

	return &LoadBalancer{
		backends:            backends,
		healthCheckInterval: healthCheckInterval,
	}
}

func (lb *LoadBalancer) SetStrategy(strategy Strategy) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.strategy = strategy
}

func (lb *LoadBalancer) GetNextBackend() (*Backend, error) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	if len(lb.backends) == 0 {
		return nil, ErrNoBackends
	}

	return lb.strategy.GetNextBackend(lb.backends)
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	backend, err := lb.GetNextBackend()
	if err != nil {
		http.Error(w, "No available backends", http.StatusServiceUnavailable)
		return
	}

	if lc, ok := lb.strategy.(interface{ ReleaseConnection(*Backend) }); ok {
		defer lc.ReleaseConnection(backend)
	}
	backend.ReverseProxy.ServeHTTP(w, r)
}

func (lb *LoadBalancer) StartHealthChecks() {
	ticker := time.NewTicker(lb.healthCheckInterval)
	for range ticker.C {
		lb.HealthCheck()
	}
}

func (lb *LoadBalancer) HealthCheck() {
	for _, backend := range lb.backends {
		status := lb.isBackendAlive(backend.URL)
		backend.SetAlive(status)
		msg := "OK"
		if !status {
			msg = "DOWN"
		}
		log.Printf("Backend %s is %s", backend.URL, msg)
	}
}

func (lb *LoadBalancer) isBackendAlive(u *url.URL) bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", u.Host, timeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func (b *Backend) SetAlive(alive bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Alive = alive
}

func (b *Backend) IsAlive() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Alive
}
