package pm

import (
	"encoding/json"
	"errors"
	"fmt"
)

// PostmanCollectionDescription contains collection description.
type PostmanCollectionDescription struct {
	Content string `json:"content,omitempty"`
	Type    string `json:"type,omitempty"`
	Version string `json:"version,omitempty"`
}

// mDescription is used for marshalling/unmarshalling.
type mDescription PostmanCollectionDescription

// MarshalJSON returns the JSON encoding of a PostmanCollectionDescription.
// If the PostmanCollectionDescription only has a content, it is returned as a string.
func (d PostmanCollectionDescription) MarshalJSON() ([]byte, error) {
	if d.Type == "" && d.Version == "" {
		return []byte(fmt.Sprintf("\"%s\"", d.Content)), nil
	}

	return json.Marshal(mDescription{
		Content: d.Content,
		Type:    d.Type,
		Version: d.Version,
	})
}

// UnmarshalJSON parses the JSON-encoded data and create a PostmanCollectionDescription from it.
// A PostmanCollectionDescription can be created from an object or a string.
func (d *PostmanCollectionDescription) UnmarshalJSON(b []byte) (err error) {
	if len(b) == 0 {
		return nil
	} else if len(b) >= 2 && b[0] == '"' && b[len(b)-1] == '"' {
		d.Content = string(string(b[1 : len(b)-1]))
	} else if len(b) >= 2 && b[0] == '{' && b[len(b)-1] == '}' {
		tmp := (*mDescription)(d)
		err = json.Unmarshal(b, &tmp)
	} else {
		err = errors.New("unsupported type for description")
	}

	return
}
