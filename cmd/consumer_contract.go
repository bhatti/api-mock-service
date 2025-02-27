package cmd

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"net/http"
	"os"
	"strings"

	"github.com/bhatti/api-mock-service/internal/contract"
	"github.com/bhatti/api-mock-service/internal/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var method string
var name string
var path string
var scenarioFile string
var headerFlags []string

// consumerContractCmd represents the consumer contract command
var consumerContractCmd = &cobra.Command{
	Use:   "consumer-contract",
	Short: "Executes consumer contracts",
	Long:  "Executes consumer contracts by simulating API requests",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Validate input parameters
		if scenarioFile == "" {
			// If scenarioFile is not provided, method and path are required
			if method == "" {
				return fmt.Errorf("HTTP method is required when scenario file is not provided")
			}
			if path == "" {
				return fmt.Errorf("URI path is required when scenario file is not provided")
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.WithFields(log.Fields{
			"DataDir":      dataDir,
			"Method":       method,
			"Name":         name,
			"Path":         path,
			"ScenarioFile": scenarioFile,
			"Headers":      headerFlags,
		}).Infof("executing consumer contract...")

		serverConfig, err := types.NewConfiguration(
			httpPort,
			proxyPort,
			dataDir,
			types.NewVersion(Version, Commit, Date))
		if err != nil {
			log.Errorf("failed to parse config: %s", err)
			os.Exit(1)
		}

		scenarioRepo, fixturesRepo, _, groupConfigRepo, err := buildRepos(serverConfig)
		if err != nil {
			log.Errorf("failed to setup repositories: %s", err)
			os.Exit(2)
		}

		// Create consumer executor
		consumerExecutor := contract.NewConsumerExecutor(
			serverConfig,
			scenarioRepo,
			fixturesRepo,
			groupConfigRepo,
		)

		// Create a mock HTTP request
		req, err := createMockRequest()
		if err != nil {
			log.Errorf("failed to create mock request: %s", err)
			os.Exit(3)
		}

		// Create key data from request parameters
		keyData, err := createKeyData()
		if err != nil {
			log.Errorf("failed to create key data: %s", err)
			os.Exit(4)
		}

		// Create response headers
		respHeaders := make(http.Header)

		// Call ExecuteWithKey which works for both web and command line
		matchedScenario, respBody, sharedVars, err := consumerExecutor.ExecuteWithKey(
			req,
			respHeaders,
			keyData,
			createOverrides())

		if err != nil {
			log.Errorf("consumer contract execution failed: %s", err)
			os.Exit(5)
		}

		// Print response
		fmt.Printf("Status: %d\n", matchedScenario.Response.StatusCode)
		fmt.Println("Content-Type:", matchedScenario.Response.ContentType(""))

		if len(sharedVars) > 0 {
			fmt.Println("\nShared Variables:")
			for k, v := range sharedVars {
				fmt.Printf("  %s: %v\n", k, v)
			}
		}

		fmt.Println("\nResponse Headers:")
		for k, v := range respHeaders {
			fmt.Printf("  %s: %s\n", k, v)
		}

		fmt.Println("\nBody:")
		fmt.Println(string(respBody))

		log.Infof("consumer contract execution completed successfully")
	},
}

func init() {
	rootCmd.AddCommand(consumerContractCmd)

	consumerContractCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	consumerContractCmd.Flags().StringVar(&dataDir, "dataDir", "", "data dir to store API contracts and fixtures")
	consumerContractCmd.Flags().StringVar(&method, "method", "", "HTTP method (GET, POST, PUT, DELETE, etc.)")
	consumerContractCmd.Flags().StringVar(&name, "name", "", "scenario name")
	consumerContractCmd.Flags().StringVar(&path, "path", "", "URI path for the API request")
	consumerContractCmd.Flags().StringVar(&scenarioFile, "scenario", "", "path to scenario file (YAML)")
	consumerContractCmd.Flags().StringSliceVar(&headerFlags, "header", []string{}, "HTTP headers in format 'key:value'")

	// No required flags since we now check requirements in PreRunE
}

// createMockRequest creates a mock HTTP request with the specified parameters
func createMockRequest() (*http.Request, error) {
	// Use values from scenario file if method or path were not specified
	methodToUse := method
	pathToUse := path

	if scenarioFile != "" && (method == "" || path == "") {
		// Try to extract method and path from scenario file
		data, err := os.ReadFile(scenarioFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read scenario file: %w", err)
		}

		keyData := &types.APIKeyData{}
		if err = yaml.Unmarshal(data, keyData); err != nil {
			return nil, fmt.Errorf("failed to parse scenario file: %w", err)
		}

		if method == "" && keyData.Method != "" {
			methodToUse = string(keyData.Method)
		}

		if path == "" && keyData.Path != "" {
			pathToUse = keyData.Path
		}
	}

	if methodToUse == "" || pathToUse == "" {
		return nil, fmt.Errorf("method and path must be specified either via command line or scenario file")
	}

	req, err := http.NewRequest(
		strings.ToUpper(methodToUse),
		fmt.Sprintf("http://localhost%s", pathToUse),
		nil,
	)
	if err != nil {
		return nil, err
	}

	// Add headers from command line
	for _, headerFlag := range headerFlags {
		parts := strings.SplitN(headerFlag, ":", 2)
		if len(parts) == 2 {
			req.Header.Add(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}

	// Add scenario name to header if provided
	if name != "" {
		req.Header.Set(types.MockScenarioHeader, name)
	}

	return req, nil
}

// createKeyData creates APIKeyData from command line parameters
func createKeyData() (keyData *types.APIKeyData, err error) {
	// Load from scenario file if specified
	if scenarioFile != "" {
		data, err := os.ReadFile(scenarioFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read scenario file: %w", err)
		}
		keyData = &types.APIKeyData{}
		err = yaml.Unmarshal(data, keyData)

		// If command line parameters are provided, they override the file
		if err == nil {
			// Add headers from command line
			if keyData.AssertHeadersPattern == nil {
				keyData.AssertHeadersPattern = make(map[string]string)
			}

			for _, headerFlag := range headerFlags {
				parts := strings.SplitN(headerFlag, ":", 2)
				if len(parts) == 2 {
					keyData.AssertHeadersPattern[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
				}
			}
		}

		return keyData, err
	}

	// Create from command line parameters
	httpMethod, err := types.ToMethod(method)
	if err != nil {
		return nil, err
	}

	keyData = &types.APIKeyData{
		Method:               httpMethod,
		Name:                 name,
		Path:                 path,
		AssertHeadersPattern: make(map[string]string),
	}

	// Add headers to assertion patterns
	for _, headerFlag := range headerFlags {
		parts := strings.SplitN(headerFlag, ":", 2)
		if len(parts) == 2 {
			keyData.AssertHeadersPattern[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	return keyData, nil
}

// createOverrides creates a map of overrides from command line parameters
func createOverrides() map[string]any {
	overrides := make(map[string]any)

	// Add name if provided
	if name != "" {
		overrides[types.MockScenarioHeader] = name
	}

	// Add headers to overrides
	for _, headerFlag := range headerFlags {
		parts := strings.SplitN(headerFlag, ":", 2)
		if len(parts) == 2 {
			overrides[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	return overrides
}
