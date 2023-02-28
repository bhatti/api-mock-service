package web

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/utils"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	log "github.com/sirupsen/logrus"
)

// borrowed basic implementation from https://github.com/awslabs/aws-sigv4-proxy
var services = map[string]endpoints.ResolvedEndpoint{}

// AWSSigner is an interface to make testing http.Client calls easier
type AWSSigner interface {
	AWSSign(req *http.Request, credentials *credentials.Credentials) (bool, error)
}

// awsSigner implements the AWSSigner interface
type awsSigner struct {
	awsConfig *types.AWSConfig
}

// NewAWSSigner constructor
func NewAWSSigner(config *types.Configuration) AWSSigner {
	return &awsSigner{
		awsConfig: &config.AWS,
	}
}

// AWSSign signs request header if needed
func (s *awsSigner) AWSSign(req *http.Request, credentials *credentials.Credentials) (bool, error) {
	if !s.isAWSSig4(req) {
		return false, nil
	}
	if !s.isAWSDateExpired(req) {
		return false, nil
	}
	signer := v4.NewSigner(credentials, func(s *v4.Signer) {})

	oldAuth := req.Header.Get("Authorization")
	if s.awsConfig.HostOverride != "" {
		req.Host = s.awsConfig.HostOverride
	}
	if s.awsConfig.SigningHostOverride != "" {
		req.Host = s.awsConfig.SigningHostOverride
	}

	service := s.getAWSService(req)
	if service == nil {
		return true, fmt.Errorf("unable to determine service from host: %s", req.Host)
	}

	req.Header.Del("Authorization")
	if err := s.sign(req, service, signer); err != nil {
		return true, err
	}

	newAuth := req.Header.Get("Authorization")
	log.WithFields(log.Fields{
		"Component": "DefaultHTTPClient",
		"URL":       req.URL,
		"Method":    req.Method,
		"OldAuth":   oldAuth,
		"NewAuth":   newAuth,
		"Service":   service.SigningName,
		"Region":    service.SigningRegion,
	}).Infof("resigned aws-sig4 auth header for http client")

	// When ContentLength is 0 we also need to set the body to http.NoBody to avoid Go http client
	// to magically set Transfer-Encoding: chunked. Service like S3 does not support chunk encoding.
	// We need to manipulate the Body value after signv4 signing because the signing process wraps
	// the original body into another struct, which will result in Transfer-Encoding: chunked being set.
	if req.ContentLength == 0 {
		req.Body = http.NoBody
	}

	// Remove any headers specified
	for _, header := range s.awsConfig.StripRequestHeaders {
		log.WithField("StripHeader", header).Debug("Stripping Header:")
		req.Header.Del(header)
	}

	return true, nil
}

// isAWSSig4 checks sig4 is defined in auth header
func (s *awsSigner) isAWSSig4(request *http.Request) bool {
	val := strings.ToUpper(request.Header.Get("Authorization"))
	return strings.Contains(val, "AWS4-HMAC-SHA256")
}

// IsAWSDateExpired checks if amz-date is expired
func (s *awsSigner) isAWSDateExpired(request *http.Request) bool {
	dateHeader := request.Header.Get("X-Amz-Date")
	dateVal, _ := time.Parse("20060102T150405Z", dateHeader)
	now := time.Now().UTC().Unix()
	return dateHeader == "" || now-dateVal.Unix() > 5
}

// GetAWSService parses service-region from auth header
func (s *awsSigner) getAWSService(request *http.Request) *endpoints.ResolvedEndpoint {
	auth := request.Header.Get("Authorization")
	if auth != "" {
		var re = regexp.MustCompile(`Credential=.*/.*/(.*)/(.*)/aws4_request`)
		matches := re.FindStringSubmatch(auth)

		if len(matches) > 2 {
			return &endpoints.ResolvedEndpoint{
				URL:           fmt.Sprintf("https://%s", request.Host),
				SigningMethod: "v4",
				SigningRegion: matches[1],
				SigningName:   matches[2],
			}
		}
	}
	if s.awsConfig.SigningNameOverride != "" && s.awsConfig.RegionOverride != "" {
		return &endpoints.ResolvedEndpoint{
			URL:           fmt.Sprintf("https://%s", request.Host),
			SigningMethod: "v4",
			SigningRegion: s.awsConfig.RegionOverride,
			SigningName:   s.awsConfig.SigningNameOverride}
	}
	return determineAWSServiceFromHost(request.Host)
}

func (s *awsSigner) sign(req *http.Request, service *endpoints.ResolvedEndpoint, signer *v4.Signer) error {
	_, body, err := utils.ReadAll(req.Body)
	req.Body = body

	// S3 service should not have any escaping applied.
	// https://github.com/aws/aws-sdk-go/blob/main/aws/signer/v4/v4.go#L467-L470
	if service.SigningName == "s3" {
		signer.DisableURIPathEscaping = true

		// Enable URI escaping for subsequent calls.
		defer func() {
			signer.DisableURIPathEscaping = false
		}()
	}

	_, err = signer.Sign(req, body, service.SigningName, service.SigningRegion, time.Now())

	return err
}

func determineAWSServiceFromHost(host string) *endpoints.ResolvedEndpoint {
	for endpoint, service := range services {
		if host == endpoint {
			return &service
		}
	}
	return nil
}

func init() {
	// Triple nested loop - 😭
	for _, partition := range endpoints.DefaultPartitions() {

		for _, service := range partition.Services() {
			for _, endpoint := range service.Endpoints() {
				resolvedEndpoint, _ := endpoint.ResolveEndpoint()
				host := strings.Replace(resolvedEndpoint.URL, "https://", "", 1)
				services[host] = resolvedEndpoint
			}
		}
	}

	// Add api gateway endpoints
	for region := range endpoints.AwsPartition().Regions() {
		host := fmt.Sprintf("execute-api.%s.amazonaws.com", region)
		services[host] = endpoints.ResolvedEndpoint{URL: fmt.Sprintf("https://%s", host), SigningMethod: "v4", SigningRegion: region, SigningName: "execute-api", PartitionID: "aws"}
	}
	// Add elasticsearch endpoints
	for region := range endpoints.AwsPartition().Regions() {
		host := fmt.Sprintf("%s.es.amazonaws.com", region)
		services[host] = endpoints.ResolvedEndpoint{URL: fmt.Sprintf("https://%s", host), SigningMethod: "v4", SigningRegion: region, SigningName: "es", PartitionID: "aws"}
	}
	// Add managed prometheus + workspace endpoints
	for region := range endpoints.AwsPartition().Regions() {
		hostAps := fmt.Sprintf("aps.%s.amazonaws.com", region)
		services[hostAps] = endpoints.ResolvedEndpoint{URL: fmt.Sprintf("https://%s", hostAps), SigningMethod: "v4", SigningRegion: region, SigningName: "aps", PartitionID: "aws"}

		hostApsws := fmt.Sprintf("aps-workspaces.%s.amazonaws.com", region)
		services[hostApsws] = endpoints.ResolvedEndpoint{URL: fmt.Sprintf("https://%s", hostApsws), SigningMethod: "v4", SigningRegion: region, SigningName: "aps", PartitionID: "aws"}
	}
}