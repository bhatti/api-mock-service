package postman

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestMarshalJSON(t *testing.T) {
	cases := []struct {
		scenario       string
		req            Request
		expectedOutput string
	}{
		{
			"Successfully marshalling a Request as an object (v2.1.0)",
			Request{
				Method: Post,
				URL: &URL{
					Raw: "http://www.google.fr",
				},
				Body: &Body{
					Mode: "raw",
					Raw:  "raw-content",
				},
			},
			"{\"url\":{\"raw\":\"http://www.google.fr\"},\"method\":\"POST\",\"body\":{\"mode\":\"raw\",\"raw\":\"raw-content\"}}",
		},
	}

	for _, tc := range cases {
		bytes, _ := tc.req.MarshalJSON()

		assert.Equal(t, tc.expectedOutput, string(bytes), tc.scenario)
	}
}

func TestRequestUnmarshalJSON(t *testing.T) {
	cases := []struct {
		scenario        string
		bytes           []byte
		expectedRequest Request
		expectedError   error
	}{
		{
			"Successfully unmarshalling a Request as a string",
			[]byte("\"http://www.google.fr\""),
			Request{
				Method: Get,
				URL: &URL{
					Raw: "http://www.google.fr",
				},
			},
			nil,
		},
		{
			"Successfully unmarshalling a Request URL with some content",
			[]byte("{\"url\":{\"raw\": \"http://www.google.fr\"},\"body\":{\"mode\":\"raw\",\"raw\":\"awesome-body\"}}"),
			Request{
				URL: &URL{
					Raw: "http://www.google.fr",
				},
				Body: &Body{
					Mode: "raw",
					Raw:  "awesome-body",
				},
			},
			nil,
		},
		{
			"Failed to unmarshal a Request because of an unsupported type",
			[]byte("not-a-valid-request"),
			Request{},
			errors.New("unsupported type"),
		},
	}

	for _, tc := range cases {
		var r Request
		err := r.UnmarshalJSON(tc.bytes)
		assert.Equal(t, tc.expectedRequest, r, tc.scenario)
		assert.Equal(t, tc.expectedError, err, tc.scenario)
	}
}
