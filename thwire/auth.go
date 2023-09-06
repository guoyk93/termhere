package thwire

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"sync"
	"time"
)

var (
	ErrInvalidEpoch     = errors.New("auth failed: invalid epoch")
	ErrInvalidNonce     = errors.New("auth failed: invalid nonce")
	ErrInvalidSignature = errors.New("auth failed: invalid signature")
)

var (
	nonceCache                 = make(map[uint64]struct{})
	nonceCacheLock sync.Locker = &sync.Mutex{}
)

func checkNonce(nonce uint64) (ok bool) {
	nonceCacheLock.Lock()
	defer nonceCacheLock.Unlock()

	if len(nonceCache) > 1000 {
		nonceCache = make(map[uint64]struct{})
	}

	if _, found := nonceCache[nonce]; found {
		return
	}

	nonceCache[nonce] = struct{}{}
	ok = true

	return
}

func calculateSignature(epoch uint64, nonce uint64, token string) []byte {
	var buf []byte
	buf = binary.BigEndian.AppendUint64(buf, epoch)
	buf = binary.BigEndian.AppendUint64(buf, nonce)
	buf = append(buf, token...)
	sum := sha256.Sum224(buf)
	return sum[:]
}

// CreateAuthFrame creates an auth frame.
func CreateAuthFrame(f *Frame, token string) (err error) {
	var nonce [8]byte
	if _, err = rand.Read(nonce[:]); err != nil {
		return
	}

	f.Kind = KindAuth
	f.Auth.Epoch = uint64(time.Now().Unix())
	f.Auth.Nonce = binary.BigEndian.Uint64(nonce[:])
	f.Auth.Signature = calculateSignature(f.Auth.Epoch, f.Auth.Nonce, token)

	// add nonce
	_ = checkNonce(f.Auth.Nonce)
	return
}

// ValidateAuthFrame validates an auth frame.
func ValidateAuthFrame(f Frame, token string) (err error) {
	if f.Kind != KindAuth {
		err = ErrInvalidFrame
		return
	}
	diff := time.Now().Unix() - int64(f.Auth.Epoch)
	if diff > 60 || diff < -60 {
		err = ErrInvalidEpoch
		return
	}
	if !checkNonce(f.Auth.Nonce) {
		err = ErrInvalidNonce
		return
	}
	if !bytes.Equal(
		calculateSignature(f.Auth.Epoch, f.Auth.Nonce, token),
		f.Auth.Signature,
	) {
		err = ErrInvalidSignature
	}
	return
}
