package utils

import (
	"bytes"
	"io"
)

// ReadAll helper method
func ReadAll(body io.ReadCloser) (b []byte, copy io.ReadCloser, err error) {
	if body == nil {
		return nil, body, nil
	}
	b, err = io.ReadAll(body)
	if err != nil {
		return nil, body, err
	}
	_ = body.Close()
	body = io.NopCloser(bytes.NewReader(b))
	return
}
