package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/bhatti/api-mock-service/internal/pm"
	"github.com/bhatti/api-mock-service/internal/types"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var postmanFile string

// importPostmanCmd represents the import-postman command
var importPostmanCmd = &cobra.Command{
	Use:   "import-postman",
	Short: "Imports Postman collections and converts them to API scenarios",
	Long:  "Imports Postman collections from a file and converts them to API scenarios that can be used for testing",
	Run: func(cmd *cobra.Command, args []string) {
		log.WithFields(log.Fields{
			"DataDir":     dataDir,
			"PostmanFile": postmanFile,
		}).Infof("importing Postman collection...")

		if postmanFile == "" {
			log.Errorf("Postman collection file path is required")
			os.Exit(1)
		}

		// Read the Postman collection file
		data, err := os.ReadFile(postmanFile)
		if err != nil {
			log.Errorf("failed to read Postman collection file: %s", err)
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
		scenarioRepo, _, _, _, err := buildRepos(serverConfig)
		if err != nil {
			log.Errorf("failed to setup repositories: %s", err)
			os.Exit(4)
		}

		// Parse the Postman collection
		collection := &pm.PostmanCollection{}
		err = json.Unmarshal(data, collection)
		if err != nil {
			log.Errorf("failed to parse Postman collection: %s", err)
			os.Exit(5)
		}

		// Convert Postman collection to API scenarios
		scenarios, vars := pm.ConvertPostmanToScenarios(serverConfig, collection, time.Time{}, time.Time{})

		// Save scenarios and history
		successCount := 0
		for _, scenario := range scenarios {
			err = scenarioRepo.Save(scenario)
			if err != nil {
				log.Warnf("failed to save scenario for %s: %s", scenario.Name, err)
				continue
			}

			// Save history
			u, err := scenario.GetURL("")
			if err == nil {
				err = scenarioRepo.SaveHistory(scenario, u.String(), scenario.StartTime, scenario.EndTime)
				if err != nil {
					log.Warnf("failed to save history for %s: %s", scenario.Name, err)
				}
			}

			successCount++
			fmt.Printf("Imported scenario - Method: %s, Path: %s, Name: %s\n",
				scenario.Method, scenario.Path, scenario.Name)
		}

		// Save variables if any
		if len(vars.Variables) > 0 {
			err = scenarioRepo.SaveVariables(vars)
			if err != nil {
				log.Warnf("failed to save variables: %s", err)
			} else {
				fmt.Printf("Saved %d variables from Postman collection\n", len(vars.Variables))
			}
		}

		// Print summary
		fmt.Printf("\nImport Summary:\n")
		fmt.Printf("Total requests in collection: %d\n", len(scenarios))
		fmt.Printf("Successfully imported scenarios: %d\n", successCount)
		fmt.Printf("Failed scenarios: %d\n", len(scenarios)-successCount)
		fmt.Printf("Variables imported: %d\n", len(vars.Variables))

		log.WithFields(log.Fields{
			"Total":        len(scenarios),
			"Succeeded":    successCount,
			"Failed":       len(scenarios) - successCount,
			"VarsImported": len(vars.Variables),
		}).Infof("completed Postman collection import")
	},
}

func init() {
	rootCmd.AddCommand(importPostmanCmd)

	importPostmanCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	importPostmanCmd.Flags().StringVar(&dataDir, "dataDir", "", "data dir to store API contracts and fixtures")
	importPostmanCmd.Flags().StringVar(&postmanFile, "file", "", "path to Postman collection file")

	_ = importPostmanCmd.MarkFlagRequired("file")
}
