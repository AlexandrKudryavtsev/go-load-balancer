package balancer

import (
	"sync/atomic"
)

type Balancer struct {
	pool    *Pool
	current uint64
}

func New(pool *Pool) *Balancer {
	return &Balancer{
		pool:    pool,
		current: 0,
	}
}

func (b *Balancer) Next() *Backend {
	n := len(b.pool.Backends)
	if n == 0 {
		return nil
	}

	current := atomic.AddUint64(&b.current, 1)

	return b.pool.Backends[(current-1)%uint64(n)]
}
