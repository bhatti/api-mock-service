package pm

import (
	"encoding/json"
	"fmt"
	"io"
)

// PostmanInfo stores data about the collection.
type PostmanInfo struct {
	Name        string                       `json:"name"`
	Description PostmanCollectionDescription `json:"description"`
	Version     string                       `json:"version"`
	Schema      string                       `json:"schema"`
}

// PostmanCollection represents a Postman PostmanCollection.
type PostmanCollection struct {
	Auth      *PostmanAuth       `json:"auth,omitempty"`
	Info      PostmanInfo        `json:"info"`
	Items     []*PostmanItems    `json:"item"`
	Events    []*PostmanEvent    `json:"event,omitempty"`
	Variables []*PostmanVariable `json:"variable,omitempty"`
}

// CreateCollection returns a new PostmanCollection.
func CreateCollection(name string, desc string) *PostmanCollection {
	return &PostmanCollection{
		Info: PostmanInfo{
			Name: name,
			Description: PostmanCollectionDescription{
				Content: desc,
			},
		},
	}
}

// AddItem appends an item (PostmanItem or PostmanItemGroup) to the existing items slice.
func (c *PostmanCollection) AddItem(item *PostmanItems) {
	c.Items = append(c.Items, item)
}

// AddItemGroup creates a new PostmanItemGroup and appends it to the existing items slice.
func (c *PostmanCollection) AddItemGroup(name string) (f *PostmanItems) {
	f = &PostmanItems{
		Name:  name,
		Items: make([]*PostmanItems, 0),
	}

	c.Items = append(c.Items, f)

	return
}

// Write encodes the PostmanCollection struct in JSON and writes it into the provided io.Writer.
func (c *PostmanCollection) Write(w io.Writer) (err error) {
	version := "v2.1.0"
	c.Info.Schema = fmt.Sprintf("https://schema.getpostman.com/json/collection/%s/collection.json", version)

	file, _ := json.MarshalIndent(c, "", "    ")

	_, err = w.Write(file)

	return
}

// ParseCollection parses the content of the provided data stream into a PostmanCollection object.
func ParseCollection(r io.Reader) (c *PostmanCollection, err error) {
	err = json.NewDecoder(r).Decode(&c)
	return
}
