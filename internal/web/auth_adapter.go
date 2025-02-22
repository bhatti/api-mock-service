package web

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/golang-jwt/jwt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// AuthAdapter is an interface to make testing http.Client calls easier
type AuthAdapter interface {
	HandleAuth(req *http.Request) (bool, string, error)
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type TokenCache struct {
	mu        sync.RWMutex
	token     string
	expiresAt time.Time
}

// authAdapter structure
type authAdapter struct {
	config     *types.Configuration
	awsSigner  AWSSigner
	tokenCache TokenCache
}

// NewAuthAdapter constructor
func NewAuthAdapter(config *types.Configuration) AuthAdapter {
	return &authAdapter{
		config:    config,
		awsSigner: NewAWSSigner(config),
	}
}

// HandleAuth handles auth if needed
func (a *authAdapter) HandleAuth(req *http.Request) (auth bool, info string, err error) {
	bearerToken := a.config.AuthBearerToken

	if a.config.APIKeyConfig.Enabled {
		req.Header.Set(types.APIKeyHeader, a.config.APIKeyConfig.APIKey)
		if a.config.APIKeyConfig.GenerateTokenPath != "" {
			u := req.URL.Scheme + "://" + req.URL.Host + a.config.APIKeyConfig.GenerateTokenPath
			if accessToken, _ := a.generateToken(a.config.APIKeyConfig.APIKey, u, "POST"); accessToken != "" {
				bearerToken = accessToken
			}
		}
	}

	if bearerToken != "" {
		return a.handleBearer(req, bearerToken)
	}

	switch a.config.GetAuthMethod() {
	case types.Basic:
		req.SetBasicAuth(a.config.BasicAuth.Username, a.config.BasicAuth.Password)
		return true, "basic-auth", nil
	case types.AWSV4:
		return a.handleAWSV4(req)
	case types.Digest:
		return a.handleDigest(req)
	case types.HMAC:
		return a.handleHMAC(req)
	case types.JWT:
		return a.handleJWT(req)
	case types.OAuth2:
		return a.handleOAuth2(req)
	}
	return
}

func (a *authAdapter) generateToken(apiKey, stringUrl, method string) (string, error) {
	a.tokenCache.mu.Lock()
	defer a.tokenCache.mu.Unlock()

	// Check if the cached token is still valid
	if a.tokenCache.token != "" && time.Now().Before(a.tokenCache.expiresAt) {
		return a.tokenCache.token, nil
	}

	// Make the API request
	req, err := http.NewRequest(method, stringUrl, nil)
	if err != nil {
		return "", fmt.Errorf("failed to request %s due to %s", stringUrl, err)
	}
	req.Header.Set(types.APIKeyHeader, apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to connect to %s due to %s", stringUrl, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch token: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var tokenResp TokenResponse
	if err = json.Unmarshal(body, &tokenResp); err != nil {
		return "", err
	}

	if tokenResp.ExpiresIn == 0 {
		tokenResp.ExpiresIn = 1800 // Default to 30 minutes
	}

	a.tokenCache.token = tokenResp.AccessToken
	a.tokenCache.expiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second) // Cache until expires_in - 1 minute

	return a.tokenCache.token, nil
}

func (a *authAdapter) handleAPIKey(req *http.Request) (bool, string, error) {
	if a.config.APIKeyConfig.APIKey == "" {
		return false, "", fmt.Errorf("API key not configured")
	}

	switch a.config.APIKeyConfig.Location {
	case "header":
		req.Header.Set(a.config.APIKeyConfig.HeaderName, a.config.APIKeyConfig.APIKey)
	case "query":
		q := req.URL.Query()
		q.Add(a.config.APIKeyConfig.QueryName, a.config.APIKeyConfig.APIKey)
		req.URL.RawQuery = q.Encode()
	case "cookie":
		cookie := &http.Cookie{
			Name:  a.config.APIKeyConfig.CookieName,
			Value: a.config.APIKeyConfig.APIKey,
		}
		req.AddCookie(cookie)
	}

	return true, "api-key", nil
}

func (a *authAdapter) handleBearer(req *http.Request, token string) (bool, string, error) {
	req.Header.Set(types.AuthorizationHeader, "Bearer "+token)
	return true, "bearer-token-auth", nil
}

func (a *authAdapter) handleAWSV4(req *http.Request) (bool, string, error) {
	staticCredentials := credentials.NewStaticCredentials(
		GetHeaderParamOrEnvValue(nil, AWSAccessKey),
		GetHeaderParamOrEnvValue(nil, AWSSecretKey),
		GetHeaderParamOrEnvValue(nil, AWSSecurityToken, AWSSessionToken),
	)
	return a.awsSigner.AWSSign(req, staticCredentials)
}

func (a *authAdapter) handleJWT(req *http.Request) (bool, string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"iat": time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte(a.config.JWT.Secret))
	if err != nil {
		return false, "", err
	}

	req.Header.Set(types.AuthorizationHeader, "Bearer "+tokenString)
	return true, "jwt", nil
}

func (a *authAdapter) handleDigest(req *http.Request) (bool, string, error) {
	// First make a request to get the nonce and realm
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 401 {
		return false, "", fmt.Errorf("expected 401 status code")
	}

	digestHeader := resp.Header.Get("WWW-Authenticate")
	if !strings.HasPrefix(digestHeader, "Digest ") {
		return false, "", fmt.Errorf("invalid digest header")
	}

	// Parse digest parameters
	params := parseDigestParams(digestHeader)

	// Generate digest response
	ha1 := md5Hash(fmt.Sprintf("%s:%s:%s", a.config.Digest.Username, params["realm"], a.config.Digest.Password))
	ha2 := md5Hash(fmt.Sprintf("%s:%s", req.Method, req.URL.Path))
	response := md5Hash(fmt.Sprintf("%s:%s:%s:%s:%s:%s",
		ha1, params["nonce"], "00000001", "0a4f113b", params["qop"], ha2))

	// Build authorization header
	auth := fmt.Sprintf(`Digest username="%s", realm="%s", nonce="%s", uri="%s", qop=%s, nc=00000001, cnonce="0a4f113b", response="%s"`,
		a.config.Digest.Username, params["realm"], params["nonce"], req.URL.Path, params["qop"], response)

	req.Header.Set(types.AuthorizationHeader, auth)
	return true, "digest", nil
}

func (a *authAdapter) handleHMAC(req *http.Request) (bool, string, error) {
	// Get the request body
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return false, "", err
	}
	// Restore the body for subsequent reads
	req.Body = io.NopCloser(strings.NewReader(string(body)))

	// Create the message to sign
	timestamp := time.Now().Unix()
	message := fmt.Sprintf("%s\n%s\n%d", req.Method, req.URL.Path, timestamp)
	if len(body) > 0 {
		message += fmt.Sprintf("\n%s", string(body))
	}

	// Create signature
	var mac []byte
	switch strings.ToUpper(a.config.HMAC.Algorithm) {
	case "SHA256":
		h := hmac.New(sha256.New, []byte(a.config.HMAC.Secret))
		h.Write([]byte(message))
		mac = h.Sum(nil)
	case "MD5":
		h := hmac.New(md5.New, []byte(a.config.HMAC.Secret))
		h.Write([]byte(message))
		mac = h.Sum(nil)
	default:
		return false, "", fmt.Errorf("unsupported HMAC algorithm")
	}

	signature := base64.StdEncoding.EncodeToString(mac)
	req.Header.Set(a.config.HMAC.HeaderName, signature)
	req.Header.Set("X-Timestamp", fmt.Sprintf("%d", timestamp))

	return true, "hmac", nil
}

func (a *authAdapter) handleOAuth2(req *http.Request) (bool, string, error) {
	a.tokenCache.mu.RLock()
	token := a.tokenCache.token
	expiresAt := a.tokenCache.expiresAt
	a.tokenCache.mu.RUnlock()

	if token == "" || time.Now().After(expiresAt) {
		if err := a.RefreshToken(); err != nil {
			return false, "", err
		}
		a.tokenCache.mu.RLock()
		token = a.tokenCache.token
		a.tokenCache.mu.RUnlock()
	}

	req.Header.Set("Authorization", "Bearer "+token)
	return true, "oauth2", nil
}

func (a *authAdapter) RefreshToken() error {
	data := map[string]string{
		"grant_type":    a.config.OAuth2.GrantType,
		"client_id":     a.config.OAuth2.ClientID,
		"client_secret": a.config.OAuth2.ClientSecret,
	}

	if a.config.OAuth2.RefreshToken != "" {
		data["refresh_token"] = a.config.OAuth2.RefreshToken
	}

	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", a.config.OAuth2.TokenURL, strings.NewReader(string(body)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	a.tokenCache.mu.Lock()
	a.tokenCache.token = result.AccessToken
	a.tokenCache.expiresAt = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)
	a.tokenCache.mu.Unlock()

	return nil
}

func parseDigestParams(header string) map[string]string {
	params := make(map[string]string)
	parts := strings.Split(header[7:], ",")
	for _, part := range parts {
		pair := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(pair) == 2 {
			params[pair[0]] = strings.Trim(pair[1], "\"")
		}
	}
	return params
}

func md5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return fmt.Sprintf("%x", hasher.Sum(nil))
}
