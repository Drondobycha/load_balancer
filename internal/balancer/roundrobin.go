package balancer

type RoundRobinStrategy struct {
	counter int
}

func (r *RoundRobinStrategy) GetNextBackend(backends []*Backend) (*Backend, error) {
	nextIndex := (r.counter + 1) % len(backends)
	r.counter = nextIndex

	// Ищем живой бэкенд
	for i := range backends {
		idx := (nextIndex + i) % len(backends)
		if backends[idx].IsAlive() {
			r.counter = idx
			return backends[idx], nil
		}
	}

	return nil, ErrNoBackends
}
