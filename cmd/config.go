package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"os"
)

var (
	jsonConfig = false
	configCmd  = &cobra.Command{
		Use:   "config",
		Short: "config will output the current configuration",
		Long:  ``,
		Run: func(_ *cobra.Command, _ []string) {
			serverConfig, err := types.NewConfiguration(httpPort, proxyPort, dataDir, types.NewVersion(Version, Commit, Date))
			if err != nil {
				log.WithFields(log.Fields{
					"Error": err}).
					Errorf("Failed to parse config...")
				os.Exit(1)
			}

			if jsonConfig {
				b, err := json.Marshal(serverConfig)
				if err != nil {
					log.WithFields(log.Fields{
						"Error": err}).
						Errorf("Failed to marshal config...")
					os.Exit(2)
				}
				fmt.Printf("%s", b)
			} else {
				b, err := yaml.Marshal(serverConfig)
				if err != nil {
					log.WithFields(log.Fields{
						"Error": err}).
						Errorf("Failed to marshal config...")
					os.Exit(3)
				}
				fmt.Printf("%s", b)
			}
		},
	}
)

func init() {
	configCmd.Flags().BoolVarP(&shortened, "json", "j", false, "JSON format.")
	rootCmd.AddCommand(configCmd)
}
