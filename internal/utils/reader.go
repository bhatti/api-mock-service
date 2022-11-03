package utils

import (
	"bytes"
	"io"
)

type nopCloser struct {
	reader *bytes.Reader
}

// ResetReader is the interface for resetting reader
type ResetReader interface {
	Reset() error
}

// NopCloser returns a ReadCloser with a no-op Close method wrapping
// the provided Reader r.
// If r implements WriterTo, the returned ReadCloser will implement WriterTo
// by forwarding calls to r.
func NopCloser(r *bytes.Reader) io.ReadCloser {
	return &nopCloser{reader: r}
}

// Read bytes
func (nop *nopCloser) Read(p []byte) (n int, err error) {
	return nop.reader.Read(p)
}

// Close NOP operation
func (nop *nopCloser) Close() error {
	return nil
}

// Reset resets pointer of buffer
func (nop *nopCloser) Reset() error {
	_, err := nop.reader.Seek(0, io.SeekStart)
	return err
}

// ReadAll helper method
func ReadAll(body io.ReadCloser) (b []byte, reader io.ReadCloser, err error) {
	if body == nil {
		return nil, body, nil
	}
	b, err = io.ReadAll(body)
	if err != nil {
		return nil, body, err
	}
	_ = body.Close()
	reader = NopCloser(bytes.NewReader(b))
	return
}
