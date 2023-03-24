package pm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsGroup(t *testing.T) {

	cases := []struct {
		scenario        string
		item            PostmanItems
		expectedIsGroup bool
	}{
		{
			"An item with PostmanItems is a group",
			PostmanItems{
				Name:  "a-name",
				Items: make([]*PostmanItems, 0),
			},
			true,
		},
		{
			"An item without PostmanItems is not a group",
			PostmanItems{
				Name: "a-name",
			},
			false,
		},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.expectedIsGroup, tc.item.IsGroup(), tc.scenario)
	}
}

func TestAddItem(t *testing.T) {
	itemGroup := PostmanItems{
		Name:  "A group of items",
		Items: make([]*PostmanItems, 0),
	}

	itemGroup.AddItem(&PostmanItems{
		Name: "A basic item",
	})

	itemGroup.AddItem(&PostmanItems{
		Name:  "A basic group item",
		Items: make([]*PostmanItems, 0),
	})

	assert.Equal(
		t,
		PostmanItems{
			Name: "A group of items",
			Items: []*PostmanItems{
				{
					Name: "A basic item",
				},
				{
					Name:  "A basic group item",
					Items: make([]*PostmanItems, 0),
				},
			},
		},
		itemGroup,
	)
}

func TestAddItemGroup(t *testing.T) {
	itemGroup := PostmanItems{
		Name:  "A group of items",
		Items: make([]*PostmanItems, 0),
	}

	itemGroup.AddItemGroup("an-item-group")
	itemGroup.AddItemGroup("another-item-group")

	assert.Equal(
		t,
		PostmanItems{
			Name: "A group of items",
			Items: []*PostmanItems{
				{
					Name:  "an-item-group",
					Items: make([]*PostmanItems, 0),
				},
				{
					Name:  "another-item-group",
					Items: make([]*PostmanItems, 0),
				},
			},
		},
		itemGroup,
	)
}

func TestItemsMarshalJSON(t *testing.T) {
	cases := []struct {
		scenario       string
		item           PostmanItems
		expectedOutput string
	}{
		{
			"Successfully marshalling an PostmanItem",
			PostmanItems{
				ID:   "a-unique-id",
				Name: "an-item",
			},
			`{
  "name": "an-item",
  "id": "a-unique-id"
}`,
		},
		{
			"Successfully marshalling a GroupItem",
			PostmanItems{
				Name:  "a-group-item",
				Items: make([]*PostmanItems, 0),
			},
			`{
  "name": "a-group-item",
  "item": []
}`,
		},
	}

	for _, tc := range cases {
		bytes, _ := tc.item.MarshalJSON()

		assert.Equal(t, tc.expectedOutput, string(bytes), tc.scenario)
	}
}

func TestCreateItem(t *testing.T) {
	c := CreatePostmanItem(PostmanItem{
		Name:        "An item",
		Description: "A description",
		Variables: []*PostmanVariable{
			{
				Name:  "variable-name",
				Value: "variable-value",
			},
		},
		Events: []*PostmanEvent{
			{
				Listen: PreRequest,
				Script: &PostmanScript{
					Type: "text/javascript",
					Exec: []string{"console.log(\"foo\")"},
				},
			},
			{
				Listen: Test,
				Script: &PostmanScript{
					Type: "text/javascript",
					Exec: []string{"console.log(\"bar\")"},
				},
			},
		},
		ProtocolProfileBehavior: "a-protocol-profile-behavior",
		ID:                      "an-id",
		Request: &PostmanRequest{
			URL: &PostmanURL{
				Raw: "http://www.google.fr",
			},
		},
		Responses: []*PostmanResponse{
			{
				Name: "a-response",
			},
		},
	})

	assert.Equal(
		t,
		&PostmanItems{
			Name:        "An item",
			Description: "A description",
			Variables: []*PostmanVariable{
				{
					Name:  "variable-name",
					Value: "variable-value",
				},
			},
			Events: []*PostmanEvent{
				{
					Listen: PreRequest,
					Script: &PostmanScript{
						Type: "text/javascript",
						Exec: []string{"console.log(\"foo\")"},
					},
				},
				{
					Listen: Test,
					Script: &PostmanScript{
						Type: "text/javascript",
						Exec: []string{"console.log(\"bar\")"},
					},
				},
			},
			ProtocolProfileBehavior: "a-protocol-profile-behavior",
			ID:                      "an-id",
			Request: &PostmanRequest{
				URL: &PostmanURL{
					Raw: "http://www.google.fr",
				},
			},
			Responses: []*PostmanResponse{
				{
					Name: "a-response",
				},
			},
		},
		c,
	)
}

func TestCreateItemGroup(t *testing.T) {
	c := CreatePostmanItemGroup(PostmanItemGroup{
		Name:        "An item",
		Description: "A description",
		Variables: []*PostmanVariable{
			{
				Name:  "variable-name",
				Value: "variable-value",
			},
		},
		Events: []*PostmanEvent{
			{
				Listen: PreRequest,
				Script: &PostmanScript{
					Type: "text/javascript",
					Exec: []string{"console.log(\"foo\")"},
				},
			},
			{
				Listen: Test,
				Script: &PostmanScript{
					Type: "text/javascript",
					Exec: []string{"console.log(\"bar\")"},
				},
			},
		},
		ProtocolProfileBehavior: "a-protocol-profile-behavior",
		Items: []*PostmanItems{
			{
				Name: "An item",
			},
		},
		Auth: &PostmanAuth{
			Type: Basic,
		},
	})

	assert.Equal(
		t,
		&PostmanItems{
			Name:        "An item",
			Description: "A description",
			Variables: []*PostmanVariable{
				{
					Name:  "variable-name",
					Value: "variable-value",
				},
			},
			Events: []*PostmanEvent{
				{
					Listen: PreRequest,
					Script: &PostmanScript{
						Type: "text/javascript",
						Exec: []string{"console.log(\"foo\")"},
					},
				},
				{
					Listen: Test,
					Script: &PostmanScript{
						Type: "text/javascript",
						Exec: []string{"console.log(\"bar\")"},
					},
				},
			},
			ProtocolProfileBehavior: "a-protocol-profile-behavior",
			Items: []*PostmanItems{
				{
					Name: "An item",
				},
			},
			Auth: &PostmanAuth{
				Type: Basic,
			},
		},
		c,
	)
}
