package pm

// A PostmanVariable allows you to store and reuse values in your requests and scripts.
type PostmanVariable struct {
	ID          string `json:"id,omitempty"`
	Key         string `json:"key,omitempty"`
	Type        string `json:"type,omitempty"`
	Name        string `json:"name,omitempty"`
	Value       string `json:"value,omitempty"`
	Description string `json:"description,omitempty"`
	System      bool   `json:"system,omitempty"`
	Disabled    bool   `json:"disabled,omitempty"`
}

// CreatePostmanVariable creates a new PostmanVariable of type string.
func CreatePostmanVariable(name string, value string) *PostmanVariable {
	return &PostmanVariable{
		Name:  name,
		Value: value,
		Type:  "string",
	}
}

// KeyName accessor
func (v *PostmanVariable) KeyName() string {
	if v.Key != "" {
		return v.Key
	}
	return v.Name
}
