package thdone

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDone_Close(t *testing.T) {
	d := New()
	require.True(t, d.Close())
	require.False(t, d.Close())
	require.False(t, d.Close())
	require.False(t, d.Close())
	require.False(t, d.Close())
	select {
	case <-d.C:
	default:
		t.Fatal("expected closed channel")
	}
}
