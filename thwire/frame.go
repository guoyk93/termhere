package thwire

import (
	"errors"
)

// Kind is the type of a frame.
type Kind byte

const (
	KindAuth Kind = iota + 1
	KindStdin
	KindStdout
	KindStderr // not used
	KindError
	KindIdle
	KindSignal
	KindResize
)

func (k Kind) String() string {
	if name, ok := map[Kind]string{
		KindAuth:   "auth",
		KindStdin:  "stdin",
		KindStdout: "stdout",
		KindStderr: "stderr",
		KindError:  "error",
		KindIdle:   "idle",
		KindSignal: "signal",
		KindResize: "resize",
	}[k]; ok {
		return name
	}
	return "unknown"
}

var (
	ErrInvalidFrame = errors.New("invalid frame")
)

// FrameAuth is a signal sent over the wire.
type FrameAuth struct {
	Epoch     uint64
	Nonce     uint64
	Signature []byte
}

// FrameResize is a signal sent over the wire.
type FrameResize struct {
	Rows uint16
	Cols uint16
	X    uint16
	Y    uint16
}

// FrameSignal is a signal sent over the wire.
type FrameSignal struct {
	Number int
}

// Frame is a single message sent over the wire.
type Frame struct {
	Kind Kind

	Auth   FrameAuth
	Signal FrameSignal
	Resize FrameResize

	Data []byte
}
