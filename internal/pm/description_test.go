package pm

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDescriptionMarshalJSON(t *testing.T) {
	cases := []struct {
		scenario       string
		description    PostmanCollectionDescription
		expectedOutput string
	}{
		{
			"Successfully marshalling a PostmanCollectionDescription as an object",
			PostmanCollectionDescription{
				Content: "My awesome collection",
				Type:    "text/plain",
				Version: "v1",
			},
			`{"content":"My awesome collection","type":"text/plain","version":"v1"}`,
		},
		{
			"Successfully marshalling a PostmanCollectionDescription as a string",
			PostmanCollectionDescription{
				Content: "My awesome collection",
			},
			`"My awesome collection"`,
		},
	}

	for _, tc := range cases {
		bytes, _ := tc.description.MarshalJSON()

		assert.Equal(t, tc.expectedOutput, string(bytes), tc.scenario)
	}
}

func TestDescriptionUnmarshalJSON(t *testing.T) {
	cases := []struct {
		scenario            string
		bytes               []byte
		expectedDescription PostmanCollectionDescription
		expectedError       error
	}{
		{
			"Successfully unmarshalling a PostmanCollectionDescription from a string",
			[]byte(`"My awesome collection"`),
			PostmanCollectionDescription{Content: "My awesome collection"},
			nil,
		},
		{
			"Successfully unmarshalling a PostmanCollectionDescription from an empty slice of bytes",
			make([]byte, 0),
			PostmanCollectionDescription{},
			nil,
		},
		{
			"Successfully unmarshalling a PostmanCollectionDescription",
			[]byte(`{"content":"My awesome collection","type":"text/plain","version":"v1"}`),
			PostmanCollectionDescription{
				Content: "My awesome collection",
				Type:    "text/plain",
				Version: "v1",
			},
			nil,
		},
		{
			"Failed to unmarshal a PostmanCollectionDescription because of an unsupported type",
			[]byte(`not-a-valid-description`),
			PostmanCollectionDescription{},
			errors.New("unsupported type for description"),
		},
	}

	for _, tc := range cases {

		var d PostmanCollectionDescription
		err := d.UnmarshalJSON(tc.bytes)

		assert.Equal(t, tc.expectedDescription, d, tc.scenario)
		assert.Equal(t, tc.expectedError, err, tc.scenario)
	}
}
