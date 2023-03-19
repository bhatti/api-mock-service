package pm

import (
	"encoding/json"
	"errors"
	"net/url"
)

// PostmanURL is a struct that contains an PostmanURL in a "broken-down way".
// Raw contains the complete PostmanURL.
type PostmanURL struct {
	Raw       string               `json:"raw"`
	Protocol  string               `json:"protocol,omitempty"`
	Host      []string             `json:"host,omitempty"`
	Path      []string             `json:"path,omitempty"`
	Port      string               `json:"port,omitempty"`
	Query     []*PostmanQueryParam `json:"query,omitempty"`
	Hash      string               `json:"hash,omitempty"`
	Variables []*PostmanVariable   `json:"variable,omitempty" mapstructure:"variable"`
}

// PostmanQueryParam param for query
type PostmanQueryParam struct {
	Key         string  `json:"key"`
	Value       string  `json:"value"`
	Description *string `json:"description"`
}

func buildURL(u *url.URL) (res *PostmanURL) {
	res = &PostmanURL{
		Raw:      u.String(),
		Protocol: u.Scheme,
		Host:     []string{u.Host},
		Path:     []string{u.Path},
		Port:     u.Port(),
	}
	for k, v := range u.Query() {
		res.Query = append(res.Query, &PostmanQueryParam{
			Key:   k,
			Value: v[0],
		})
	}
	return
}

// String returns the raw version of the PostmanURL.
func (u *PostmanURL) String() string {
	return u.Raw
}

// MarshalJSON returns the JSON encoding of an PostmanURL.
// It encodes the PostmanURL as a string if it does not contain any variable.
// In case it contains any variable, it gets encoded as an object.
func (u *PostmanURL) MarshalJSON() ([]byte, error) {
	return json.Marshal(PostmanURL{
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

// UnmarshalURLJSON parses the JSON-encoded data and create an PostmanURL from it.
// An PostmanURL can be created from an object or a string.
// If a string, the value is assumed to be the Raw attribute of the PostmanURL.
func UnmarshalURLJSON(b []byte) (u *PostmanURL, err error) {
	u = &PostmanURL{}
	if b[0] == '"' {
		u.Raw = string(b[1 : len(b)-1])
	} else if b[0] == '{' {
		err = json.Unmarshal(b, u)
	} else {
		err = errors.New("unsupported type")
	}

	return
}
