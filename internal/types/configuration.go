package types

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Configuration for mock services
type Configuration struct {
	// HTTPPort for server
	HTTPPort int `yaml:"http_port" mapstructure:"http_port" env:"HTTP_PORT"`
	// ProxyPort for server
	ProxyPort int `yaml:"proxy_port" mapstructure:"proxy_port" env:"PROXY_PORT"`
	// ConnectionTimeout for remote server
	ConnectionTimeout int `yaml:"connection_timeout" mapstructure:"connection_timeout"`
	// DataDir for storing mock responses
	DataDir string `yaml:"data_dir" mapstructure:"data_dir" env:"DATA_DIR"`
	// AssetDir for storing static assets
	AssetDir string `yaml:"asset_dir" mapstructure:"asset_dir" env:"ASSET_DIR"`
	// UserAgent for mock server
	UserAgent string `yaml:"user_agent" mapstructure:"user_agent" env:"USER_AGENT"`
	// ProxyURL for mock server
	ProxyURL string `yaml:"proxy_url" mapstructure:"proxy_url" env:"PROXY_URL"`
	// AssertHeadersPattern to always match HTTP headers and store them in match-header property of mock scenario
	AssertHeadersPattern string `yaml:"assert_headers_pattern" mapstructure:"assert_headers_pattern" env:"ASSERT_HEADERS_PATTERN"`
	// AssertQueryParamsPattern to always match HTTP query params and store them in match-query parameters of mock scenario
	AssertQueryParamsPattern string `yaml:"assert_query_params_pattern" mapstructure:"assert_query_params_pattern" env:"ASSERT_QUERY_PATTERN"`
	Debug                    bool   `yaml:"debug" mapstructure:"debug" env:"MOCK_DEBUG"`
	CORS                     string `yaml:"cors" mapstructure:"cors" env:"MOCK_CORS"`
	// Version of API
	Version *Version `yaml:"-" mapstructure:"-" json:"-"`
	// AWSConfig
	AWS AWSConfig `yaml:"aws" mapstructure:"aws"`
}

// AWSConfig config
type AWSConfig struct {
	StripRequestHeaders   []string `yaml:"strip" mapstructure:"strip" env:"AWS_STRIP_HEADERS"`
	SigningNameOverride   string   `yaml:"name" mapstructure:"name" env:"AWS_SIGNING_NAME"`
	SigningHostOverride   string   `yaml:"sign-host" mapstructure:"sign-host" env:"AWS_SIGN_HOST"`
	HostOverride          string   `yaml:"host" mapstructure:"host" env:"AWS_HOST"`
	RegionOverride        string   `yaml:"region" mapstructure:"region" env:"AWS_REGION"`
	ResignOnlyExpiredDate bool     `yaml:"resign_only_expired_date" mapstructure:"resign_only_expired_date" env:"AWS_RESIGN_ONLY_EXPIRED"`
	Debug                 bool     `yaml:"debug" mapstructure:"debug" env:"AWS_DEBUG"`
}

// NewConfiguration -- Initializes the default config
func NewConfiguration(
	httpPort int,
	proxyPort int,
	dataDir string,
	assetDir string,
	version *Version) (config *Configuration, err error) {
	viper.SetDefault("log_level", "info")
	viper.SetDefault("http_port", "8080")
	viper.SetDefault("proxy_port", "8081")
	viper.SetDefault("data_dir", "default_mocks_data")
	viper.SetDefault("match_header_regex", "Target")
	viper.SetDefault("match_query_regex", "Target")
	viper.SetDefault("asset_dir", "")
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
	if assetDir != "" {
		config.AssetDir = assetDir
	}

	if config.DataDir == "" {
		config.DataDir = "default_mocks_data"
	}
	if os.Getenv("AWS_DEBUG") != "" {
		config.AWS.Debug = os.Getenv("AWS_DEBUG") == "true"
	}
	if os.Getenv("AWS_RESIGN_ONLY_EXPIRED") != "" {
		config.AWS.ResignOnlyExpiredDate = os.Getenv("AWS_RESIGN_ONLY_EXPIRED") == "true"
	}
	if os.Getenv("MOCK_CORS") != "" {
		config.CORS = os.Getenv("MOCK_CORS")
	}
	if config.CORS == "" {
		config.CORS = "*"
	}

	config.Version = version
	log.WithFields(log.Fields{
		"Component":  "Mock-API-Service",
		"Port":       config.HTTPPort,
		"DataDir":    config.DataDir,
		"AssetDir":   config.AssetDir,
		"Version":    version,
		"UsedConfig": viper.ConfigFileUsed(),
	}).Infof("loaded config file...")
	return config, nil
}

// AssertHeader match header
func (c *Configuration) AssertHeader(h string) bool {
	return matchRegex(c.AssertHeadersPattern, h)
}

// AssertQueryParams match query params
func (c *Configuration) AssertQueryParams(p string) bool {
	return matchRegex(c.AssertQueryParamsPattern, p)
}

func matchRegex(re string, str string) bool {
	if re == "" || str == "" {
		return false
	}
	match, err := regexp.Match("(?i)"+re, []byte(str))
	return err == nil && match
}
