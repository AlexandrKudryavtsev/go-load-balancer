package balancer

import "sync"

type Balancer struct {
	pool *Pool
	mu   sync.Mutex
}

func New(pool *Pool) *Balancer {
	return &Balancer{
		pool: pool,
		mu:   sync.Mutex{},
	}
}

func (b *Balancer) Next() *Backend {
	b.mu.Lock()
	defer b.mu.Unlock()
	var best *Backend
	totalWeight := 0

	for _, backend := range b.pool.Backends {
		if !backend.Alive.Load() {
			continue
		}

		backend.CurrentWeight += backend.Weight
		totalWeight += backend.Weight
		if best == nil || backend.CurrentWeight > best.CurrentWeight {
			best = backend
		}
	}

	if best == nil {
		return nil
	}

	best.CurrentWeight -= totalWeight

	return best
}
