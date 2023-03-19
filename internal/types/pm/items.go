package pm

import "encoding/json"

// PostmanItems are the basic unit for a Postman collection.
// It can either be a request (PostmanItem) or a folder (PostmanItemGroup).
type PostmanItems struct {
	// Common fields.
	Name                    string             `json:"name"`
	Description             string             `json:"description,omitempty"`
	Variables               []*PostmanVariable `json:"variable,omitempty"`
	Events                  []*PostmanEvent    `json:"event,omitempty"`
	ProtocolProfileBehavior interface{}        `json:"protocolProfileBehavior,omitempty"`
	// Fields specific to PostmanItem
	ID        string             `json:"id,omitempty"`
	Request   *PostmanRequest    `json:"request,omitempty"`
	Responses []*PostmanResponse `json:"response,omitempty"`
	// Fields specific to PostmanItemGroup
	Items []*PostmanItems `json:"item"`
	Auth  *PostmanAuth    `json:"auth,omitempty"`
}

// An PostmanItem is an entity which contain an actual HTTP request, and sample responses attached to it.
type PostmanItem struct {
	Name                    string             `json:"name"`
	Description             string             `json:"description,omitempty"`
	Variables               []*PostmanVariable `json:"variable,omitempty"`
	Events                  []*PostmanEvent    `json:"event,omitempty"`
	ProtocolProfileBehavior interface{}        `json:"protocolProfileBehavior,omitempty"`
	ID                      string             `json:"id,omitempty"`
	Request                 *PostmanRequest    `json:"request,omitempty"`
	Responses               []*PostmanResponse `json:"response,omitempty"`
}

// A PostmanItemGroup is an ordered set of requests.
type PostmanItemGroup struct {
	Name                    string             `json:"name"`
	Description             string             `json:"description,omitempty"`
	Variables               []*PostmanVariable `json:"variable,omitempty"`
	Events                  []*PostmanEvent    `json:"event,omitempty"`
	ProtocolProfileBehavior interface{}        `json:"protocolProfileBehavior,omitempty"`
	Items                   []*PostmanItems    `json:"item"`
	Auth                    *PostmanAuth       `json:"auth,omitempty"`
}

// CreatePostmanItem is a helper to create a new PostmanItem.
func CreatePostmanItem(i PostmanItem) *PostmanItems {
	return &PostmanItems{
		Name:                    i.Name,
		Description:             i.Description,
		Variables:               i.Variables,
		Events:                  i.Events,
		ProtocolProfileBehavior: i.ProtocolProfileBehavior,
		ID:                      i.ID,
		Request:                 i.Request,
		Responses:               i.Responses,
	}
}

// CreatePostmanItemGroup is a helper to create a new PostmanItemGroup.
func CreatePostmanItemGroup(ig PostmanItemGroup) *PostmanItems {
	return &PostmanItems{
		Name:                    ig.Name,
		Description:             ig.Description,
		Variables:               ig.Variables,
		Events:                  ig.Events,
		ProtocolProfileBehavior: ig.ProtocolProfileBehavior,
		Items:                   ig.Items,
		Auth:                    ig.Auth,
	}
}

// IsGroup returns false as an PostmanItem is not a group.
func (i PostmanItems) IsGroup() bool {
	if i.Items != nil {
		return true
	}

	return false
}

// AddItem appends an item to the existing items slice.
func (i *PostmanItems) AddItem(item *PostmanItems) {
	i.Items = append(i.Items, item)
}

// AddItemGroup creates a new PostmanItem folder and appends it to the existing items slice.
func (i *PostmanItems) AddItemGroup(name string) (f *PostmanItems) {
	f = &PostmanItems{
		Name:  name,
		Items: make([]*PostmanItems, 0),
	}

	i.Items = append(i.Items, f)

	return
}

// MarshalJSON returns the JSON encoding of an PostmanItem/PostmanItemGroup.
func (i PostmanItems) MarshalJSON() ([]byte, error) {

	if i.IsGroup() {
		return json.Marshal(PostmanItemGroup{
			Name:                    i.Name,
			Description:             i.Description,
			Variables:               i.Variables,
			Events:                  i.Events,
			ProtocolProfileBehavior: i.ProtocolProfileBehavior,
			Items:                   i.Items,
			Auth:                    i.Auth,
		})
	}

	return json.Marshal(PostmanItem{
		Name:                    i.Name,
		Description:             i.Description,
		Variables:               i.Variables,
		Events:                  i.Events,
		ProtocolProfileBehavior: i.ProtocolProfileBehavior,
		ID:                      i.ID,
		Request:                 i.Request,
		Responses:               i.Responses,
	})
}
