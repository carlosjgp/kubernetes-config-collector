package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version = "v0.0.1"
	Short   bool
)

func init() {
	rootCmd.PersistentFlags().BoolVar(&Short, "short", false, "verbose output")

	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of config-collector",
	Long:  `All software has versions. This is config-collector's`,
	Run: func(cmd *cobra.Command, args []string) {
		var message string
		if Short {
			message = "%s"
		} else {
			message = "Kubernetes config collector %s -- HEAD"
		}
		fmt.Printf(message, Version)
	},
}
