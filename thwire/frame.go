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
	KindExit
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
		KindExit:   "exit",
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
	Env       map[string]string
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

type FrameExit struct {
	Code    int
	Message []byte
}

// Frame is a single message sent over the wire.
type Frame struct {
	Kind Kind

	Auth   FrameAuth
	Signal FrameSignal
	Resize FrameResize
	Exit   FrameExit

	Data []byte
}
