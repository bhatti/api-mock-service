package web

import (
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/bhatti/api-mock-service/internal/types"
	"net/http"
)

// AuthAdapter is an interface to make testing http.Client calls easier
type AuthAdapter interface {
	HandleAuth(req *http.Request) (bool, string, error)
}

// authAdapter structure
type authAdapter struct {
	config    *types.Configuration
	awsSigner AWSSigner
}

// NewAuthAdapter constructor
func NewAuthAdapter(config *types.Configuration) AuthAdapter {
	return &authAdapter{
		config:    config,
		awsSigner: NewAWSSigner(config),
	}
}

// HandleAuth handles auth if needed
func (a *authAdapter) HandleAuth(req *http.Request) (bool, string, error) {
	staticCredentials := credentials.NewStaticCredentials(
		GetHeaderParamOrEnvValue(nil, AWSAccessKey),
		GetHeaderParamOrEnvValue(nil, AWSSecretKey),
		GetHeaderParamOrEnvValue(nil, AWSSecurityToken, AWSSessionToken),
	)

	if a.config.AuthBearerToken != "" {
		req.Header.Set(types.AuthorizationHeader, "Bearer "+a.config.AuthBearerToken)
		return true, "bearer-token-auth", nil
	} else if a.config.BasicAuth.Username != "" && a.config.BasicAuth.Password != "" {
		req.SetBasicAuth(a.config.BasicAuth.Username, a.config.BasicAuth.Password)
		return true, "basic-auth", nil
	} else {
		return a.awsSigner.AWSSign(req, staticCredentials)
	}
}
