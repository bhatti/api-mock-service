package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/oapi"
	"github.com/bhatti/api-mock-service/internal/speccompare"
	"github.com/bhatti/api-mock-service/internal/types"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var baseSpecFile string
var headSpecFile string
var failOnBreaking bool
var outputJSON bool

// compareSpecsCmd compares two OpenAPI specs and reports breaking changes.
var compareSpecsCmd = &cobra.Command{
	Use:   "compare-specs",
	Short: "Compare two OpenAPI specs for breaking changes",
	Long: `Compares two OpenAPI specs (base vs head) and reports breaking and non-breaking changes.

Breaking changes include: removed paths/methods, added required parameters,
type changes on existing fields, removed response fields, narrowed enums.

Exit code 1 is returned when breaking changes are found and --fail-on-breaking is set.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if baseSpecFile == "" {
			return fmt.Errorf("--base is required")
		}
		if headSpecFile == "" {
			return fmt.Errorf("--head is required")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		baseData, err := os.ReadFile(baseSpecFile)
		if err != nil {
			log.Errorf("failed to read base spec %s: %s", baseSpecFile, err)
			os.Exit(1)
		}
		headData, err := os.ReadFile(headSpecFile)
		if err != nil {
			log.Errorf("failed to read head spec %s: %s", headSpecFile, err)
			os.Exit(1)
		}

		serverConfig, err := types.NewConfiguration(httpPort, proxyPort, dataDir, types.NewVersion(Version, Commit, Date))
		if err != nil {
			log.Errorf("failed to create config: %s", err)
			os.Exit(1)
		}
		dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 1)

		_, _, baseDoc, err := oapi.Parse(context.Background(), serverConfig, baseData, dataTemplate)
		if err != nil {
			log.Errorf("failed to parse base spec: %s", err)
			os.Exit(1)
		}
		_, _, headDoc, err := oapi.Parse(context.Background(), serverConfig, headData, dataTemplate)
		if err != nil {
			log.Errorf("failed to parse head spec: %s", err)
			os.Exit(1)
		}

		report := speccompare.Diff(baseDoc, headDoc)

		if outputJSON {
			b, _ := json.MarshalIndent(report, "", "  ")
			fmt.Println(string(b))
		} else {
			printSpecDiffReport(report)
		}

		if failOnBreaking && report.HasBreakingChanges() {
			os.Exit(2)
		}
	},
}

func init() {
	rootCmd.AddCommand(compareSpecsCmd)

	compareSpecsCmd.Flags().StringVar(&baseSpecFile, "base", "", "path to base OpenAPI spec (YAML/JSON)")
	compareSpecsCmd.Flags().StringVar(&headSpecFile, "head", "", "path to head OpenAPI spec to compare against base (YAML/JSON)")
	compareSpecsCmd.Flags().BoolVar(&failOnBreaking, "fail-on-breaking", false, "exit with code 2 when breaking changes are detected (useful for CI)")
	compareSpecsCmd.Flags().BoolVar(&outputJSON, "json", false, "output report as JSON instead of human-readable table")
	compareSpecsCmd.Flags().StringVar(&dataDir, "dataDir", "", "data dir (for config)")
}

// printSpecDiffReport prints a human-friendly spec diff report.
func printSpecDiffReport(report *speccompare.SpecDiffReport) {
	sep := "──────────────────────────────────────────────────────────────"
	fmt.Printf("\n%s\n", colorize("SPEC DIFF REPORT", ansiBold))
	fmt.Println(colorize(sep, ansiBold))

	if len(report.AddedPaths) > 0 {
		fmt.Printf("\n%s\n", colorize("Added paths (non-breaking):", ansiGreen))
		for _, p := range report.AddedPaths {
			fmt.Printf("  + %s\n", p)
		}
	}
	if len(report.RemovedPaths) > 0 {
		fmt.Printf("\n%s\n", colorize("Removed paths (BREAKING):", ansiRed))
		for _, p := range report.RemovedPaths {
			fmt.Printf("  - %s\n", p)
		}
	}
	if len(report.BreakingChanges) > 0 {
		fmt.Printf("\n%s\n", colorize("Breaking changes:", ansiRed))
		for _, c := range report.BreakingChanges {
			fmt.Printf("  ✗ [%s %s] %s: %s → %s\n",
				c.Method, c.Path, c.ChangeType, c.Before, c.After)
		}
	}
	if len(report.NonBreakingChanges) > 0 {
		fmt.Printf("\n%s\n", colorize("Non-breaking changes:", ansiYellow))
		for _, c := range report.NonBreakingChanges {
			fmt.Printf("  ~ [%s %s] %s\n", c.Method, c.Path, c.ChangeType)
		}
	}

	fmt.Println(colorize(sep, ansiBold))
	summary := report.Summary()
	if report.HasBreakingChanges() {
		fmt.Println(colorize(summary, ansiRed))
	} else {
		fmt.Println(colorize(summary, ansiGreen))
	}
	fmt.Println()
}
