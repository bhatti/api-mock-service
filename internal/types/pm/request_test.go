package pm

import (
	"errors"
	"github.com/bhatti/api-mock-service/internal/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestMarshalJSON(t *testing.T) {
	cases := []struct {
		scenario       string
		req            PostmanRequest
		expectedOutput string
	}{
		{
			"Successfully marshalling a PostmanRequest as an object (v2.1.0)",
			PostmanRequest{
				Method: types.Post,
				URL: &PostmanURL{
					Raw: "http://www.google.fr",
				},
				Body: &PostmanRequestBody{
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
		expectedRequest PostmanRequest
		expectedError   error
	}{
		{
			"Successfully unmarshalling a PostmanRequest as a string",
			[]byte("\"http://www.google.fr\""),
			PostmanRequest{
				Method: types.Get,
				URL: &PostmanURL{
					Raw: "http://www.google.fr",
				},
			},
			nil,
		},
		{
			"Successfully unmarshalling a PostmanRequest PostmanURL with some content",
			[]byte("{\"url\":{\"raw\": \"http://www.google.fr\"},\"body\":{\"mode\":\"raw\",\"raw\":\"awesome-body\"}}"),
			PostmanRequest{
				URL: &PostmanURL{
					Raw: "http://www.google.fr",
				},
				Body: &PostmanRequestBody{
					Mode: "raw",
					Raw:  "awesome-body",
				},
			},
			nil,
		},
		{
			"Failed to unmarshal a PostmanRequest because of an unsupported type",
			[]byte("not-a-valid-request"),
			PostmanRequest{},
			errors.New("unsupported type"),
		},
	}

	for _, tc := range cases {
		var r PostmanRequest
		err := r.UnmarshalJSON(tc.bytes)
		assert.Equal(t, tc.expectedRequest, r, tc.scenario)
		assert.Equal(t, tc.expectedError, err, tc.scenario)
	}
}
