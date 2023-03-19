package postman

import (
	"encoding/json"
	"errors"
	"net/url"
)

// URL is a struct that contains an URL in a "broken-down way".
// Raw contains the complete URL.
type URL struct {
	Raw       string        `json:"raw"`
	Protocol  string        `json:"protocol,omitempty"`
	Host      []string      `json:"host,omitempty"`
	Path      []string      `json:"path,omitempty"`
	Port      string        `json:"port,omitempty"`
	Query     []*QueryParam `json:"query,omitempty"`
	Hash      string        `json:"hash,omitempty"`
	Variables []*Variable   `json:"variable,omitempty" mapstructure:"variable"`
}

type QueryParam struct {
	Key         string  `json:"key"`
	Value       string  `json:"value"`
	Description *string `json:"description"`
}

func buildURL(u *url.URL) (res *URL) {
	res = &URL{
		Raw:      u.String(),
		Protocol: u.Scheme,
		Host:     []string{u.Host},
		Path:     []string{u.Path},
		Port:     u.Port(),
	}
	for k, v := range u.Query() {
		res.Query = append(res.Query, &QueryParam{
			Key:   k,
			Value: v[0],
		})
	}
	return
}

// String returns the raw version of the URL.
func (u *URL) String() string {
	return u.Raw
}

// MarshalJSON returns the JSON encoding of an URL.
// It encodes the URL as a string if it does not contain any variable.
// In case it contains any variable, it gets encoded as an object.
func (u *URL) MarshalJSON() ([]byte, error) {
	return json.Marshal(URL{
		Raw:       u.Raw,
		Protocol:  u.Protocol,
		Host:      u.Host,
		Path:      u.Path,
		Port:      u.Port,
		Query:     u.Query,
		Hash:      u.Hash,
		Variables: u.Variables,
	})
}

// UnmarshalURLJSON parses the JSON-encoded data and create an URL from it.
// An URL can be created from an object or a string.
// If a string, the value is assumed to be the Raw attribute of the URL.
func UnmarshalURLJSON(b []byte) (u *URL, err error) {
	u = &URL{}
	if b[0] == '"' {
		u.Raw = string(b[1 : len(b)-1])
	} else if b[0] == '{' {
		err = json.Unmarshal(b, u)
	} else {
		err = errors.New("unsupported type")
	}

	return
}
