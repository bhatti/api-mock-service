package utils

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"io"
	"testing"
)

func Test_ShouldNotReadAllWithNilBody(t *testing.T) {
	// GIVEN reader
	// WHEN reading with nil closer
	b, r, err := ReadAll(nil)
	// THEN it should return nil
	require.Nil(t, b)
	require.Nil(t, r)
	require.Nil(t, err)
}

func Test_ShouldReadAllAgain(t *testing.T) {
	// GIVEN a string
	str := "test data"
	// WHEN reading all
	b, r, err := ReadAll(io.NopCloser(bytes.NewReader([]byte(str))))
	// THEN it should return byte data without error
	require.NoError(t, err)
	require.Equal(t, "test data", string(b))
	b, r, err = ReadAll(r)
	require.NoError(t, err)
	require.Equal(t, "test data", string(b))
}

func Test_ShouldReadAll(t *testing.T) {
	// GIVEN a string
	str := "test data"
	// WHEN reading all
	b, r, err := ReadAll(io.NopCloser(bytes.NewReader([]byte(str))))
	// THEN it should return byte data without error
	require.NoError(t, err)
	require.Equal(t, "test data", string(b))

	// Reading it again should not return any data
	n, err := r.Read(make([]byte, 0))
	require.NoError(t, err)
	require.Equal(t, 0, n)
	r.(ResetReader).Reset()
	n, err = r.Read(b)
	require.NoError(t, err)
	require.Equal(t, len(str), n)
	r.Close()
}
