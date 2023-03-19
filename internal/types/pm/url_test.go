package pm

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURLString(t *testing.T) {
	u := PostmanURL{
		Raw: "a-raw-url",
	}

	assert.Equal(t, "a-raw-url", u.String())
}

func TestURLMarshalJSON(t *testing.T) {
	cases := []struct {
		scenario       string
		url            PostmanURL
		expectedOutput string
	}{
		{
			"Successfully marshalling a raw PostmanURL as a struct (v2.1.0)",
			PostmanURL{
				Raw: "http://www.google.fr",
			},
			"{\"raw\":\"http://www.google.fr\"}",
		},
		{
			"Successfully marshalling an PostmanURL with variables as a struct (v2.1.0)",
			PostmanURL{
				Raw: "http://www.google.fr",
				Variables: []*PostmanVariable{
					{
						Name:  "a-variable",
						Value: "an-awesome-value",
					},
				},
			},
			"{\"raw\":\"http://www.google.fr\",\"variable\":[{\"name\":\"a-variable\",\"value\":\"an-awesome-value\"}]}",
		},
	}

	for _, tc := range cases {
		bytes, _ := tc.url.MarshalJSON()

		assert.Equal(t, tc.expectedOutput, string(bytes), tc.scenario)
	}
}

func TestURLUnmarshalJSON(t *testing.T) {
	cases := []struct {
		scenario      string
		bytes         []byte
		expectedURL   PostmanURL
		expectedError error
	}{
		{
			"Successfully unmarshalling an PostmanURL as a string",
			[]byte("\"http://www.google.fr\""),
			PostmanURL{
				Raw: "http://www.google.fr",
			},
			nil,
		},
		{
			"Successfully unmarshalling an PostmanURL with variables as a struct",
			[]byte("{\"raw\":\"http://www.google.fr\",\"variable\":[{\"name\":\"a-variable\",\"value\":\"an-awesome-value\"}]}"),
			PostmanURL{
				Raw: "http://www.google.fr",
				Variables: []*PostmanVariable{
					{
						Name:  "a-variable",
						Value: "an-awesome-value",
					},
				},
			},
			nil,
		},
		{
			"Successfully unmarshalling an PostmanURL with query as a struct",
			[]byte("{\"raw\":\"http://www.google.fr\",\"query\":[{\"key\":\"param1\",\"value\":\"an-awesome-value\"}]}"),
			PostmanURL{
				Raw: "http://www.google.fr",
				Query: []*PostmanQueryParam{
					{
						Key:   "param1",
						Value: "an-awesome-value",
					},
				},
			},
			nil,
		},
		{
			"Failed to unmarshal an PostmanURL because of an unsupported type",
			[]byte("not-a-valid-url"),
			PostmanURL{},
			errors.New("unsupported type"),
		},
	}

	for _, tc := range cases {

		u, err := UnmarshalURLJSON(tc.bytes)

		assert.Equal(t, &tc.expectedURL, u, tc.scenario)
		assert.Equal(t, tc.expectedError, err, tc.scenario)
	}
}
