package types

// MockRecordMode header
const MockRecordMode = "X-Mock-Record"

// MockGroup header
const MockGroup = "X-Mock-Group"

// MockPlayback header
const MockPlayback = "X-Mock-Playback"

// MockRecordModeDisabled disabled value
const MockRecordModeDisabled = "false"

// MockRecordModeEnabled enabled value
const MockRecordModeEnabled = "true"

// MockURL header
const MockURL = "X-Mock-Url"

// MockScenarioHeader header
const MockScenarioHeader = "X-Mock-Scenario"

// MockScenarioPath header
const MockScenarioPath = "X-Mock-Path"

// MockChaosEnabled header
const MockChaosEnabled = "X-Mock-Chaos-Enabled"

// ContentTypeHeader header
const ContentTypeHeader = "Content-Type"

// ContentLengthHeader header
const ContentLengthHeader = "Content-Length"

// AuthorizationHeader constant
const AuthorizationHeader = "Authorization"

// MockRequestCount header
const MockRequestCount = "X-Mock-Request-Count"

// MockResponseStatus header
const MockResponseStatus = "X-Mock-Response-Status"

// MockWaitBeforeReply header
const MockWaitBeforeReply = "X-Mock-Wait-Before-Reply"

// ScenarioExt extension
const ScenarioExt = ".yaml"

// AuthType for API authorization
type AuthType string

const (
	// APIKey stands for API Key Authentication.
	APIKey AuthType = "apikey"
	// AWSV4 is Amazon AWS Authentication.
	AWSV4 AuthType = "awsv4"
	// Basic Authentication.
	Basic AuthType = "basic"
	// Bearer Token Authentication.
	Bearer AuthType = "bearer"
	// Digest Authentication.
	Digest AuthType = "digest"
	// Hawk Authentication.
	Hawk AuthType = "hawk"
	// NoAuth Authentication.
	NoAuth AuthType = "noauth"
	// OAuth1 Authentication.
	OAuth1 AuthType = "oauth1"
	// OAuth2 Authentication.
	OAuth2 AuthType = "oauth2"
	// NTLM Authentication.
	NTLM AuthType = "ntlm"
)
