package types

import (
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Configuration for mock services
type Configuration struct {
	// HTTPPort for server
	HTTPPort int `yaml:"http_port" mapstructure:"http_port" env:"HTTP_PORT"`
	// ConnectionTimeout for remote server
	ConnectionTimeout int `yaml:"connection_timeout" mapstructure:"connection_timeout"`
	// DataDir for storing mock responses
	DataDir string `yaml:"data_dir" mapstructure:"data_dir" env:"DATA_DIR"`
	// AssetDir for storing static assets
	AssetDir string `yaml:"asset_dir" mapstructure:"asset_dir" env:"ASSET_DIR"`
	// UserAgent for mock server
	UserAgent string `yaml:"user_agent" mapstructure:"user_agent" env:"USER_AGENT"`
	// ProxyURL for mock server
	ProxyURL string   `yaml:"proxy_url" mapstructure:"proxy_url" env:"PROXY_URL"`
	Version  *Version `yaml:"-" mapstructure:"-" json:"-"`
}

// NewConfiguration -- Initializes the default config
func NewConfiguration(
	port int,
	dataDir string,
	assetDir string,
	version *Version) (config *Configuration, err error) {
	viper.SetDefault("log_level", "info")
	viper.SetDefault("http_port", "7000")
	viper.SetDefault("data_dir", "default_mocks_data")
	viper.SetDefault("asset_dir", "default_assets")
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

	if port > 0 {
		config.HTTPPort = port
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

	if config.AssetDir == "" {
		config.AssetDir = "default_assets"
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
