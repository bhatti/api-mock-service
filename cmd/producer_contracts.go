package cmd

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v3"
	"net/http"
	"os"

	"github.com/bhatti/api-mock-service/internal/contract"
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/oapi"
	"github.com/bhatti/api-mock-service/internal/shrink"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/web"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var group string
var baseURL string
var executionTimes int
var verbose bool
var specFile string
var trackCoverage bool
var runMutations bool
var dryRun bool
var runShrink bool

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
			"SpecFile":     specFile,
			"Mutations":    runMutations,
			"Coverage":     trackCoverage,
			"DryRun":       dryRun,
			"Shrink":       runShrink,
		}).Debugf("executing producer contracts...")

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
		contractReq.TrackCoverage = trackCoverage
		contractReq.DryRun = dryRun

		executor := contract.NewProducerExecutor(
			scenarioRepo,
			groupConfigRepo,
			web.NewHTTPClient(serverConfig, web.NewAuthAdapter(serverConfig)),
		)

		// If an OpenAPI spec is provided, wire it up for response schema validation.
		if specFile != "" {
			specData, err := os.ReadFile(specFile)
			if err != nil {
				log.Errorf("failed to read spec file %s: %s", specFile, err)
				os.Exit(4)
			}
			_, _, doc, err := oapi.Parse(context.Background(), serverConfig, specData, dataTemplate)
			if err != nil {
				log.Errorf("failed to parse spec file %s: %s", specFile, err)
				os.Exit(5)
			}
			router, err := oapi.BuildRouter(doc)
			if err != nil {
				log.Errorf("failed to build router from spec file %s: %s", specFile, err)
				os.Exit(6)
			}
			executor = executor.WithOpenAPISpec(doc, router)
			log.Infof("OpenAPI spec loaded from %s — responses will be validated against the schema", specFile)
		}

		var contractRes *types.ProducerContractResponse

		if runMutations {
			// Run mutation testing for the group
			if group == "" {
				log.Errorf("--group is required when using --mutations")
				os.Exit(7)
			}
			contractRes = executor.ExecuteMutationsByGroup(context.Background(), &http.Request{}, group, dataTemplate, contractReq)
		} else if scenarioFile != "" {
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

		printContractResultsTable(contractRes)
		if contractRes.Coverage != nil {
			printCoverageReport(contractRes.Coverage)
		}

		// If shrinking is requested, find the minimal failing payload for each failure.
		if runShrink && len(contractRes.Errors) > 0 && group != "" {
			fmt.Printf("\n%s\n", colorize("SHRINK ANALYSIS", ansiBold))
			detector := contract.NewProducerExecutorFailureDetector(executor, baseURL, dataTemplate, contractReq)
			scenarioKeys := scenarioRepo.LookupAllByGroup(group)
			for _, sk := range scenarioKeys {
				if _, isFailed := contractRes.Errors[sk.Name+"_0"]; !isFailed {
					// Check without index suffix too
					if _, isFailed2 := contractRes.Errors[sk.Name]; !isFailed2 {
						continue
					}
				}
				scenario, lookupErr := scenarioRepo.Lookup(sk, contractReq.Overrides())
				if lookupErr != nil {
					continue
				}
				fmt.Printf("Shrinking %s ...\n", sk.Name)
				shrinkResult, shrinkErr := shrink.Shrink(context.Background(), detector, scenario, shrink.ShrinkOptions{})
				if shrinkErr != nil {
					fmt.Printf("  error: %s\n", shrinkErr)
					continue
				}
				if shrinkResult.Reduced {
					fmt.Printf("  %s reduced in %d attempts\n", colorize("✓", ansiGreen), shrinkResult.Attempts)
					fmt.Printf("  Minimal body: %s\n", shrinkResult.Minimal.Request.Contents)
				} else {
					fmt.Printf("  %s no reduction possible (%d attempts)\n", colorize("~", ansiYellow), shrinkResult.Attempts)
				}
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
	producerContractCmd.Flags().StringVar(&specFile, "spec", "", "path to OpenAPI spec file (YAML/JSON) for response schema validation")
	producerContractCmd.Flags().BoolVar(&trackCoverage, "track-coverage", false, "include OpenAPI coverage report in output (requires --spec)")
	producerContractCmd.Flags().BoolVar(&runMutations, "mutations", false, "run mutation testing instead of normal contract execution")
	producerContractCmd.Flags().BoolVar(&dryRun, "dry-run", false, "list scenarios that would run without executing them")
	producerContractCmd.Flags().BoolVar(&runShrink, "shrink", false, "shrink failing mutation payloads to minimal reproducing inputs (requires failures)")
}

// isTTY returns true when stdout is a terminal (ANSI colors are safe to use).
func isTTY() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

const (
	ansiReset  = "\033[0m"
	ansiRed    = "\033[31m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiCyan   = "\033[36m"
	ansiBold   = "\033[1m"
)

func colorize(s, color string) string {
	if !isTTY() {
		return s
	}
	return color + s + ansiReset
}

// printContractResultsTable prints a human-friendly table of contract execution results.
func printContractResultsTable(res *types.ProducerContractResponse) {
	sep := "──────────────────────────────────────────────────────────────"
	fmt.Printf("\n%s\n", colorize(sep, ansiBold))
	fmt.Printf("%-40s %-10s %s\n",
		colorize("SCENARIO", ansiBold),
		colorize("STATUS", ansiBold),
		colorize("LATENCY", ansiBold))
	fmt.Println(colorize(sep, ansiBold))

	// Print successes
	for name := range res.Results {
		if _, failed := res.Errors[name]; !failed {
			fmt.Printf("%-40s %s\n", truncate(name, 40), colorize("✓ PASS", ansiGreen))
		}
	}
	// Print failures with detail
	for name, errMsg := range res.Errors {
		fmt.Printf("%-40s %s\n", truncate(name, 40), colorize("✗ FAIL", ansiRed))
		if detail, ok := res.ErrorDetails[name]; ok {
			if len(detail.MissingFields) > 0 {
				fmt.Printf("  %s %v\n", colorize("Missing:", ansiYellow), detail.MissingFields)
			}
			if len(detail.ValueMismatches) > 0 {
				for k, v := range detail.ValueMismatches {
					fmt.Printf("  %s %s (expected %v, got %v)\n",
						colorize("Mismatch:", ansiYellow), k, v.Expected, v.Actual)
				}
			}
			if len(detail.SchemaViolations) > 0 {
				for _, sv := range detail.SchemaViolations {
					fmt.Printf("  %s %s\n", colorize("Schema:", ansiRed), sv.Message)
				}
			}
		} else {
			fmt.Printf("  %s\n", colorize(truncate(errMsg, 100), ansiRed))
		}
	}

	fmt.Println(colorize(sep, ansiBold))
	total := res.Succeeded + res.Failed + res.Mismatched
	summary := fmt.Sprintf("TOTAL %d  Passed: %d  Failed: %d  Mismatched: %d",
		total, res.Succeeded, res.Failed, res.Mismatched)
	if res.Failed > 0 {
		summary = colorize(summary, ansiRed)
	} else {
		summary = colorize(summary, ansiGreen)
	}
	fmt.Println(summary)

	if len(res.URLs) > 0 && verbose {
		fmt.Printf("\nURLs: ")
		for u, n := range res.URLs {
			fmt.Printf("%s (%dx)  ", u, n)
		}
		fmt.Println()
	}
}

// printCoverageReport prints the coverage summary.
func printCoverageReport(c *types.CoverageSummary) {
	sep := "──────────────────────────────────────────────────────────────"
	fmt.Printf("\n%s\n", colorize("COVERAGE REPORT", ansiBold))
	fmt.Println(colorize(sep, ansiBold))
	coverageStr := fmt.Sprintf("Overall: %.1f%%  (%d/%d paths)",
		c.Coverage, c.CoveredPaths, c.TotalPaths)
	if c.Coverage < 80 {
		fmt.Println(colorize(coverageStr, ansiYellow))
	} else {
		fmt.Println(colorize(coverageStr, ansiGreen))
	}
	if len(c.UncoveredPaths) > 0 {
		fmt.Printf("\n%s\n", colorize("Uncovered paths:", ansiYellow))
		for _, p := range c.UncoveredPaths {
			fmt.Printf("  ✗ %s\n", p)
		}
	}
	if len(c.MethodCoverage) > 0 {
		fmt.Printf("\n%s\n", colorize("Method coverage:", ansiBold))
		for method, pct := range c.MethodCoverage {
			bar := colorize(fmt.Sprintf("%-6s %.1f%%", method, pct), ansiCyan)
			fmt.Printf("  %s\n", bar)
		}
	}
	if len(c.FieldCoverageByScenario) > 0 && verbose {
		fmt.Printf("\n%s\n", colorize("Field coverage by scenario:", ansiBold))
		for scenario, pct := range c.FieldCoverageByScenario {
			fmt.Printf("  %-40s %.1f%%\n", truncate(scenario, 40), pct)
		}
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
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
