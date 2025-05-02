package balancer

import (
	"sync"
)

type LeastConnStrategy struct {
	mu    sync.Mutex
	conns map[*Backend]int
}

func NewLeastConnStrategy() *LeastConnStrategy {
	return &LeastConnStrategy{
		conns: make(map[*Backend]int),
	}
}

func (s *LeastConnStrategy) GetNextBackend(backends []*Backend) (*Backend, error) {
	if len(backends) == 0 {
		return nil, ErrNoBackends
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var selected *Backend
	minConns := int(^uint(0) >> 1) // Max int

	for _, backend := range backends {
		if !backend.IsAlive() {
			continue
		}

		conns := s.conns[backend]
		if conns < minConns {
			minConns = conns
			selected = backend
		}
	}

	if selected != nil {
		s.conns[selected]++
	}

	return selected, nil
}

func (s *LeastConnStrategy) ReleaseConnection(backend *Backend) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.conns[backend]--
}
