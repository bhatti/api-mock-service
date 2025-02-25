package pm

// These constants represent the available raw languages.
const (
	HTML       string = "html"
	Javascript string = "javascript"
	JSON       string = "json"
	Text       string = "text"
	XML        string = "xml"
)

// PostmanRequestBody represents the data usually contained in the request body.
type PostmanRequestBody struct {
	Mode       string                     `json:"mode"`
	Raw        string                     `json:"raw,omitempty"`
	URLEncoded interface{}                `json:"urlencoded,omitempty"`
	FormData   interface{}                `json:"formdata,omitempty"`
	File       interface{}                `json:"file,omitempty"`
	GraphQL    interface{}                `json:"graphql,omitempty"`
	Disabled   bool                       `json:"disabled,omitempty"`
	Options    *PostmanRequestBodyOptions `json:"options,omitempty"`
}

// PostmanRequestBodyOptions holds body options.
type PostmanRequestBodyOptions struct {
	Raw PostmanRequestBodyOptionsRaw `json:"raw,omitempty"`
}

// PostmanRequestBodyOptionsRaw represents the acutal language to use in postman. (See possible options in the cost above)
type PostmanRequestBodyOptionsRaw struct {
	Language string `json:"language,omitempty"`
}
