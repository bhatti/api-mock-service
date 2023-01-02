package cmd

import (
	"context"
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"os"

	"github.com/bhatti/api-mock-service/internal/contract"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/web"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var group string
var baseURL string
var executionTimes int
var verbose bool

// contractCmd represents the contract command
var contractCmd = &cobra.Command{
	Use:   "contract",
	Short: "Executes contract client",
	Long:  "Executes contract client",
	Run: func(cmd *cobra.Command, args []string) {
		log.WithFields(log.Fields{
			"DataDir":   dataDir,
			"BaseURL":   baseURL,
			"ExecTimes": executionTimes,
			"Verbose":   verbose}).
			Infof("executing contracts...")
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

		dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 1)
		contractReq := types.NewContractRequest(baseURL, executionTimes)
		contractReq.Verbose = verbose
		executor := contract.NewExecutor(scenarioRepo, web.NewHTTPClient(serverConfig))
		contractRes := executor.ExecuteByGroup(context.Background(), group, dataTemplate, contractReq)
		log.WithFields(log.Fields{
			"Errors":    contractRes.Errors,
			"Succeeded": contractRes.Succeeded,
			"Failed":    contractRes.Failed,
		}).Infof("completed all executions")
	},
}

func init() {
	rootCmd.AddCommand(contractCmd)

	contractCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	contractCmd.Flags().StringVar(&dataDir, "dataDir", "", "data dir to store mock scenarios")
	contractCmd.Flags().StringVar(&group, "group", "", "group of service APIs")
	contractCmd.Flags().StringVar(&baseURL, "base_url", "", "base-url for remote service")
	contractCmd.Flags().IntVar(&executionTimes, "times", 10, "execution times")
	contractCmd.Flags().BoolVar(&verbose, "verbose", false, "verbose logging")
}
