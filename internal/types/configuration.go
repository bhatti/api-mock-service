package types

import (
	"fmt"
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
	// ProxyURL for mock server
	MatchHeaderRegex string `yaml:"match_header_regex" mapstructure:"match_header_regex" env:"MATCH_HEADER_REGEX"`
	// Version of API
	Version *Version `yaml:"-" mapstructure:"-" json:"-"`
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

// MatchHeader match header
func (c *Configuration) MatchHeader(h string) bool {
	if c.MatchHeaderRegex == "" || h == "" {
		return false
	}
	match, err := regexp.Match(c.MatchHeaderRegex, []byte(h))
	return err == nil && match
}
