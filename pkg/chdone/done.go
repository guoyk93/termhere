package chdone

import (
	"errors"
	"sync"
)

var (
	// ErrAlreadyClosed is returned when the channel is already closed.
	ErrAlreadyClosed = errors.New("channel already closed")
)

// Done is a channel that can close multiple times without panicking.
type Done struct {
	// C is the channel that will be closed
	C chan struct{}

	once *sync.Once
}

// TryClose closes the channel if it has not been closed yet.
// It returns true if the channel was closed for first time.
func (d *Done) TryClose() (ok bool) {
	d.once.Do(func() {
		ok = true
		close(d.C)
	})
	return
}

// Close implements io.Closer.
func (d *Done) Close() error {
	if d.TryClose() {
		return nil
	}
	return ErrAlreadyClosed
}

// New creates a new Done channel.
func New() *Done {
	return &Done{
		C:    make(chan struct{}),
		once: &sync.Once{},
	}
}
