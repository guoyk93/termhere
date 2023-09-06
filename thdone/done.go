package thdone

import "sync"

// Done is a channel that can be closed only once.
type Done struct {
	C    chan struct{}
	once *sync.Once
}

// Close closes the channel if it has not been closed yet.
// It returns true if the channel was closed for first time.
func (d Done) Close() (ok bool) {
	d.once.Do(func() {
		ok = true
		close(d.C)
	})
	return
}

func New() Done {
	return Done{
		C:    make(chan struct{}),
		once: &sync.Once{},
	}
}
