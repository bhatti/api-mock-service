package cmd

import (
	"fmt"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/spf13/cobra"
)

var (
	shortened  = false
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Version will output the current build information",
		Long:  ``,
		Run: func(_ *cobra.Command, _ []string) {
			versionOutput := types.NewVersion(Version, Commit, Date)

			if shortened {
				fmt.Printf("%+v", versionOutput.ToShortened())
			} else {
				fmt.Printf("%+v", versionOutput.ToJSON())
			}
		},
	}
)

func init() {
	versionCmd.Flags().BoolVarP(&shortened, "short", "s", false, "Prints the version number.")
	rootCmd.AddCommand(versionCmd)
}
