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
	Short: "Starts api mock service",
	Long:  "Starts api mock service",
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
	scenarioRepo, fixturesRepo, oapiRepo, groupConfigRepo, err := buildRepos(serverConfig)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).
			Errorf("failed to setup repositories...")
		os.Exit(2)
	}
	webServer := web.NewDefaultWebServer(serverConfig)
	httpClient := web.NewHTTPClient(serverConfig, web.NewAuthAdapter(serverConfig))
	if err = buildControllers(serverConfig, scenarioRepo, fixturesRepo, oapiRepo, groupConfigRepo, httpClient, webServer); err != nil {
		log.WithFields(log.Fields{"Error": err}).
			Errorf("failed to setup controller...")
		os.Exit(3)
	}
	go func() {
		fmt.Printf("⇨ http proxy started on \x1b[32m[::]:%d\033[0m\n", serverConfig.ProxyPort)
		adapter := web.NewWebServerAdapter()
		recorder := proxy.NewRecorder(serverConfig, httpClient, scenarioRepo, groupConfigRepo)
		executor := contract.NewProducerExecutor(scenarioRepo, groupConfigRepo, httpClient)
		_ = controller.NewOAPIController(serverConfig, InternalOAPI, scenarioRepo, oapiRepo, adapter)
		_ = controller.NewGroupConfigController(groupConfigRepo, adapter)
		_ = controller.NewAPIScenarioController(scenarioRepo, oapiRepo, adapter)
		_ = controller.NewAPIHistoryController(serverConfig, scenarioRepo, adapter)
		_ = controller.NewAPIFixtureController(fixturesRepo, adapter)
		_ = controller.NewAPIProxyController(recorder, adapter)
		_ = controller.NewProducerContractController(executor, adapter)
		webServer.Embed(SwaggerContent, "/swagger-ui/*", "swagger-ui")
		log.Fatal(proxy.NewProxyHandler(serverConfig,
			web.NewAuthAdapter(serverConfig), scenarioRepo, fixturesRepo, groupConfigRepo, adapter).Start())
	}()

	webServer.Start(":" + strconv.Itoa(serverConfig.HTTPPort))
}

func init() {
	cobra.OnInitialize(initConfig)

	// Cobra also supports local flags, which will only run when this action is called directly.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	rootCmd.Flags().StringVar(&dataDir, "dataDir", "", "data dir to store API contracts and history")
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
	scenarioRepo repository.APIScenarioRepository,
	fixtureRepo repository.APIFixtureRepository,
	oapiRepo repository.OAPIRepository,
	groupConfigRepo repository.GroupConfigRepository,
	err error) {
	scenarioRepo, err = repository.NewFileAPIScenarioRepository(serverConfig)
	if err != nil {
		return
	}
	fixtureRepo, err = repository.NewFileFixtureRepository(serverConfig)
	if err != nil {
		return
	}
	oapiRepo, err = repository.NewFileOAPIRepository(serverConfig)
	if err != nil {
		return
	}
	groupConfigRepo, err = repository.NewFileGroupConfigRepository(serverConfig)
	return
}

func buildControllers(
	serverConfig *types.Configuration,
	scenarioRepo repository.APIScenarioRepository,
	fixtureRepo repository.APIFixtureRepository,
	oapiRepo repository.OAPIRepository,
	groupConfigRepo repository.GroupConfigRepository,
	httpClient web.HTTPClient,
	webServer web.Server,
) (err error) {
	recorder := proxy.NewRecorder(serverConfig, httpClient, scenarioRepo, groupConfigRepo)
	player := contract.NewConsumerExecutor(serverConfig, scenarioRepo, fixtureRepo, groupConfigRepo)
	executor := contract.NewProducerExecutor(scenarioRepo, groupConfigRepo, httpClient)
	_ = controller.NewOAPIController(serverConfig, InternalOAPI, scenarioRepo, oapiRepo, webServer)
	_ = controller.NewGroupConfigController(groupConfigRepo, webServer)
	_ = controller.NewAPIScenarioController(scenarioRepo, oapiRepo, webServer)
	_ = controller.NewAPIHistoryController(serverConfig, scenarioRepo, webServer)
	_ = controller.NewAPIFixtureController(fixtureRepo, webServer)
	_ = controller.NewAPIProxyController(recorder, webServer)
	_ = controller.NewProducerContractController(executor, webServer)
	_ = controller.NewRootController(player, webServer)
	assetDir := filepath.Join(serverConfig.DataDir, "assets")
	webServer.Static("/_assets", assetDir)
	webServer.Embed(SwaggerContent, "/swagger-ui/*", "swagger-ui")

	return nil
}
