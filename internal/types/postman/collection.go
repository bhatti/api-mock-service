package postman

import (
	"encoding/json"
	"fmt"
	"io"
)

// Info stores data about the collection.
type Info struct {
	Name        string      `json:"name"`
	Description Description `json:"description"`
	Version     string      `json:"version"`
	Schema      string      `json:"schema"`
}

// Collection represents a Postman Collection.
type Collection struct {
	Auth      *Auth       `json:"auth,omitempty"`
	Info      Info        `json:"info"`
	Items     []*Items    `json:"item"`
	Events    []*Event    `json:"event,omitempty"`
	Variables []*Variable `json:"variable,omitempty"`
}

// CreateCollection returns a new Collection.
func CreateCollection(name string, desc string) *Collection {
	return &Collection{
		Info: Info{
			Name: name,
			Description: Description{
				Content: desc,
			},
		},
	}
}

// AddItem appends an item (Item or ItemGroup) to the existing items slice.
func (c *Collection) AddItem(item *Items) {
	c.Items = append(c.Items, item)
}

// AddItemGroup creates a new ItemGroup and appends it to the existing items slice.
func (c *Collection) AddItemGroup(name string) (f *Items) {
	f = &Items{
		Name:  name,
		Items: make([]*Items, 0),
	}

	c.Items = append(c.Items, f)

	return
}

// Write encodes the Collection struct in JSON and writes it into the provided io.Writer.
func (c *Collection) Write(w io.Writer) (err error) {
	version := "v2.1.0"
	c.Info.Schema = fmt.Sprintf("https://schema.getpostman.com/json/collection/%s/collection.json", version)

	file, _ := json.MarshalIndent(c, "", "    ")

	_, err = w.Write(file)

	return
}

// ParseCollection parses the content of the provided data stream into a Collection object.
func ParseCollection(r io.Reader) (c *Collection, err error) {
	err = json.NewDecoder(r).Decode(&c)
	return
}
