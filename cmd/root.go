package cmd

import (
	"embed"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/contract"
	"github.com/bhatti/api-mock-service/internal/controller"
	"github.com/bhatti/api-mock-service/internal/proxy"
	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/web"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var cfgFile string
var dataDir string
var httpPort int
var proxyPort int

// Version of the queen server
var Version string

// Commit of the last change
var Commit string

// Date of the build
var Date string

// SwaggerContent for embedded swagger-ui
var SwaggerContent embed.FS

// InternalOAPI for embedded open-api specs
var InternalOAPI embed.FS

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
func Execute(version string, commit string, date string, swaggerContent embed.FS, internalOAPI embed.FS) {
	Version = version
	Commit = commit
	Date = date
	SwaggerContent = swaggerContent
	InternalOAPI = internalOAPI
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
		"HTTPPort":  httpPort,
		"ProxyPort": proxyPort,
	}).Infof("starting Mock API-server...")

	serverConfig, err := types.NewConfiguration(httpPort, proxyPort, dataDir, types.NewVersion(Version, Commit, Date))
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err}).
			Errorf("Failed to parse config...")
		os.Exit(1)
	}
	scenarioRepo, fixturesRepo, oapiRepo, err := buildRepos(serverConfig)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).
			Errorf("failed to setup repositories...")
		os.Exit(2)
	}
	webServer := web.NewDefaultWebServer(serverConfig)
	httpClient := web.NewHTTPClient(serverConfig, web.NewAWSSigner(serverConfig))
	if err = buildControllers(serverConfig, scenarioRepo, fixturesRepo, oapiRepo, httpClient, webServer); err != nil {
		log.WithFields(log.Fields{"Error": err}).
			Errorf("failed to setup controller...")
		os.Exit(3)
	}
	go func() {
		fmt.Printf("⇨ http proxy started on \x1b[32m[::]:%d\033[0m\n", serverConfig.ProxyPort)
		adapter := web.NewWebServerAdapter()
		recorder := proxy.NewRecorder(serverConfig, httpClient, scenarioRepo)
		executor := contract.NewProducerExecutor(scenarioRepo, httpClient)
		_ = controller.NewMockOAPIController(InternalOAPI, scenarioRepo, oapiRepo, adapter)
		_ = controller.NewMockScenarioController(scenarioRepo, oapiRepo, adapter)
		_ = controller.NewMockFixtureController(fixturesRepo, adapter)
		_ = controller.NewMockProxyController(recorder, adapter)
		_ = controller.NewContractController(executor, adapter)
		webServer.Embed(SwaggerContent, "/swagger-ui/*", "swagger-ui")
		log.Fatal(proxy.NewProxyHandler(serverConfig, web.NewAWSSigner(serverConfig), scenarioRepo, fixturesRepo, adapter).Start())
	}()

	webServer.Start(":" + strconv.Itoa(serverConfig.HTTPPort))
}

func init() {
	cobra.OnInitialize(initConfig)

	// Cobra also supports local flags, which will only run when this action is called directly.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	rootCmd.Flags().StringVar(&dataDir, "dataDir", "", "data dir to store mock contracts and history")
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
	oapiRepo repository.OAPIRepository,
	err error) {
	scenarioRepo, err = repository.NewFileMockScenarioRepository(serverConfig)
	if err != nil {
		return
	}
	fixtureRepo, err = repository.NewFileFixtureRepository(serverConfig)
	if err != nil {
		return
	}
	oapiRepo, err = repository.NewFileOAPIRepository(serverConfig)
	return
}

func buildControllers(
	serverConfig *types.Configuration,
	scenarioRepo repository.MockScenarioRepository,
	fixtureRepo repository.MockFixtureRepository,
	oapiRepo repository.OAPIRepository,
	httpClient web.HTTPClient,
	webServer web.Server,
) (err error) {
	recorder := proxy.NewRecorder(serverConfig, httpClient, scenarioRepo)
	player := proxy.NewConsumerExecutor(scenarioRepo, fixtureRepo)
	executor := contract.NewProducerExecutor(scenarioRepo, httpClient)
	_ = controller.NewMockOAPIController(InternalOAPI, scenarioRepo, oapiRepo, webServer)
	_ = controller.NewMockScenarioController(scenarioRepo, oapiRepo, webServer)
	_ = controller.NewMockFixtureController(fixtureRepo, webServer)
	_ = controller.NewMockProxyController(recorder, webServer)
	_ = controller.NewContractController(executor, webServer)
	_ = controller.NewRootController(player, webServer)
	assetDir := filepath.Join(serverConfig.DataDir, "assets")
	webServer.Static("/_assets", assetDir)
	webServer.Embed(SwaggerContent, "/swagger-ui/*", "swagger-ui")

	return nil
}
