package postman

import (
	"encoding/json"
	"errors"
)

type authType string

const (
	// APIKey stands for API Key Authentication.
	APIKey authType = "apikey"
	// AWSV4 is Amazon AWS Authentication.
	AWSV4 authType = "awsv4"
	// Basic Authentication.
	Basic authType = "basic"
	// Bearer Token Authentication.
	Bearer authType = "bearer"
	// Digest Authentication.
	Digest authType = "digest"
	// Hawk Authentication.
	Hawk authType = "hawk"
	// NoAuth Authentication.
	NoAuth authType = "noauth"
	// OAuth1 Authentication.
	OAuth1 authType = "oauth1"
	// OAuth2 Authentication.
	OAuth2 authType = "oauth2"
	// NTLM Authentication.
	NTLM authType = "ntlm"
)

// AuthParam represents an attribute for any authentication method provided by Postman.
// For example "username" and "password" are set as auth attributes for Basic Authentication method.
type AuthParam struct {
	Key   string      `json:"key,omitempty"`
	Value interface{} `json:"value,omitempty"`
	Type  string      `json:"type,omitempty"`
}

// Auth contains the authentication method used and its associated parameters.
type Auth struct {
	Type   authType     `json:"type,omitempty"`
	APIKey []*AuthParam `json:"apikey,omitempty"`
	AWSV4  []*AuthParam `json:"awsv4,omitempty"`
	Basic  []*AuthParam `json:"basic,omitempty"`
	Bearer []*AuthParam `json:"bearer,omitempty"`
	Digest []*AuthParam `json:"digest,omitempty"`
	Hawk   []*AuthParam `json:"hawk,omitempty"`
	NoAuth []*AuthParam `json:"noauth,omitempty"`
	OAuth1 []*AuthParam `json:"oauth1,omitempty"`
	OAuth2 []*AuthParam `json:"oauth2,omitempty"`
	NTLM   []*AuthParam `json:"ntlm,omitempty"`
}

// mAuth is used for marshalling/unmarshalling.
type mAuth struct {
	Type   authType        `json:"type,omitempty"`
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
func (a *Auth) GetParams() []*AuthParam {
	switch a.Type {
	case APIKey:
		return a.APIKey
	case AWSV4:
		return a.AWSV4
	case Basic:
		return a.Basic
	case Bearer:
		return a.Bearer
	case Digest:
		return a.Digest
	case Hawk:
		return a.Hawk
	case NoAuth:
		return a.NoAuth
	case OAuth1:
		return a.OAuth1
	case OAuth2:
		return a.OAuth2
	case NTLM:
		return a.NTLM
	}

	return nil
}

func (a *Auth) setParams(params []*AuthParam) {
	switch a.Type {
	case APIKey:
		a.APIKey = params
	case AWSV4:
		a.AWSV4 = params
	case Basic:
		a.Basic = params
	case Bearer:
		a.Bearer = params
	case Digest:
		a.Digest = params
	case Hawk:
		a.Hawk = params
	case NoAuth:
		a.NoAuth = params
	case OAuth1:
		a.OAuth1 = params
	case OAuth2:
		a.OAuth2 = params
	case NTLM:
		a.NTLM = params
	}
}

// UnmarshalJSON parses the JSON-encoded data and create an Auth from it.
// Depending on the Postman Collection version, an auth property can either be an array or an object.
//   - v2.1.0 : Array
//   - v2.0.0 : Object
func (a *Auth) UnmarshalJSON(b []byte) (err error) {
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

func unmarshalAuthParam(b []byte) (a []*AuthParam, err error) {
	if len(b) > 0 {
		if b[0] != '{' && b[0] != '[' {
			err = errors.New("unsupported type")
		} else {
			_ = json.Unmarshal(b, &a)
		}
	}
	return
}

// MarshalJSON returns the JSON encoding of an Auth.
func (a *Auth) MarshalJSON() ([]byte, error) {
	return json.Marshal(Auth{
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

func authParamsToMap(authParams []*AuthParam) map[string]interface{} {
	authParamsMap := make(map[string]interface{})

	for _, authParam := range authParams {
		authParamsMap[authParam.Key] = authParam.Value
	}

	return authParamsMap
}

// CreateAuth creates a new Auth struct with the given parameters.
func CreateAuth(a authType, params ...*AuthParam) *Auth {
	auth := &Auth{
		Type: a,
	}
	auth.setParams(params)
	return auth
}

// CreateAuthParam creates a new AuthParam of type string.
func CreateAuthParam(key string, value string) *AuthParam {
	return &AuthParam{
		Key:   key,
		Value: value,
		Type:  "string",
	}
}
