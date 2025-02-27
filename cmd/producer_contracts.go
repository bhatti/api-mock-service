package cmd

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v3"
	"net/http"
	"os"

	"github.com/bhatti/api-mock-service/internal/contract"
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/web"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var group string
var baseURL string
var executionTimes int
var verbose bool

// producerContractCmd represents the contract command
var producerContractCmd = &cobra.Command{
	Use:   "producer-contract",
	Short: "Executes producer contracts",
	Long:  "Executes producer contracts",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// If scenario file is not provided, group is required
		if scenarioFile == "" && group == "" {
			return fmt.Errorf("either group or scenario file must be specified")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.WithFields(log.Fields{
			"DataDir":      dataDir,
			"BaseURL":      baseURL,
			"ExecTimes":    executionTimes,
			"Verbose":      verbose,
			"ScenarioFile": scenarioFile,
		}).Infof("executing producer contracts...")

		serverConfig, err := types.NewConfiguration(
			httpPort,
			proxyPort,
			dataDir,
			types.NewVersion(Version, Commit, Date))
		if err != nil {
			log.Errorf("failed to parse config %s", err)
			os.Exit(1)
		}

		scenarioRepo, _, _, groupConfigRepo, err := buildRepos(serverConfig)
		if err != nil {
			log.Errorf("failed to setup scenario repository %s", err)
			os.Exit(2)
		}

		dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 1)
		contractReq := types.NewProducerContractRequest(baseURL, executionTimes, 0)
		contractReq.Verbose = verbose

		executor := contract.NewProducerExecutor(
			scenarioRepo,
			groupConfigRepo,
			web.NewHTTPClient(serverConfig, web.NewAuthAdapter(serverConfig)),
		)

		var contractRes *types.ProducerContractResponse

		if scenarioFile != "" {
			// Load scenario file and create key data
			keyData, err := loadScenarioKeyData(scenarioFile)
			if err != nil {
				log.Errorf("failed to load scenario file: %s", err)
				os.Exit(3)
			}

			// Execute with specific scenario
			contractRes = executor.Execute(context.Background(), &http.Request{}, keyData, dataTemplate, contractReq)
		} else {
			// Execute by group
			contractRes = executor.ExecuteByGroup(context.Background(), &http.Request{}, group, dataTemplate, contractReq)
		}

		// Print execution summary
		fmt.Printf("\nExecution Summary:\n")
		fmt.Printf("Total Executions: %d (Succeeded: %d, Failed: %d, Mismatched: %d)\n",
			contractRes.Succeeded+contractRes.Failed+contractRes.Mismatched,
			contractRes.Succeeded, contractRes.Failed, contractRes.Mismatched)

		// Print URLs accessed
		if len(contractRes.URLs) > 0 {
			fmt.Printf("\nURLs Accessed:\n")
			for url, count := range contractRes.URLs {
				fmt.Printf("  %s: %d executions\n", url, count)
			}
		}

		// Print metrics
		if len(contractRes.Metrics) > 0 {
			fmt.Printf("\nPerformance Metrics:\n")
			for metric, value := range contractRes.Metrics {
				fmt.Printf("  %s: %.2f\n", metric, value)
			}
		}

		// Print errors if any
		if len(contractRes.Errors) > 0 {
			fmt.Printf("\nErrors (%d):\n", len(contractRes.Errors))
			for scenario, errMsg := range contractRes.Errors {
				fmt.Printf("  %s: %s\n", scenario, errMsg)
			}
		}

		// Print results if verbose
		if verbose && len(contractRes.Results) > 0 {
			fmt.Printf("\nDetailed Results:\n")
			for key, result := range contractRes.Results {
				fmt.Printf("  %s: %v\n", key, result)
			}
		}

		log.WithFields(log.Fields{
			"Errors":     len(contractRes.Errors),
			"Succeeded":  contractRes.Succeeded,
			"Failed":     contractRes.Failed,
			"Mismatched": contractRes.Mismatched,
		}).Infof("completed all executions")
	},
}

func init() {
	rootCmd.AddCommand(producerContractCmd)

	producerContractCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	producerContractCmd.Flags().StringVar(&dataDir, "dataDir", "", "data dir to store api test scenarios")
	producerContractCmd.Flags().StringVar(&group, "group", "", "group of service APIs")
	producerContractCmd.Flags().StringVar(&baseURL, "base_url", "", "base-url for remote service")
	producerContractCmd.Flags().IntVar(&executionTimes, "times", 10, "execution times")
	producerContractCmd.Flags().BoolVar(&verbose, "verbose", false, "verbose logging")
	producerContractCmd.Flags().StringVar(&scenarioFile, "scenario", "", "path to scenario file (YAML)")
}

// loadScenarioKeyData loads a scenario file and creates an APIKeyData object
func loadScenarioKeyData(filename string) (*types.APIKeyData, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read scenario file: %w", err)
	}

	keyData := &types.APIKeyData{}
	err = yaml.Unmarshal(data, keyData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse scenario file: %w", err)
	}

	return keyData, nil
}
