package pm

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetParams(t *testing.T) {
	auth := PostmanAuth{
		Type: APIKey,
		APIKey: []*PostmanAuthParam{
			{
				Type:  "string",
				Key:   "apikey-key",
				Value: "apikey-value",
			},
		},
		AWSV4: []*PostmanAuthParam{
			{
				Type:  "string",
				Key:   "awsv4-key",
				Value: "awsv4-value",
			},
		},
		Basic: []*PostmanAuthParam{
			{
				Type:  "string",
				Key:   "basic-key",
				Value: "basic-value",
			},
		},
		Bearer: []*PostmanAuthParam{
			{
				Type:  "string",
				Key:   "bearer-key",
				Value: "bearer-value",
			},
		},
		Digest: []*PostmanAuthParam{
			{
				Type:  "string",
				Key:   "digest-key",
				Value: "digest-value",
			},
		},
		Hawk: []*PostmanAuthParam{
			{
				Type:  "string",
				Key:   "hawk-key",
				Value: "hawk-value",
			},
		},
		NoAuth: []*PostmanAuthParam{
			{
				Type:  "string",
				Key:   "noauth-key",
				Value: "noauth-value",
			},
		},
		OAuth1: []*PostmanAuthParam{
			{
				Type:  "string",
				Key:   "oauth1-key",
				Value: "oauth1-value",
			},
		},
		OAuth2: []*PostmanAuthParam{
			{
				Type:  "string",
				Key:   "oauth2-key",
				Value: "oauth2-value",
			},
		},
		NTLM: []*PostmanAuthParam{
			{
				Type:  "string",
				Key:   "ntlm-key",
				Value: "ntlm-value",
			},
		},
	}

	cases := []struct {
		scenario       string
		authType       authType
		expectedParams []*PostmanAuthParam
	}{
		{
			"GetParams for ApiKey",
			APIKey,
			[]*PostmanAuthParam{
				{
					Type:  "string",
					Key:   "apikey-key",
					Value: "apikey-value",
				},
			},
		},
		{
			"GetParams for AWSV4",
			AWSV4,
			[]*PostmanAuthParam{
				{
					Type:  "string",
					Key:   "awsv4-key",
					Value: "awsv4-value",
				},
			},
		},
		{
			"GetParams for Basic",
			Basic,
			[]*PostmanAuthParam{
				{
					Type:  "string",
					Key:   "basic-key",
					Value: "basic-value",
				},
			},
		},
		{
			"GetParams for Bearer",
			Bearer,
			[]*PostmanAuthParam{
				{
					Type:  "string",
					Key:   "bearer-key",
					Value: "bearer-value",
				},
			},
		},
		{
			"GetParams for Digest",
			Digest,
			[]*PostmanAuthParam{
				{
					Type:  "string",
					Key:   "digest-key",
					Value: "digest-value",
				},
			},
		},
		{
			"GetParams for Hawk",
			Hawk,
			[]*PostmanAuthParam{
				{
					Type:  "string",
					Key:   "hawk-key",
					Value: "hawk-value",
				},
			},
		},
		{
			"GetParams for NoAuth",
			NoAuth,
			[]*PostmanAuthParam{
				{
					Type:  "string",
					Key:   "noauth-key",
					Value: "noauth-value",
				},
			},
		},
		{
			"GetParams for OAuth1",
			OAuth1,
			[]*PostmanAuthParam{
				{
					Type:  "string",
					Key:   "oauth1-key",
					Value: "oauth1-value",
				},
			},
		},
		{
			"GetParams for OAuth2",
			OAuth2,
			[]*PostmanAuthParam{
				{
					Type:  "string",
					Key:   "oauth2-key",
					Value: "oauth2-value",
				},
			},
		},
		{
			"GetParams for NTLM",
			NTLM,
			[]*PostmanAuthParam{
				{
					Type:  "string",
					Key:   "ntlm-key",
					Value: "ntlm-value",
				},
			},
		},
		{
			"GetParams for an unimplemented authentication method",
			"an-unimplemented-authentication-method",
			nil,
		},
	}

	for _, tc := range cases {
		auth.Type = tc.authType

		assert.Equal(
			t,
			tc.expectedParams,
			auth.GetParams(),
			tc.scenario,
		)
	}
}

func TestAuthUnmarshalJSON(t *testing.T) {
	cases := []struct {
		scenario      string
		bytes         []byte
		expectedAuth  *PostmanAuth
		expectedError error
	}{
		{
			"Successfully unmarshalling a basic PostmanAuth v2.1.0",
			[]byte("{\"type\":\"basic\",\"basic\":[{\"key\":\"a-key\",\"value\":\"a-value\"}]}"),
			&PostmanAuth{
				Type: Basic,
				Basic: []*PostmanAuthParam{
					{
						Key:   "a-key",
						Value: "a-value",
					},
				},
			},
			nil,
		},

		{
			"Failed to unmarshal apiKey auth because of an unsupported format",
			[]byte("{\"type\":\"apikey\",\"apikey\":\"invalid-auth-param\"}"),
			&PostmanAuth{
				Type: APIKey,
			},
			errors.New("unsupported type"),
		},
		{
			"Failed to unmarshal awsv4 auth because of an unsupported format",
			[]byte("{\"type\":\"awsv4\",\"awsv4\":\"invalid-auth-param\"}"),
			&PostmanAuth{
				Type: AWSV4,
			},
			errors.New("unsupported type"),
		},
		{
			"Failed to unmarshal basic auth because of an unsupported format",
			[]byte("{\"type\":\"basic\",\"basic\":\"invalid-auth-param\"}"),
			&PostmanAuth{
				Type: Basic,
			},
			errors.New("unsupported type"),
		},
		{
			"Failed to unmarshal bearer auth because of an unsupported format",
			[]byte("{\"type\":\"bearer\",\"bearer\":\"invalid-auth-param\"}"),
			&PostmanAuth{
				Type: Bearer,
			},
			errors.New("unsupported type"),
		},
		{
			"Failed to unmarshal digest auth because of an unsupported format",
			[]byte("{\"type\":\"digest\",\"digest\":\"invalid-auth-param\"}"),
			&PostmanAuth{
				Type: Digest,
			},
			errors.New("unsupported type"),
		},
		{
			"Failed to unmarshal hawk auth because of an unsupported format",
			[]byte("{\"type\":\"hawk\",\"hawk\":\"invalid-auth-param\"}"),
			&PostmanAuth{
				Type: Hawk,
			},
			errors.New("unsupported type"),
		},
		{
			"Failed to unmarshal noauth auth because of an unsupported format",
			[]byte("{\"type\":\"noauth\",\"noauth\":\"invalid-auth-param\"}"),
			&PostmanAuth{
				Type: NoAuth,
			},
			errors.New("unsupported type"),
		},
		{
			"Failed to unmarshal oauth1 auth because of an unsupported format",
			[]byte("{\"type\":\"oauth1\",\"oauth1\":\"invalid-auth-param\"}"),
			&PostmanAuth{
				Type: OAuth1,
			},
			errors.New("unsupported type"),
		},
		{
			"Failed to unmarshal oauth2 auth because of an unsupported format",
			[]byte("{\"type\":\"oauth2\",\"oauth2\":\"invalid-auth-param\"}"),
			&PostmanAuth{
				Type: OAuth2,
			},
			errors.New("unsupported type"),
		},
		{
			"Failed to unmarshal ntlm auth because of an unsupported format",
			[]byte("{\"type\":\"ntlm\",\"ntlm\":\"invalid-auth-param\"}"),
			&PostmanAuth{
				Type: NTLM,
			},
			errors.New("unsupported type"),
		},
	}

	for _, tc := range cases {

		a := new(PostmanAuth)
		err := a.UnmarshalJSON(tc.bytes)

		assert.Equal(t, tc.expectedAuth, a, tc.scenario)
		assert.Equal(t, tc.expectedError, err, tc.scenario)
	}
}

func TestAuthMarshalJSON(t *testing.T) {
	cases := []struct {
		scenario       string
		auth           *PostmanAuth
		expectedOutput string
	}{
		{
			"Successfully marshalling an PostmanAuth v2.1.0",
			&PostmanAuth{
				Type: Basic,
				Basic: []*PostmanAuthParam{
					{
						Key:   "a-key",
						Value: "a-value",
					},
				},
			},
			"{\"type\":\"basic\",\"basic\":[{\"key\":\"a-key\",\"value\":\"a-value\"}]}",
		},
	}

	for _, tc := range cases {
		bytes, _ := tc.auth.MarshalJSON()

		assert.Equal(t, tc.expectedOutput, string(bytes), tc.scenario)
	}
}

func TestCreateAuth(t *testing.T) {

	cases := []struct {
		scenario     string
		auth         *PostmanAuth
		expectedAuth *PostmanAuth
	}{
		{
			scenario: "Create apikey auth",
			auth: CreateAuth(APIKey, &PostmanAuthParam{
				Key:   "a-key",
				Value: "a-value",
			}),
			expectedAuth: &PostmanAuth{
				Type: "apikey",
				APIKey: []*PostmanAuthParam{
					{
						Key:   "a-key",
						Value: "a-value",
					},
				},
			},
		},
		{
			scenario: "Create awsv4 auth",
			auth: CreateAuth(AWSV4, &PostmanAuthParam{
				Key:   "a-key",
				Value: "a-value",
			}),
			expectedAuth: &PostmanAuth{
				Type: "awsv4",
				AWSV4: []*PostmanAuthParam{
					{
						Key:   "a-key",
						Value: "a-value",
					},
				},
			},
		},
		{
			scenario: "Create basic auth",
			auth: CreateAuth(Basic, &PostmanAuthParam{
				Key:   "a-key",
				Value: "a-value",
			}),
			expectedAuth: &PostmanAuth{
				Type: "basic",
				Basic: []*PostmanAuthParam{
					{
						Key:   "a-key",
						Value: "a-value",
					},
				},
			},
		},
		{
			scenario: "Create bearer auth",
			auth: CreateAuth(Bearer, &PostmanAuthParam{
				Key:   "a-key",
				Value: "a-value",
			}),
			expectedAuth: &PostmanAuth{
				Type: "bearer",
				Bearer: []*PostmanAuthParam{
					{
						Key:   "a-key",
						Value: "a-value",
					},
				},
			},
		},
		{
			scenario: "Create digest auth",
			auth: CreateAuth(Digest, &PostmanAuthParam{
				Key:   "a-key",
				Value: "a-value",
			}),
			expectedAuth: &PostmanAuth{
				Type: "digest",
				Digest: []*PostmanAuthParam{
					{
						Key:   "a-key",
						Value: "a-value",
					},
				},
			},
		},
		{
			scenario: "Create hawk auth",
			auth: CreateAuth(Hawk, &PostmanAuthParam{
				Key:   "a-key",
				Value: "a-value",
			}),
			expectedAuth: &PostmanAuth{
				Type: "hawk",
				Hawk: []*PostmanAuthParam{
					{
						Key:   "a-key",
						Value: "a-value",
					},
				},
			},
		},
		{
			scenario: "Create noauth auth",
			auth: CreateAuth(NoAuth, &PostmanAuthParam{
				Key:   "a-key",
				Value: "a-value",
			}),
			expectedAuth: &PostmanAuth{
				Type: "noauth",
				NoAuth: []*PostmanAuthParam{
					{
						Key:   "a-key",
						Value: "a-value",
					},
				},
			},
		},
		{
			scenario: "Create oauth1 auth",
			auth: CreateAuth(OAuth1, &PostmanAuthParam{
				Key:   "a-key",
				Value: "a-value",
			}),
			expectedAuth: &PostmanAuth{
				Type: "oauth1",
				OAuth1: []*PostmanAuthParam{
					{
						Key:   "a-key",
						Value: "a-value",
					},
				},
			},
		},
		{
			scenario: "Create oauth2 auth",
			auth: CreateAuth(OAuth2, &PostmanAuthParam{
				Key:   "a-key",
				Value: "a-value",
			}),
			expectedAuth: &PostmanAuth{
				Type: "oauth2",
				OAuth2: []*PostmanAuthParam{
					{
						Key:   "a-key",
						Value: "a-value",
					},
				},
			},
		},
		{
			scenario: "Create ntlm auth",
			auth: CreateAuth(NTLM, &PostmanAuthParam{
				Key:   "a-key",
				Value: "a-value",
			}),
			expectedAuth: &PostmanAuth{
				Type: "ntlm",
				NTLM: []*PostmanAuthParam{
					{
						Key:   "a-key",
						Value: "a-value",
					},
				},
			},
		},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.expectedAuth, tc.auth, tc.scenario)
	}
}

func TestCreateAuthParam(t *testing.T) {
	assert.Equal(
		t,
		&PostmanAuthParam{
			Key:   "key",
			Value: "value",
			Type:  "string",
		},
		CreateAuthParam("key", "value"),
	)
}
