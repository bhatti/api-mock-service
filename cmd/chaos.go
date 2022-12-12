package cmd

import (
	"context"
	"os"

	"github.com/bhatti/api-mock-service/internal/chaos"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/web"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var group string
var baseURL string
var executionTimes int
var verbose bool

// chaosCmd represents the chaos command
var chaosCmd = &cobra.Command{
	Use:   "chaos",
	Short: "Executes chaos client",
	Long:  "Executes chaos client",
	Run: func(cmd *cobra.Command, args []string) {
		log.WithFields(log.Fields{
			"DataDir":   dataDir,
			"BaseURL":   baseURL,
			"ExecTimes": executionTimes,
			"Verbose":   verbose}).
			Infof("executing chaos...")
		if baseURL == "" {
			log.Errorf("baseURL is not specified")
			os.Exit(1)
		}
		if group == "" {
			log.Errorf("group is not specified")
			os.Exit(2)
		}
		serverConfig, err := types.NewConfiguration(httpPort, proxyPort, dataDir, assetDir, types.NewVersion(Version, Commit, Date))
		if err != nil {
			log.Errorf("failed to parse config %s", err)
			os.Exit(3)
		}
		scenarioRepo, _, err := buildRepos(serverConfig)
		if err != nil {
			log.Errorf("failed to setup scenario repository %s", err)
			os.Exit(4)
		}

		dataTemplate := types.NewDataTemplateRequest(false, 1, 1)
		chaosReq := types.NewChaosRequest(baseURL, executionTimes)
		chaosReq.Verbose = verbose
		executor := chaos.NewExecutor(scenarioRepo, web.NewHTTPClient(serverConfig))
		res := executor.ExecuteByGroup(context.Background(), group, dataTemplate, chaosReq)
		log.WithFields(log.Fields{
			"Errors":    res.Errors,
			"Succeeded": res.Succeeded,
			"Failed":    res.Failed,
		}).Infof("completed all executions")
	},
}

func init() {
	rootCmd.AddCommand(chaosCmd)

	chaosCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	chaosCmd.Flags().StringVar(&dataDir, "dataDir", "", "data dir to store mock scenarios")
	chaosCmd.Flags().StringVar(&group, "group", "", "group of service APIs")
	chaosCmd.Flags().StringVar(&baseURL, "base_url", "", "base-url for remote service")
	chaosCmd.Flags().IntVar(&executionTimes, "times", 10, "execution times")
	chaosCmd.Flags().BoolVar(&verbose, "verbose", false, "verbose logging")
}
