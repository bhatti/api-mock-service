package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/bhatti/api-mock-service/internal/controller"
	"github.com/bhatti/api-mock-service/internal/proxy"
	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/web"

	"github.com/mitchellh/go-homedir"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var dataDir string
var assetDir string
var httpPort int
var proxyPort int

// Version of the queen server
var Version string

// Commit of the last change
var Commit string

// Date of the build
var Date string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "api-mock-service",
	Short: "Starts mock service",
	Long:  "Starts mock service",
	Run: func(cmd *cobra.Command, args []string) {
		RunServer(cmd, args)
	},
}

// Execute is called by main.main() to start the server
func Execute(version string, commit string, date string) {
	Version = version
	Commit = commit
	Date = date
	if err := rootCmd.Execute(); err != nil {
		log.WithFields(log.Fields{"Error": err}).
			Errorf("failed to execute command...")
		os.Exit(1)
	}
}

// RunServer starts queen server for formicary
func RunServer(_ *cobra.Command, args []string) {
	log.WithFields(log.Fields{
		"Args":      args,
		"DataDir":   dataDir,
		"AssetDir":  assetDir,
		"HTTPPort":  httpPort,
		"ProxyPort": proxyPort,
	}).Infof("starting Mock API-server...")

	serverConfig, err := types.NewConfiguration(httpPort, proxyPort, dataDir, assetDir, types.NewVersion(Version, Commit, Date))
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err}).
			Errorf("Failed to parse config...")
		os.Exit(1)
	}
	scenarioRepo, fixturesRepo, err := buildRepos(serverConfig)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).
			Errorf("failed to setup repositories...")
		os.Exit(2)
	}
	webServer := web.NewDefaultWebServer(serverConfig)
	httpClient := web.NewHTTPClient(serverConfig)
	if err = buildControllers(serverConfig, scenarioRepo, fixturesRepo, httpClient, webServer); err != nil {
		log.WithFields(log.Fields{"Error": err}).
			Errorf("failed to setup controller...")
		os.Exit(3)
	}
	go func() {
		fmt.Printf("⇨ http proxy started on \x1b[32m[::]:%d\033[0m\n", serverConfig.ProxyPort)
		adapter := web.NewWebServerAdapter()
		recorder := proxy.NewRecorder(serverConfig, httpClient, scenarioRepo)
		_ = controller.NewMockOAPIController(scenarioRepo, adapter)
		_ = controller.NewMockScenarioController(scenarioRepo, adapter)
		_ = controller.NewMockFixtureController(fixturesRepo, adapter)
		_ = controller.NewMockProxyController(recorder, adapter)
		log.Fatal(proxy.NewProxyHandler(serverConfig, scenarioRepo, fixturesRepo, adapter).Start())
	}()

	webServer.Start(":" + strconv.Itoa(serverConfig.HTTPPort))
}

func init() {
	cobra.OnInitialize(initConfig)

	// Cobra also supports local flags, which will only run when this action is called directly.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	rootCmd.Flags().StringVar(&dataDir, "dataDir", "", "data dir to store mock scenarios")
	rootCmd.Flags().StringVar(&assetDir, "assetDir", "", "asset dir to store static assets/fixtures")
	rootCmd.Flags().IntVar(&httpPort, "httpPort", 0, "HTTP port to listen")
	rootCmd.Flags().IntVar(&proxyPort, "proxyPort", 0, "Proxy port to listen")

	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
		if log.IsLevelEnabled(log.DebugLevel) {
			log.WithFields(log.Fields{
				"ConfigFile": cfgFile,
			}).Debugf("specifying default config file...")
		}
	} else {
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")
		// Find home directory.
		if home, err := homedir.Dir(); err == nil {
			// Search config in home directory
			viper.AddConfigPath(home)
		}

		viper.SetConfigName("api-mock-service")
		viper.SetConfigType("yaml")
		viper.SetEnvPrefix("")
		viper.AutomaticEnv()
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	}
}

func buildRepos(serverConfig *types.Configuration) (
	scenarioRepo repository.MockScenarioRepository,
	fixtureRepo repository.MockFixtureRepository,
	err error) {
	scenarioRepo, err = repository.NewFileMockScenarioRepository(serverConfig)
	if err != nil {
		return
	}
	fixtureRepo, err = repository.NewFileFixtureRepository(serverConfig)
	return
}

func buildControllers(
	serverConfig *types.Configuration,
	scenarioRepo repository.MockScenarioRepository,
	fixtureRepo repository.MockFixtureRepository,
	httpClient web.HTTPClient,
	webServer web.Server,
) (err error) {
	recorder := proxy.NewRecorder(serverConfig, httpClient, scenarioRepo)
	player := proxy.NewPlayer(scenarioRepo, fixtureRepo)
	_ = controller.NewMockOAPIController(scenarioRepo, webServer)
	_ = controller.NewMockScenarioController(scenarioRepo, webServer)
	_ = controller.NewMockFixtureController(fixtureRepo, webServer)
	_ = controller.NewMockProxyController(recorder, webServer)
	_ = controller.NewRootController(player, webServer)
	if serverConfig.AssetDir != "" {
		webServer.Static(serverConfig.AssetDir)
	}

	return nil
}
