package pm

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/types"
	"net/url"
)

// A PostmanRequest represents an HTTP request.
type PostmanRequest struct {
	URL         *PostmanURL         `json:"url"`
	Auth        *PostmanAuth        `json:"auth,omitempty"`
	Proxy       interface{}         `json:"proxy,omitempty"`
	Certificate interface{}         `json:"certificate,omitempty"`
	Method      types.MethodType    `json:"method"`
	Description interface{}         `json:"description,omitempty"`
	Header      []*PostmanHeader    `json:"header,omitempty"`
	Body        *PostmanRequestBody `json:"body,omitempty"`
}

// mRequest is used for marshalling/unmarshalling.
type mRequest PostmanRequest

func (r *PostmanRequest) bodyText() []byte {
	if r.Body != nil {
		return []byte(replaceTemplateVariables(r.Body.Raw))
	}
	return nil
}

func (r *PostmanRequest) formParams() (res map[string][]string) {
	if r.Body != nil && r.Body.FormData != nil {
		res, _ = url.ParseQuery(fmt.Sprintf("%v", r.Body.FormData))
	}
	if res == nil {
		res = make(map[string][]string)
	}
	return res
}

func (r *PostmanRequest) contentType() string {
	for _, header := range r.Header {
		if header.Key == types.ContentTypeHeader {
			return header.Value
		}
	}
	return ""
}

func (r *PostmanRequest) headersMap() (res map[string][]string) {
	res = make(map[string][]string)
	for _, header := range r.Header {
		if !header.Disabled {
			res[header.Key] = res[header.Value]
		}
	}
	return
}

// MarshalJSON returns the JSON encoding of a PostmanRequest.
// If the PostmanRequest only contains an PostmanURL with the Get HTTP method, it is returned as a string.
func (r *PostmanRequest) MarshalJSON() ([]byte, error) {
	if r.Auth == nil && r.Proxy == nil && r.Certificate == nil && r.Description == nil && r.Header == nil && r.Body == nil &&
		r.Method == types.Get {
		return []byte(fmt.Sprintf("\"%s\"", r.URL)), nil
	}

	return json.MarshalIndent(PostmanRequest{
		URL:         r.URL,
		Auth:        r.Auth,
		Proxy:       r.Proxy,
		Certificate: r.Certificate,
		Method:      r.Method,
		Description: r.Description,
		Header:      r.Header,
		Body:        r.Body,
	}, "", "  ")
}

// UnmarshalJSON parses the JSON-encoded data and create a PostmanRequest from it.
// A PostmanRequest can be created from an object or a string.
// If a string, the string is assumed to be the request PostmanURL and the method is assumed to be 'GET'.
func (r *PostmanRequest) UnmarshalJSON(b []byte) (err error) {
	if b[0] == '"' {
		r.Method = types.Get
		r.URL = &PostmanURL{
			Raw: string(b[1 : len(b)-1]),
		}
	} else if b[0] == '{' {
		tmp := (*mRequest)(r)
		err = json.Unmarshal(b, &tmp)
	} else {
		err = errors.New("unsupported type")
	}

	return
}
