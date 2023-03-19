package postman

import "github.com/bhatti/api-mock-service/internal/types"

// A Response represents an HTTP response.
type Response struct {
	ID              string      `json:"id,omitempty"`
	OriginalRequest *Request    `json:"originalRequest,omitempty"`
	ResponseTime    interface{} `json:"responseTime,omitempty"`
	Timings         interface{} `json:"timings,omitempty"`
	Headers         *HeaderList `json:"header,omitempty"`
	Cookies         []*Cookie   `json:"cookie,omitempty"`
	Body            string      `json:"body,omitempty"`
	Status          string      `json:"status,omitempty"`
	Code            int         `json:"code,omitempty"`
	Name            string      `json:"name,omitempty"`
	PreviewLanguage string      `json:"_postman_previewlanguage,omitempty"`
}

func (r *Response) bodyText() []byte {
	return []byte(replaceTemplateVariables(r.Body))
}

func (r *Response) contentType() string {
	for _, header := range r.Headers.Headers {
		if header.Key == types.ContentTypeHeader {
			return header.Value
		}
	}
	return ""
}

func (r *Response) headersMap() (res map[string][]string) {
	res = make(map[string][]string)
	for _, header := range r.Headers.Headers {
		if !header.Disabled {
			res[header.Key] = res[header.Value]
		}
	}
	return
}
