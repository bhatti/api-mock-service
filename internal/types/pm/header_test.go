package pm

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeaderListMarshalJSON(t *testing.T) {
	cases := []struct {
		scenario       string
		headerList     PostmanHeaderList
		expectedOutput string
	}{
		{
			"Successfully marshalling a PostmanHeaderList",
			PostmanHeaderList{
				Headers: []*PostmanHeader{
					{
						Key:         "Content-Type",
						Value:       "application/json",
						Description: "The content type",
					},
					{
						Key:         "Authorization",
						Value:       "Bearer a-bearer-token",
						Description: "The Bearer token",
					},
				},
			},
			"[{\"key\":\"Content-Type\",\"value\":\"application/json\",\"description\":\"The content type\"},{\"key\":\"Authorization\",\"value\":\"Bearer a-bearer-token\",\"description\":\"The Bearer token\"}]",
		},
	}

	for _, tc := range cases {
		bytes, _ := tc.headerList.MarshalJSON()

		assert.Equal(t, tc.expectedOutput, string(bytes), tc.scenario)
	}
}

func TestHeaderListUnmarshalJSON(t *testing.T) {
	cases := []struct {
		scenario           string
		bytes              []byte
		expectedHeaderList PostmanHeaderList
		expectedError      error
	}{
		{
			"Successfully unmarshalling a PostmanHeaderList from a string",
			[]byte("\"Content-Type: application/json\nAuthorization: Bearer a-bearer-token\n\""),
			PostmanHeaderList{
				Headers: []*PostmanHeader{
					{
						Key:   "Content-Type",
						Value: "application/json",
					},
					{
						Key:   "Authorization",
						Value: "Bearer a-bearer-token",
					},
				},
			},
			nil,
		},
		{
			"Successfully unmarshalling a PostmanHeaderList from an empty slice of bytes",
			make([]byte, 0),
			PostmanHeaderList{},
			nil,
		},
		{
			"Successfully unmarshalling a PostmanHeaderList from an array of objects",
			[]byte("[{\"key\": \"Content-Type\",\"value\": \"application\\/json\",\"description\": \"The content type\"},{\"key\": \"Authorization\",\"value\": \"Bearer a-bearer-token\",\"description\": \"The Bearer token\"}]"),
			PostmanHeaderList{
				Headers: []*PostmanHeader{
					{
						Key:         "Content-Type",
						Value:       "application/json",
						Description: "The content type",
					},
					{
						Key:         "Authorization",
						Value:       "Bearer a-bearer-token",
						Description: "The Bearer token",
					},
				},
			},
			nil,
		},
		{
			"Failed to unmarshal a PostmanHeaderList because of an invalid header",
			[]byte("\"Content-Type\n\""),
			PostmanHeaderList{},
			errors.New("invalid header, missing key or value: Content-Type"),
		},
		{
			"Failed to unmarshal a PostmanHeaderList because of an unsupported type",
			[]byte("not-a-valid-header-list"),
			PostmanHeaderList{},
			errors.New("unsupported type for header list"),
		},
	}

	for _, tc := range cases {

		var hl PostmanHeaderList
		err := hl.UnmarshalJSON(tc.bytes)

		assert.Equal(t, tc.expectedHeaderList, hl, tc.scenario)
		assert.Equal(t, tc.expectedError, err, tc.scenario)
	}
}
