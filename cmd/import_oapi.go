package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/oapi"
	"github.com/bhatti/api-mock-service/internal/types"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var openAPIFile string

// importOpenAPICmd represents the import-openapi command
var importOpenAPICmd = &cobra.Command{
	Use:   "import-openapi",
	Short: "Imports OpenAPI specifications and converts them to API scenarios",
	Long:  "Imports OpenAPI specifications from a file and converts them to API scenarios that can be used for testing",
	Run: func(cmd *cobra.Command, args []string) {
		log.WithFields(log.Fields{
			"DataDir":     dataDir,
			"OpenAPIFile": openAPIFile,
		}).Infof("importing OpenAPI specs...")

		if openAPIFile == "" {
			log.Errorf("OpenAPI file path is required")
			os.Exit(1)
		}

		// Read the OpenAPI file
		data, err := os.ReadFile(openAPIFile)
		if err != nil {
			log.Errorf("failed to read OpenAPI file: %s", err)
			os.Exit(2)
		}

		// Create server configuration
		serverConfig, err := types.NewConfiguration(
			httpPort,
			proxyPort,
			dataDir,
			types.NewVersion(Version, Commit, Date))
		if err != nil {
			log.Errorf("failed to parse config: %s", err)
			os.Exit(3)
		}

		// Create repositories
		scenarioRepo, _, oapiRepo, _, err := buildRepos(serverConfig)

		if err != nil {
			log.Errorf("failed to setup repositories: %s", err)
			os.Exit(4)
		}

		// Create data template for fuzzing
		dataTemplate := fuzz.NewDataTemplateRequest(true, 1, 1)

		// Parse the OpenAPI specs and create scenarios
		specs, updated, err := oapi.Parse(context.Background(), serverConfig, data, dataTemplate)
		if err != nil {
			log.Errorf("failed to parse OpenAPI specs: %s", err)
			os.Exit(5)
		}

		// Convert specs to API scenarios and save them
		var successCount int
		apiKeyData := make([]*types.APIKeyData, 0)
		for _, spec := range specs {
			scenario, err := spec.BuildMockScenario(dataTemplate)
			if err != nil {
				log.Warnf("failed to build mock scenario for %s: %s", spec.Path, err)
				continue
			}

			err = scenarioRepo.Save(scenario)
			if err != nil {
				log.Warnf("failed to save scenario for %s: %s", spec.Path, err)
				continue
			}

			apiKeyData = append(apiKeyData, scenario.ToKeyData())
			successCount++

			fmt.Printf("Imported scenario - Method: %s, Path: %s, Name: %s\n",
				scenario.Method, scenario.Path, scenario.Name)
		}

		// Save the raw OpenAPI spec
		if len(specs) > 0 {
			title := specs[0].Title
			if title == "" {
				title = "imported-openapi"
			}

			err = oapiRepo.SaveRaw(title, updated)
			if err != nil {
				log.Warnf("failed to save raw OpenAPI specs: %s", err)
			} else {
				fmt.Printf("Saved raw OpenAPI specification as '%s'\n", title)
			}
		}

		// Print summary
		fmt.Printf("\nImport Summary:\n")
		fmt.Printf("Total specs parsed: %d\n", len(specs))
		fmt.Printf("Successfully imported scenarios: %d\n", successCount)
		fmt.Printf("Failed scenarios: %d\n", len(specs)-successCount)

		log.WithFields(log.Fields{
			"Total":     len(specs),
			"Succeeded": successCount,
			"Failed":    len(specs) - successCount,
		}).Infof("completed OpenAPI import")
	},
}

func init() {
	rootCmd.AddCommand(importOpenAPICmd)

	importOpenAPICmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	importOpenAPICmd.Flags().StringVar(&dataDir, "dataDir", "", "data dir to store API contracts and fixtures")
	importOpenAPICmd.Flags().StringVar(&openAPIFile, "file", "", "path to OpenAPI specification file")

	_ = importOpenAPICmd.MarkFlagRequired("file")
}
