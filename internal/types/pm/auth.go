package pm

import (
	"encoding/json"
	"errors"
	"github.com/bhatti/api-mock-service/internal/types"
)

// PostmanAuthParam represents an attribute for any authentication method provided by Postman.
// For example "username" and "password" are set as auth attributes for Basic Authentication method.
type PostmanAuthParam struct {
	Key   string      `json:"key,omitempty"`
	Value interface{} `json:"value,omitempty"`
	Type  string      `json:"type,omitempty"`
}

// PostmanAuth contains the authentication method used and its associated parameters.
type PostmanAuth struct {
	Type   types.AuthType      `json:"type,omitempty"`
	APIKey []*PostmanAuthParam `json:"apikey,omitempty"`
	AWSV4  []*PostmanAuthParam `json:"awsv4,omitempty"`
	Basic  []*PostmanAuthParam `json:"basic,omitempty"`
	Bearer []*PostmanAuthParam `json:"bearer,omitempty"`
	Digest []*PostmanAuthParam `json:"digest,omitempty"`
	Hawk   []*PostmanAuthParam `json:"hawk,omitempty"`
	NoAuth []*PostmanAuthParam `json:"noauth,omitempty"`
	OAuth1 []*PostmanAuthParam `json:"oauth1,omitempty"`
	OAuth2 []*PostmanAuthParam `json:"oauth2,omitempty"`
	NTLM   []*PostmanAuthParam `json:"ntlm,omitempty"`
}

// mAuth is used for marshalling/unmarshalling.
type mAuth struct {
	Type   types.AuthType  `json:"type,omitempty"`
	APIKey json.RawMessage `json:"apikey,omitempty"`
	AWSV4  json.RawMessage `json:"awsv4,omitempty"`
	Basic  json.RawMessage `json:"basic,omitempty"`
	Bearer json.RawMessage `json:"bearer,omitempty"`
	Digest json.RawMessage `json:"digest,omitempty"`
	Hawk   json.RawMessage `json:"hawk,omitempty"`
	NoAuth json.RawMessage `json:"noauth,omitempty"`
	OAuth1 json.RawMessage `json:"oauth1,omitempty"`
	OAuth2 json.RawMessage `json:"oauth2,omitempty"`
	NTLM   json.RawMessage `json:"ntlm,omitempty"`
}

// GetParams returns the parameters related to the authentication method in use.
func (a *PostmanAuth) GetParams() []*PostmanAuthParam {
	switch a.Type {
	case types.APIKey:
		return a.APIKey
	case types.AWSV4:
		return a.AWSV4
	case types.Basic:
		return a.Basic
	case types.Bearer:
		return a.Bearer
	case types.Digest:
		return a.Digest
	case types.Hawk:
		return a.Hawk
	case types.NoAuth:
		return a.NoAuth
	case types.OAuth1:
		return a.OAuth1
	case types.OAuth2:
		return a.OAuth2
	case types.NTLM:
		return a.NTLM
	}

	return nil
}

func (a *PostmanAuth) setParams(params []*PostmanAuthParam) {
	switch a.Type {
	case types.APIKey:
		a.APIKey = params
	case types.AWSV4:
		a.AWSV4 = params
	case types.Basic:
		a.Basic = params
	case types.Bearer:
		a.Bearer = params
	case types.Digest:
		a.Digest = params
	case types.Hawk:
		a.Hawk = params
	case types.NoAuth:
		a.NoAuth = params
	case types.OAuth1:
		a.OAuth1 = params
	case types.OAuth2:
		a.OAuth2 = params
	case types.NTLM:
		a.NTLM = params
	}
}

// UnmarshalJSON parses the JSON-encoded data and create an PostmanAuth from it.
// Depending on the Postman PostmanCollection version, an auth property can either be an array or an object.
//   - v2.1.0 : Array
//   - v2.0.0 : Object
func (a *PostmanAuth) UnmarshalJSON(b []byte) (err error) {
	var tmp mAuth
	err = json.Unmarshal(b, &tmp)

	a.Type = tmp.Type

	if a.APIKey, err = unmarshalAuthParam(tmp.APIKey); err != nil {
		return
	}
	if a.AWSV4, err = unmarshalAuthParam(tmp.AWSV4); err != nil {
		return
	}
	if a.Basic, err = unmarshalAuthParam(tmp.Basic); err != nil {
		return
	}
	if a.Bearer, err = unmarshalAuthParam(tmp.Bearer); err != nil {
		return
	}
	if a.Digest, err = unmarshalAuthParam(tmp.Digest); err != nil {
		return
	}
	if a.Hawk, err = unmarshalAuthParam(tmp.Hawk); err != nil {
		return
	}
	if a.NoAuth, err = unmarshalAuthParam(tmp.NoAuth); err != nil {
		return
	}
	if a.OAuth1, err = unmarshalAuthParam(tmp.OAuth1); err != nil {
		return
	}
	if a.OAuth2, err = unmarshalAuthParam(tmp.OAuth2); err != nil {
		return
	}
	if a.NTLM, err = unmarshalAuthParam(tmp.NTLM); err != nil {
		return
	}

	return
}

func unmarshalAuthParam(b []byte) (a []*PostmanAuthParam, err error) {
	if len(b) > 0 {
		if b[0] != '{' && b[0] != '[' {
			err = errors.New("unsupported type")
		} else {
			_ = json.Unmarshal(b, &a)
		}
	}
	return
}

// MarshalJSON returns the JSON encoding of an PostmanAuth.
func (a *PostmanAuth) MarshalJSON() ([]byte, error) {
	return json.Marshal(PostmanAuth{
		Type:   a.Type,
		APIKey: a.APIKey,
		AWSV4:  a.AWSV4,
		Basic:  a.Basic,
		Bearer: a.Bearer,
		Digest: a.Digest,
		Hawk:   a.Hawk,
		NoAuth: a.NoAuth,
		OAuth1: a.OAuth1,
		OAuth2: a.OAuth2,
		NTLM:   a.NTLM,
	})
}

func authParamsToMap(authParams []*PostmanAuthParam) map[string]interface{} {
	authParamsMap := make(map[string]interface{})

	for _, authParam := range authParams {
		authParamsMap[authParam.Key] = authParam.Value
	}

	return authParamsMap
}

// CreateAuth creates a new PostmanAuth struct with the given parameters.
func CreateAuth(a types.AuthType, params ...*PostmanAuthParam) *PostmanAuth {
	auth := &PostmanAuth{
		Type: a,
	}
	auth.setParams(params)
	return auth
}

// CreateAuthParam creates a new PostmanAuthParam of type string.
func CreateAuthParam(key string, value string) *PostmanAuthParam {
	return &PostmanAuthParam{
		Key:   key,
		Value: value,
		Type:  "string",
	}
}
