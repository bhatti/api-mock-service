package web

import (
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func Test_ShouldParseServiceRegionForAWSSig4(t *testing.T) {
	signer := &awsSigner{
		awsConfig: &types.AWSConfig{},
	}
	req := &http.Request{
		Header: http.Header{
			Authorization: []string{
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
			Authorization: []string{
				"AWS4-HMAC-SHA256 Credential=ASI/20230217/us-west-2/my-service/aws4_request SignedHeaders=content-encoding;host;x-amz-date;x-amz-requestsupertrace;x-amz-target  Signature=bbb2",
			},
			"Content-Type": []string{"application/json; charset=UTF-8"},
			"X-Amz-Date":   []string{"20230217T182455Z"},
		},
	}
	cred := credentials.NewStaticCredentials("a", "b", "c")
	signed, err := signer.AWSSign(req, cred)
	require.NoError(t, err)
	require.True(t, signed)
}

func Test_ShouldNotSignNonAWSRequest(t *testing.T) {
	signer := NewAWSSigner(&types.Configuration{})
	req := &http.Request{
		Header: http.Header{
			Authorization: []string{
				"Blah",
			},
			"Content-Type": []string{"application/json; charset=UTF-8"},
		},
	}
	cred := credentials.NewStaticCredentials("a", "b", "c")
	signed, err := signer.AWSSign(req, cred)
	require.NoError(t, err)
	require.False(t, signed)
}
