package cmd

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

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
var targetUrl string
var requestBody string
var requestBodyFile string
var queryParams []string

// consumerContractCmd represents the consumer contract command
var consumerContractCmd = &cobra.Command{
	Use:   "consumer-contract",
	Short: "Executes consumer contracts",
	Long:  "Executes consumer contracts by simulating API requests or sending real HTTP requests",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Validate input parameters
		if scenarioFile == "" && targetUrl == "" {
			// If neither scenarioFile nor URL is provided, method and path are required
			if method == "" {
				return fmt.Errorf("HTTP method is required when neither scenario file nor URL is provided")
			}
			if path == "" {
				return fmt.Errorf("URI path is required when neither scenario file nor URL is provided")
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.WithFields(log.Fields{
			"DataDir":         dataDir,
			"Method":          method,
			"Name":            name,
			"Path":            path,
			"ScenarioFile":    scenarioFile,
			"Headers":         headerFlags,
			"URL":             targetUrl,
			"RequestBody":     requestBody != "",
			"RequestBodyFile": requestBodyFile != "",
			"QueryParams":     queryParams,
		}).Debugf("executing consumer contract...")

		// If URL is provided, send a real HTTP request
		if targetUrl != "" {
			executeRealRequest()
			return
		}

		// Otherwise use consumer executor for mock testing
		executeMockRequest()
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
	consumerContractCmd.Flags().StringVar(&targetUrl, "url", "", "URL to send actual HTTP request instead of using consumer executor")
	consumerContractCmd.Flags().StringVar(&requestBody, "body", "", "request body content")
	consumerContractCmd.Flags().StringVar(&requestBodyFile, "body-file", "", "path to file containing request body content")
	consumerContractCmd.Flags().StringSliceVar(&queryParams, "query", []string{}, "query parameters in format 'key=value'")
}

// executeRealRequest builds and sends an actual HTTP request to the specified URL
func executeRealRequest() {
	// Get request body
	var bodyContent []byte
	var err error

	if requestBodyFile != "" {
		bodyContent, err = os.ReadFile(requestBodyFile)
		if err != nil {
			log.Errorf("failed to read request body file: %s", err)
			os.Exit(1)
		}
	} else if requestBody != "" {
		bodyContent = []byte(requestBody)
	}

	// Determine method and path from scenario file if not provided
	methodToUse := method
	pathToUse := path

	if scenarioFile != "" && (method == "" || path == "") {
		data, err := os.ReadFile(scenarioFile)
		if err != nil {
			log.Errorf("failed to read scenario file: %s", err)
			os.Exit(2)
		}

		keyData := &types.APIScenario{}
		if err = yaml.Unmarshal(data, keyData); err != nil {
			log.Errorf("failed to parse scenario file: %s", err)
			os.Exit(3)
		}

		if method == "" && keyData.Method != "" {
			methodToUse = string(keyData.Method)
		}

		if path == "" && keyData.Path != "" {
			pathToUse = keyData.Path
		}
		if requestBody == "" && requestBodyFile == "" && keyData.Request.Contents != "" {
			requestBody = keyData.Request.Contents
		}
		// If scenario file has content pattern, use it as body if no body specified
		//if len(bodyContent) == 0 && keyData.Request.AssertContentsPattern != "" {
		//	bodyContent = []byte(keyData.Request.AssertContentsPattern)
		//}
	}

	if methodToUse == "" {
		log.Errorf("HTTP method must be specified")
		os.Exit(4)
	}

	// Build the full URL
	fullURL := targetUrl
	if pathToUse != "" && !strings.HasSuffix(targetUrl, pathToUse) {
		// Add path to URL if it's not already there
		if !strings.HasSuffix(targetUrl, "/") && !strings.HasPrefix(pathToUse, "/") {
			fullURL += "/"
		}
		fullURL += pathToUse
	}

	// Parse URL and add query parameters
	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		log.Errorf("failed to parse URL: %s", err)
		os.Exit(5)
	}

	q := parsedURL.Query()
	for _, param := range queryParams {
		parts := strings.SplitN(param, "=", 2)
		if len(parts) == 2 {
			q.Add(parts[0], parts[1])
		}
	}
	parsedURL.RawQuery = q.Encode()

	// Create and send the HTTP request
	req, err := http.NewRequest(
		strings.ToUpper(methodToUse),
		parsedURL.String(),
		bytes.NewBuffer(bodyContent),
	)
	if err != nil {
		log.Errorf("failed to create HTTP request: %s", err)
		os.Exit(6)
	}

	// Add headers
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

	// Set content type if not already set and body is present
	if len(bodyContent) > 0 && req.Header.Get(types.ContentTypeHeader) == "" {
		req.Header.Set(types.ContentTypeHeader, "application/json")
	}

	// Send the request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	startTime := time.Now()
	resp, err := client.Do(req)
	endTime := time.Now()
	if err != nil {
		log.Errorf("failed to send HTTP request: %s", err)
		os.Exit(7)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("failed to read response body: %s", err)
		os.Exit(8)
	}

	// Print response
	fmt.Printf("Request sent to: %s\n", parsedURL.String())
	fmt.Printf("Method: %s\n", req.Method)
	fmt.Printf("Status: %d %s\n", resp.StatusCode, http.StatusText(resp.StatusCode))
	fmt.Printf("Time: %v\n", endTime.Sub(startTime))

	fmt.Println("\nRequest Headers:")
	for k, v := range req.Header {
		fmt.Printf("  %s: %s\n", k, strings.Join(v, ", "))
	}

	if len(bodyContent) > 0 {
		fmt.Println("\nRequest Body:")
		fmt.Println(string(bodyContent))
	}

	fmt.Println("\nResponse Headers:")
	for k, v := range resp.Header {
		fmt.Printf("  %s: %s\n", k, strings.Join(v, ", "))
	}

	fmt.Println("\nResponse Body:")
	fmt.Println(string(respBody))

	log.Infof("HTTP request completed successfully")
}

// executeMockRequest uses the consumer executor to simulate an API request
func executeMockRequest() {
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

	// Create a mock HTTP request with body if specified
	body, err := getRequestBody()
	if err != nil {
		log.Errorf("failed to get request body: %s", err)
		os.Exit(3)
	}

	req, err := createMockRequest(body)
	if err != nil {
		log.Errorf("failed to create mock request: %s", err)
		os.Exit(4)
	}

	// Create key data from request parameters
	keyData, err := createKeyData()
	if err != nil {
		log.Errorf("failed to create key data: %s", err)
		os.Exit(5)
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
		os.Exit(6)
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
}

// getRequestBody reads the request body from either the --body flag or --body-file flag
func getRequestBody() ([]byte, error) {
	if requestBodyFile != "" {
		return os.ReadFile(requestBodyFile)
	}
	return []byte(requestBody), nil
}

// createMockRequest creates a mock HTTP request with the specified parameters
func createMockRequest(body []byte) (*http.Request, error) {
	// Use values from scenario file if method or path were not specified
	methodToUse := method
	pathToUse := path

	if scenarioFile != "" && (method == "" || path == "") {
		// Try to extract method and path from scenario file
		data, err := os.ReadFile(scenarioFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read scenario file: %w", err)
		}

		keyData := &types.APIScenario{}
		if err = yaml.Unmarshal(data, keyData); err != nil {
			return nil, fmt.Errorf("failed to parse scenario file: %w", err)
		}

		if method == "" && keyData.Method != "" {
			methodToUse = string(keyData.Method)
		}

		if path == "" && keyData.Path != "" {
			pathToUse = keyData.Path
		}
		if requestBody == "" && requestBodyFile == "" {
			requestBody = keyData.Request.Contents
		}
	}

	if methodToUse == "" || pathToUse == "" {
		return nil, fmt.Errorf("method and path must be specified either via command line or scenario file")
	}

	// Create request URL with query parameters
	reqURL, err := url.Parse(fmt.Sprintf("http://localhost:8080%s", pathToUse))
	if err != nil {
		return nil, err
	}

	q := reqURL.Query()
	for _, param := range queryParams {
		parts := strings.SplitN(param, "=", 2)
		if len(parts) == 2 {
			q.Add(parts[0], parts[1])
		}
	}
	reqURL.RawQuery = q.Encode()

	// Create request with body
	req, err := http.NewRequest(
		strings.ToUpper(methodToUse),
		reqURL.String(),
		bytes.NewBuffer(body),
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

	// Set content type if not already set and body is present
	if len(body) > 0 && req.Header.Get(types.ContentTypeHeader) == "" {
		req.Header.Set(types.ContentTypeHeader, "application/json")
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
		scenario := &types.APIScenario{}
		err = yaml.Unmarshal(data, scenario)

		// If command line parameters are provided, they override the file
		if err == nil {
			keyData = &types.APIKeyData{}
		} else {
			keyData = scenario.ToKeyData()
			if method != "" {
				httpMethod, err := types.ToMethod(method)
				if err != nil {
					return nil, err
				}
				keyData.Method = httpMethod
			}

			if path != "" {
				keyData.Path = path
			}

			if name != "" {
				keyData.Name = name
			}

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

			// Add body if provided
			if requestBody != "" || requestBodyFile != "" {
				body, err := getRequestBody()
				if err == nil && len(body) > 0 {
					keyData.AssertContentsPattern = string(body)
				}
			} else {
				requestBody = scenario.Request.Contents
			}

			// Add query parameters
			if len(queryParams) > 0 {
				if keyData.AssertQueryParamsPattern == nil {
					keyData.AssertQueryParamsPattern = make(map[string]string)
				}

				for _, param := range queryParams {
					parts := strings.SplitN(param, "=", 2)
					if len(parts) == 2 {
						keyData.AssertQueryParamsPattern[parts[0]] = parts[1]
					}
				}
			}
		}

		return keyData, err
	}

	// Get body if provided
	body, err := getRequestBody()
	if err != nil {
		return nil, err
	}

	// Create from command line parameters
	httpMethod, err := types.ToMethod(method)
	if err != nil {
		return nil, err
	}

	keyData = &types.APIKeyData{
		Method:                   httpMethod,
		Name:                     name,
		Path:                     path,
		AssertQueryParamsPattern: make(map[string]string),
		AssertHeadersPattern:     make(map[string]string),
		AssertContentsPattern:    string(body),
	}

	// Add headers to assertion patterns
	for _, headerFlag := range headerFlags {
		parts := strings.SplitN(headerFlag, ":", 2)
		if len(parts) == 2 {
			keyData.AssertHeadersPattern[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	// Add query parameters
	for _, param := range queryParams {
		parts := strings.SplitN(param, "=", 2)
		if len(parts) == 2 {
			keyData.AssertQueryParamsPattern[parts[0]] = parts[1]
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

	// Add query parameters to overrides
	for _, param := range queryParams {
		parts := strings.SplitN(param, "=", 2)
		if len(parts) == 2 {
			overrides[parts[0]] = parts[1]
		}
	}

	return overrides
}
