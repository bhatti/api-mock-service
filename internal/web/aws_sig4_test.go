package web

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"
)

func Test_ShouldParseServiceRegionForAWSSig4(t *testing.T) {
	signer := &awsSigner{
		awsConfig: &types.AWSConfig{},
	}
	req := &http.Request{
		Header: http.Header{
			types.AuthorizationHeader: []string{
				"AWS4-HMAC-SHA256 Credential=ASI/20230217/us-west-2/my-service/aws4_request SignedHeaders=content-encoding;host;x-amz-date;x-amz-requestsupertrace;x-amz-target  Signature=bbb2",
			},
			"Content-Encoding": []string{"amz-1.0"},
			"Content-Type":     []string{"application/json; charset=UTF-8"},
			"X-Amz-Date":       []string{"20230217T182455Z"},
		},
	}
	service := signer.getAWSService(req)
	assert.Equal(t, "my-service", service.SigningName)
	assert.Equal(t, "us-west-2", service.SigningRegion)
}

func Test_ShouldCheckExpiredDateForAWSSig4(t *testing.T) {
	signer := &awsSigner{
		awsConfig: &types.AWSConfig{},
	}
	req := &http.Request{
		Header: http.Header{
			"Content-Encoding": []string{"amz-1.0"},
			"Content-Type":     []string{"application/json; charset=UTF-8"},
			"X-Amz-Date":       []string{"20230301T175147Z"},
		},
	}
	expired, diff := signer.isAWSDateExpired(req)
	assert.True(t, diff > 5)
	assert.True(t, expired)
}

func Test_ShouldCheckExpiredDateWithValidDateForAWSSig4(t *testing.T) {
	signer := &awsSigner{
		awsConfig: &types.AWSConfig{},
	}
	date := time.Now().UTC().Format("20060102T150405Z")
	req := &http.Request{
		Header: http.Header{
			"Content-Encoding": []string{"amz-1.0"},
			"Content-Type":     []string{"application/json; charset=UTF-8"},
			"X-Amz-Date":       []string{date},
		},
	}
	expired, _ := signer.isAWSDateExpired(req)
	assert.False(t, expired)
}

func Test_ShouldCheckExpiredDateWithoutHeaderForAWSSig4(t *testing.T) {
	signer := &awsSigner{
		awsConfig: &types.AWSConfig{},
	}
	req := &http.Request{
		Header: http.Header{
			"Content-Encoding": []string{"amz-1.0"},
			"Content-Type":     []string{"application/json; charset=UTF-8"},
		},
	}
	expired, _ := signer.isAWSDateExpired(req)
	assert.True(t, expired)
}

func Test_ShouldSignAWSRequest(t *testing.T) {
	signer := NewAWSSigner(&types.Configuration{})
	u, _ := url.Parse("http://localhost:8080")
	req := &http.Request{
		URL: u,
		Header: http.Header{
			types.AuthorizationHeader: []string{
				"AWS4-HMAC-SHA256 Credential=ASI/20230217/us-west-2/my-service/aws4_request SignedHeaders=content-encoding;host;x-amz-date;x-amz-requestsupertrace;x-amz-target  Signature=bbb2",
			},
			"Content-Type": []string{"application/json; charset=UTF-8"},
			"X-Amz-Date":   []string{"20230217T182455Z"},
		},
	}
	cred := credentials.NewStaticCredentials("a", "b", "c")
	signed, _, err := signer.AWSSign(req, cred)
	require.NoError(t, err)
	require.True(t, signed)
}

func Test_ShouldNotSignNonAWSRequest(t *testing.T) {
	signer := NewAWSSigner(&types.Configuration{})
	req := &http.Request{
		Header: http.Header{
			types.AuthorizationHeader: []string{
				"Blah",
			},
			"Content-Type": []string{"application/json; charset=UTF-8"},
		},
	}
	cred := credentials.NewStaticCredentials("a", "b", "c")
	signed, _, err := signer.AWSSign(req, cred)
	require.NoError(t, err)
	require.False(t, signed)
}

func Test_ShouldSignAndVerifySignature4(t *testing.T) {
	u, err := url.Parse("https://cognito-idp.us-west-2.amazonaws.com")
	require.NoError(t, err)
	req := &http.Request{
		Header: http.Header{
			types.AuthorizationHeader: []string{"AWS4-HMAC-SHA256 Credential"},
			"X-Amz-Date":              []string{"20230308T024331Z"},
			"X-Amz-Target":            []string{"AWSCognitoIdentityProviderService.AdminGetUser"},
			"Content-Type":            []string{"application/x-amz-json-1.1"},
		},
		Method: "POST",
		URL:    u,
	}
	awsCred := credentials.NewStaticCredentials(
		"ABC",
		"XYZ",
		"",
	)
	signer := v4.NewSigner(awsCred)
	signer.Debug = aws.LogDebugWithSigning
	signer.Logger = awsLoggerAdapter{}
	_, body, err := utils.ReadAll(io.NopCloser(bytes.NewReader([]byte(`{"UserPoolId":"us-west-2_xxxx","Username":"bob@gmail.com"}`))))
	require.NoError(t, err)
	headers, err := signer.Sign(req, body, "cognito-idp", "us-west-2", time.Now())
	require.NoError(t, err)
	require.True(t, len(headers) > 0)
	ctx := parseAuthHeader(req)
	require.NotNil(t, ctx)
}

type signingCtx struct {
	ServiceName      string
	Region           string
	Request          *http.Request `yaml:"-" json:"-"`
	Body             io.ReadSeeker `yaml:"-" json:"-"`
	Query            url.Values
	Time             time.Time
	ExpireTime       time.Duration
	SignedHeaderVals http.Header
	HeaderNames      []string
	BodyDigest       string
	Authorization    string

	key              string
	date             string
	signedHeaders    string
	canonicalHeaders string
	canonicalString  string
	credentialString string
	stringToSign     string
	signature        string
}

func parseAuthHeader(req *http.Request) *signingCtx {
	auth := req.Header.Get("Authorization")

	var re = regexp.MustCompile(`AWS4-HMAC-SHA256 Credential=(.*)/(.*)/(.*)/(.*)/aws4_request, SignedHeaders=(.*), Signature=(.*)`)
	matches := re.FindStringSubmatch(auth)

	if len(matches) < 6 {
		return nil
	}
	b, body, err := utils.ReadAll(req.Body)
	if err != nil {
		return nil
	}
	date, err := time.Parse("20060102T150405Z", req.Header.Get("X-Amz-Date"))
	if err != nil {
		return nil
	}
	return &signingCtx{
		key:              matches[1],
		date:             matches[2],
		Region:           matches[3],
		ServiceName:      matches[4],
		HeaderNames:      strings.Split(matches[5], ";"),
		signature:        matches[6],
		Request:          req,
		Body:             body,
		Query:            req.URL.Query(),
		SignedHeaderVals: req.Header,
		Authorization:    auth,
		BodyDigest:       fmt.Sprintf("%x", sha256.Sum256(b)),
		Time:             date,
		//ExpireTime, date.Duration
		//signedHeaders    string
		//canonicalHeaders string
		//canonicalString  string
		//credentialString string
		//stringToSign     string
	}
}
