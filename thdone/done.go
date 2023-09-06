package thdone

import "sync"

type Done struct {
	C    chan struct{}
	once *sync.Once
}

func (d Done) Close() {
	d.once.Do(func() {
		close(d.C)
	})
}

func New() Done {
	return Done{
		C:    make(chan struct{}),
		once: &sync.Once{},
	}
}
