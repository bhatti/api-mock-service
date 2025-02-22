package types

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Configuration for mock api service
type Configuration struct {
	// HTTPPort for server
	HTTPPort int `yaml:"http_port" mapstructure:"http_port" env:"HTTP_PORT"`
	// ProxyPort for server
	ProxyPort int `yaml:"proxy_port" mapstructure:"proxy_port" env:"PROXY_PORT"`
	// ProxyURLFilter for filtering url
	ProxyURLFilter string `yaml:"proxy_url_filter" mapstructure:"proxy_url_filter" env:"PROXY_URL_FILTER"`
	// ConnectionTimeout for remote server
	ConnectionTimeout int `yaml:"connection_timeout" mapstructure:"connection_timeout"`
	// DataDir for storing api scenarios, history, fixtures, assets, etc.
	DataDir string `yaml:"data_dir" mapstructure:"data_dir" env:"DATA_DIR"`
	// MaxHistory for max limit of storing execution history
	MaxHistory int `yaml:"max_history" mapstructure:"max_history" env:"MAX_HISTORY"`
	// UserAgent for mock api server
	UserAgent string `yaml:"user_agent" mapstructure:"user_agent" env:"USER_AGENT"`
	// ProxyURL for mock api server
	ProxyURL string `yaml:"proxy_url" mapstructure:"proxy_url" env:"PROXY_URL"`
	// AssertHeadersPattern to always match HTTP headers and store them in match-header property of api scenario
	AssertHeadersPattern string `yaml:"assert_headers_pattern" mapstructure:"assert_headers_pattern" env:"ASSERT_HEADERS_PATTERN"`
	// AssertQueryParamsPattern to always match HTTP query params and store them in match-query parameters of api scenario
	AssertQueryParamsPattern string `yaml:"assert_query_params_pattern" mapstructure:"assert_query_params_pattern" env:"ASSERT_QUERY_PATTERN"`
	// AssertPostParamsPattern to always match HTTP post params and store them in match-query parameters of api scenario
	AssertPostParamsPattern string `yaml:"assert_post_params_pattern" mapstructure:"assert_post_params_pattern" env:"ASSERT_POST_PATTERN"`
	CORS                    string `yaml:"cors" mapstructure:"cors" env:"MOCK_CORS"`
	RecordOnly              bool   `yaml:"record_only" mapstructure:"record_only" env:"RECORD_ONLY"`
	Debug                   bool   `yaml:"debug" mapstructure:"debug" env:"MOCK_DEBUG"`
	// Version of API
	Version *Version `yaml:"-" mapstructure:"-" json:"-"`
	// Bearer token for mock api server
	AuthBearerToken string `yaml:"auth_bearer_token" mapstructure:"auth_bearer_token" env:"AUTH_TOKEN"`
	// AWSConfig
	AWS AWSConfig `yaml:"aws" mapstructure:"aws"`
	// BasicAuth
	BasicAuth BasicAuthConfig `yaml:"basic_auth" mapstructure:"basic_auth"`

	OAuth2       OAuth2Config     `yaml:"oauth2" mapstructure:"oauth2"`
	HMAC         HMACConfig       `yaml:"hmac" mapstructure:"hmac"`
	JWT          JWTConfig        `yaml:"jwt" mapstructure:"jwt"`
	Digest       DigestAuthConfig `yaml:"digest" mapstructure:"digest"`
	APIKeyConfig APIKeyConfig     `yaml:"api_key_config" mapstructure:"api_key_config"`

	TestEnvironments []string `yaml:"test_env" mapstructure:"test_env"`
}

// BasicAuthConfig config
type BasicAuthConfig struct {
	Enabled  bool   `yaml:"enabled" mapstructure:"enabled" env:"BASIC_AUTH_ENABLED"`
	Username string `yaml:"username" mapstructure:"username" env:"USERNAME"`
	Password string `yaml:"password" mapstructure:"password" env:"PASSWORD"`
}

// OAuth2Config configuration
type OAuth2Config struct {
	Enabled      bool     `yaml:"enabled" mapstructure:"enabled" env:"OAUTH2_ENABLED"`
	TokenURL     string   `yaml:"token_url" mapstructure:"token_url" env:"OAUTH2_TOKEN_URL"`
	ClientID     string   `yaml:"client_id" mapstructure:"client_id" env:"OAUTH2_CLIENT_ID"`
	ClientSecret string   `yaml:"client_secret" mapstructure:"client_secret" env:"OAUTH2_CLIENT_SECRET"`
	Scopes       []string `yaml:"scopes" mapstructure:"scopes" env:"OAUTH2_SCOPES"`
	GrantType    string   `yaml:"grant_type" mapstructure:"grant_type" env:"OAUTH2_GRANT_TYPE"`
	RefreshToken string   `yaml:"refresh_token" mapstructure:"refresh_token" env:"OAUTH2_REFRESH_TOKEN"`
}

// HMACConfig configuration
type HMACConfig struct {
	Enabled    bool   `yaml:"enabled" mapstructure:"enabled" env:"HMAC_ENABLED"`
	Secret     string `yaml:"secret" mapstructure:"secret" env:"HMAC_SECRET"`
	Algorithm  string `yaml:"algorithm" mapstructure:"algorithm" env:"HMAC_ALGORITHM"`
	HeaderName string `yaml:"header_name" mapstructure:"header_name" env:"HMAC_HEADER_NAME"`
}

// JWTConfig configuration
type JWTConfig struct {
	Enabled     bool   `yaml:"enabled" mapstructure:"enabled" env:"JWT_ENABLED"`
	Secret      string `yaml:"secret" mapstructure:"secret" env:"JWT_SECRET"`
	Algorithm   string `yaml:"algorithm" mapstructure:"algorithm" env:"JWT_ALGORITHM"`
	Issuer      string `yaml:"issuer" mapstructure:"issuer" env:"JWT_ISSUER"`
	Audience    string `yaml:"audience" mapstructure:"audience" env:"JWT_AUDIENCE"`
	ExpiryHours int    `yaml:"expiry_hours" mapstructure:"expiry_hours" env:"JWT_EXPIRY_HOURS"`
}

// DigestAuthConfig configuration
type DigestAuthConfig struct {
	Enabled  bool   `yaml:"enabled" mapstructure:"enabled" env:"DIGEST_ENABLED"`
	Username string `yaml:"username" mapstructure:"username" env:"DIGEST_USERNAME"`
	Password string `yaml:"password" mapstructure:"password" env:"DIGEST_PASSWORD"`
	Realm    string `yaml:"realm" mapstructure:"realm" env:"DIGEST_REALM"`
}

// APIKeyConfig configuration
type APIKeyConfig struct {
	Enabled    bool   `yaml:"enabled" mapstructure:"enabled" env:"API_KEY_ENABLED"`
	Location   string `yaml:"location" mapstructure:"location" env:"API_KEY_LOCATION"`     // header, query, cookie
	HeaderName string `yaml:"header_name" mapstructure:"header_name" env:"API_KEY_HEADER"` // e.g., X-API-Key
	QueryName  string `yaml:"query_name" mapstructure:"query_name" env:"API_KEY_QUERY"`    // e.g., api_key
	CookieName string `yaml:"cookie_name" mapstructure:"cookie_name" env:"API_KEY_COOKIE"` // e.g., api_key
	// API key for mock api server
	APIKey string `yaml:"api_key" mapstructure:"api_key" env:"API_KEY"`
	// GenerateTokenUrl key for mock api server
	GenerateTokenPath string `yaml:"generate_token_path" mapstructure:"generate_token_path" env:"API_KEY_PATH"`
}

// AWSConfig config
type AWSConfig struct {
	Enabled               bool     `yaml:"enabled" mapstructure:"enabled" env:"AWS_ENABLED"`
	StripRequestHeaders   []string `yaml:"strip" mapstructure:"strip" env:"AWS_STRIP_HEADERS"`
	SigningNameOverride   string   `yaml:"name" mapstructure:"name" env:"AWS_SIGNING_NAME"`
	SigningRegionOverride string   `yaml:"aws_region" mapstructure:"aws_region" env:"AWS_REGION"`
	SigningHostOverride   string   `yaml:"sign_host" mapstructure:"sign_host" env:"AWS_SIGN_HOST"`
	HostOverride          string   `yaml:"host" mapstructure:"host" env:"AWS_HOST"`
	ResignAllRequests     bool     `yaml:"resign_all_requests" mapstructure:"resign_all_requests" env:"AWS_RESIGN_ALL_REQUESTS"`
	ResignOnlyExpiredDate bool     `yaml:"resign_only_expired_date" mapstructure:"resign_only_expired_date" env:"AWS_RESIGN_ONLY_EXPIRED"`
	Debug                 bool     `yaml:"debug" mapstructure:"debug" env:"AWS_DEBUG"`
}

// NewConfiguration -- Initializes the default config
func NewConfiguration(
	httpPort int,
	proxyPort int,
	dataDir string,
	version *Version) (config *Configuration, err error) {
	viper.SetDefault("log_level", "info")
	viper.SetDefault("http_port", "8080")
	viper.SetDefault("proxy_port", "8081")
	viper.SetDefault("proxy_url_filter", "")
	viper.SetDefault("data_dir", "default_mocks_data")
	viper.SetDefault("user_agent", "MockService_"+version.Version)
	viper.SetDefault("match_header_regex", "Target")
	viper.SetDefault("match_query_regex", "Target")
	viper.SetDefault("max_history", "500")
	viper.SetDefault("assert_headers_pattern", "")
	viper.SetDefault("assert_query_params_pattern", "")
	viper.SetDefault("cors", "*")
	viper.SetDefault("auth_bearer_token", "")
	viper.SetDefault("record_only", "false")
	viper.SetDefault("debug", "false")
	viper.SetDefault("basic_auth.username", "")
	viper.SetDefault("basic_auth.password", "")

	viper.SetDefault("oauth2.enabled", false)
	viper.SetDefault("oauth2.grant_type", "client_credentials")

	viper.SetDefault("hmac.enabled", false)
	viper.SetDefault("hmac.algorithm", "SHA256")
	viper.SetDefault("hmac.header_name", "X-HMAC-Signature")

	viper.SetDefault("jwt.enabled", false)
	viper.SetDefault("jwt.algorithm", "HS256")
	viper.SetDefault("jwt.expiry_hours", 24)

	viper.SetDefault("digest.enabled", false)
	viper.SetDefault("digest.realm", "protected")

	viper.SetDefault("api_key_config.enabled", false)
	viper.SetDefault("api_key_config.location", "header")
	viper.SetDefault("api_key_config.header_name", "X-API-Key")

	viper.SetDefault("aws.strip", "")
	viper.SetDefault("aws.name", "")
	viper.SetDefault("aws.aws_region", "")
	viper.SetDefault("aws.sign_host", "")
	viper.SetDefault("aws.host", "")
	viper.SetDefault("aws.resign_all_requests", "false")
	viper.SetDefault("aws.resign_only_expired_date", "false")
	viper.SetDefault("aws.debug", "false")
	viper.SetEnvPrefix("")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// If a config file is found, read it in.
	if err = viper.ReadInConfig(); err != nil {
		log.WithFields(log.Fields{
			"Component":  "Configuration",
			"Error":      err,
			"UsedConfig": viper.ConfigFileUsed(),
		}).Debugf("failed to load config file")
	}
	if err = viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	if httpPort > 0 {
		config.HTTPPort = httpPort
	}
	if proxyPort > 0 {
		config.ProxyPort = proxyPort
	}
	if config.HTTPPort == config.ProxyPort {
		return nil, fmt.Errorf("http-port %d cannot be same as proxy-port %d", config.HTTPPort, config.ProxyPort)
	}
	if dataDir != "" {
		config.DataDir = dataDir
	}
	if config.MaxHistory <= 0 {
		config.MaxHistory = 500
	}
	if config.DataDir == "" {
		config.DataDir = "default_mocks_data"
	}
	if os.Getenv("AWS_RESIGN_ONLY_EXPIRED") != "" {
		config.AWS.ResignOnlyExpiredDate = os.Getenv("AWS_RESIGN_ONLY_EXPIRED") == "true"
	}
	if config.CORS == "" {
		config.CORS = "*"
	}
	if os.Getenv("AWS_REGION") != "" {
		config.AWS.SigningRegionOverride = os.Getenv("AWS_REGION")
	}
	if os.Getenv("AWS_SIGNING_NAME") != "" {
		config.AWS.SigningNameOverride = os.Getenv("AWS_SIGNING_NAME")
	}
	if os.Getenv("AWS_RESIGN_ALL_REQUESTS") != "" {
		config.AWS.ResignAllRequests = os.Getenv("AWS_RESIGN_ALL_REQUESTS") == "true"
	}
	if os.Getenv("TEST_ENVS") != "" {
		config.TestEnvironments = strings.Split(os.Getenv("TEST_ENVS"), ",")
	}

	config.APIKeyConfig.Enabled = config.APIKeyConfig.APIKey != ""
	config.BasicAuth.Enabled = config.BasicAuth.Username != "" && config.BasicAuth.Password != ""
	config.Digest.Enabled = config.Digest.Username != "" && config.Digest.Password != ""
	config.HMAC.Enabled = config.HMAC.Secret != "" && config.HMAC.Algorithm != ""
	config.JWT.Enabled = config.JWT.Secret != "" && config.JWT.Issuer != ""
	config.OAuth2.Enabled = config.OAuth2.TokenURL != "" && config.OAuth2.ClientID != "" && config.OAuth2.ClientSecret != ""

	config.Version = version
	log.WithFields(log.Fields{
		"Component":         "Mock-API-Service",
		"Port":              config.HTTPPort,
		"DataDir":           config.DataDir,
		"Version":           version,
		"ResignAllRequests": config.AWS.ResignAllRequests,
		"RecordOnly":        config.RecordOnly,
		"UsedConfig":        viper.ConfigFileUsed(),
	}).Infof("loaded config file...")
	return config, nil
}

func (c *Configuration) GetAuthMethod() AuthType {
	if c.BasicAuth.Enabled {
		return Basic
	}
	if c.AWS.Enabled {
		return AWSV4
	}
	if c.Digest.Enabled {
		return Digest
	}
	if c.HMAC.Enabled {
		return HMAC
	}
	if c.JWT.Enabled {
		return JWT
	}
	if c.OAuth2.TokenURL != "" {
		return OAuth2
	}
	if c.APIKeyConfig.Enabled {
		return APIKey
	}
	return NoAuth
}

// AssertHeader match header
func (c *Configuration) AssertHeader(h string) bool {
	return matchRegex(c.AssertHeadersPattern, h)
}

// AssertQueryParams match query params
func (c *Configuration) AssertQueryParams(p string) bool {
	return matchRegex(c.AssertQueryParamsPattern, p)
}

// AssertPostParams match post params
func (c *Configuration) AssertPostParams(p string) bool {
	return matchRegex(c.AssertPostParamsPattern, p)
}

func matchRegex(re string, str string) bool {
	if re == "" || str == "" {
		return false
	}
	match, err := regexp.Match("(?i)"+re, []byte(str))
	return err == nil && match
}

// BuildTestConfig helper method
func BuildTestConfig() *Configuration {
	return &Configuration{
		UserAgent:                "MockAPI",
		DataDir:                  "../../mock_tests",
		MaxHistory:               500,
		ProxyPort:                8081,
		AssertQueryParamsPattern: "target",
		AssertHeadersPattern:     "target",
		Version:                  NewVersion("1.0", "dev", "x"),
	}
}
