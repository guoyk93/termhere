package thwire

import (
	"bytes"
	"encoding/gob"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreateValidateAuthFrame(t *testing.T) {
	var f Frame
	err := CreateAuthFrame(&f, "hello")
	require.NoError(t, err)

	buf := &bytes.Buffer{}

	err = gob.NewEncoder(buf).Encode(f)
	require.NoError(t, err)

	var f2 Frame
	err = gob.NewDecoder(bytes.NewReader(buf.Bytes())).Decode(&f2)
	require.NoError(t, err)

	err = ValidateAuthFrame(f2, "hello")
	require.Error(t, err)
	require.Equal(t, ErrInvalidNonce, err)

	clear(nonceCache)

	err = ValidateAuthFrame(f2, "hello")
	require.NoError(t, err)

	err = ValidateAuthFrame(f2, "hello")
	require.Error(t, err)
	require.Equal(t, ErrInvalidNonce, err)
}
