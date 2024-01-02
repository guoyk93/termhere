package chdone

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDone_TryClose(t *testing.T) {
	d := New()
	require.True(t, d.TryClose())
	require.False(t, d.TryClose())

	select {
	case <-d.C:
	default:
		require.Fail(t, "channel should be closed")
	}
}

func TestDone_Close(t *testing.T) {
	d := New()
	require.NoError(t, d.Close())
	require.Equal(t, ErrAlreadyClosed, d.Close())

	select {
	case <-d.C:
	default:
		require.Fail(t, "channel should be closed")
	}
}
