package balancer

import "errors"

var ErrNoBackends = errors.New("no available backends")

type Strategy interface {
	GetNextBackend(backends []*Backend) (*Backend, error)
}
