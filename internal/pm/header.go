package pm

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// PostmanHeader represents an HTTP PostmanHeader.
type PostmanHeader struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Disabled    bool   `json:"disabled,omitempty"`
	Description string `json:"description,omitempty"`
}

// PostmanHeaderList contains a list of headers.
type PostmanHeaderList struct {
	Headers []*PostmanHeader
}

func buildHeaders(kv map[string]string) (res []*PostmanHeader) {
	for k, v := range kv {
		res = append(res, &PostmanHeader{
			Key:   k,
			Value: v,
		})
	}
	return
}

// buildHeadersArray converts header map to Postman header array
func buildHeadersArray(headers http.Header) []*PostmanHeader {
	var result []*PostmanHeader
	for key, values := range headers {
		if len(values) > 0 {
			result = append(result, &PostmanHeader{
				Key:   key,
				Value: headers.Get(key),
			})
		}
	}
	return result
}

// MarshalJSON returns the JSON encoding of a PostmanHeaderList.
func (hl PostmanHeaderList) MarshalJSON() ([]byte, error) {
	return json.Marshal(hl.Headers)
}

// UnmarshalJSON parses the JSON-encoded data and create a PostmanHeaderList from it.
// A PostmanHeaderList can be created from an array or a string.
func (hl *PostmanHeaderList) UnmarshalJSON(b []byte) (err error) {
	if len(b) == 0 {
		return nil
	} else if len(b) >= 2 && b[0] == '"' && b[len(b)-1] == '"' {
		headersString := string(b[1 : len(b)-1])
		for _, header := range strings.Split(headersString, "\n") {
			if strings.TrimSpace(header) == "" {
				continue
			}

			headerParts := strings.Split(header, ":")

			if len(headerParts) != 2 {
				return fmt.Errorf("invalid header, missing key or value: %s", header)
			}

			hl.Headers = append(hl.Headers, &PostmanHeader{
				Key:   strings.TrimSpace(headerParts[0]),
				Value: strings.TrimSpace(string(headerParts[1])),
			})
		}
	} else if len(b) >= 2 && b[0] == '[' && b[len(b)-1] == ']' {
		err = json.Unmarshal(b, &hl.Headers)
	} else {
		err = errors.New("unsupported type for header list")
	}

	return
}
