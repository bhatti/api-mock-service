package pm

// ListenType defines the kind of script attached to an event.
type ListenType string

const (
	// PreRequest script is usually executed before the HTTP request is sent.
	PreRequest ListenType = "prerequest"
	// Test script is usually executed after the actual HTTP request is sent, and the response is received.
	Test ListenType = "test"
)

// A PostmanScript is a snippet of Javascript code that can be used to to perform setup or teardown operations on a particular response.
type PostmanScript struct {
	ID   string      `json:"id,omitempty"`
	Type string      `json:"type,omitempty"`
	Exec []string    `json:"exec,omitempty"`
	Src  *PostmanURL `json:"src,omitempty"`
	Name string      `json:"name,omitempty"`
}

// An PostmanEvent defines a script associated with an associated event name.
type PostmanEvent struct {
	ID       string         `json:"id,omitempty"`
	Listen   ListenType     `json:"listen,omitempty"`
	Script   *PostmanScript `json:"script,omitempty"`
	Disabled bool           `json:"disabled,omitempty"`
}

// CreatePostmanEvent creates a new PostmanEvent of type text/javascript.
func CreatePostmanEvent(listenType ListenType, script []string) *PostmanEvent {
	return &PostmanEvent{
		Listen: listenType,
		Script: &PostmanScript{
			Type: "text/javascript",
			Exec: script,
		},
	}
}
