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
	Debug                   bool   `yaml:"debug" mapstructure:"debug" env:"MOCK_DEBUG"`
	// Version of API
	Version *Version `yaml:"-" mapstructure:"-" json:"-"`
	// AWSConfig
	AWS AWSConfig `yaml:"aws" mapstructure:"aws"`
}

// AWSConfig config
type AWSConfig struct {
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
	viper.SetDefault("data_dir", "default_mocks_data")
	viper.SetDefault("user_agent", "MocService_"+version.Version)
	viper.SetDefault("match_header_regex", "Target")
	viper.SetDefault("match_query_regex", "Target")
	viper.SetDefault("max_history", "500")
	viper.SetDefault("assert_headers_pattern", "")
	viper.SetDefault("assert_query_params_pattern", "")
	viper.SetDefault("cors", "*")
	viper.SetDefault("debug", "false")
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

	config.Version = version
	log.WithFields(log.Fields{
		"Component":             "Mock-API-Service",
		"Port":                  config.HTTPPort,
		"DataDir":               config.DataDir,
		"Version":               version,
		"ResignAllRequests":     config.AWS.ResignAllRequests,
		"ResignOnlyExpiredDate": config.AWS.ResignOnlyExpiredDate,
		"UsedConfig":            viper.ConfigFileUsed(),
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
