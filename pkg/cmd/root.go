package cmd

import (
	"os"

	"github.com/carlosjgp/kubernetes-config-collector/pkg/app"
	"github.com/carlosjgp/kubernetes-config-collector/pkg/client"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var (
	rootCmd = &cobra.Command{
		Use:   "config-collector",
		Short: "An app to collect ConfigMaps",
		Long: `conig-collector will collect each ConfigMap
and will extract each data key as a file on the
configured directory.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientset, err := client.GetClient()
			if err != nil {
				return err
			}
			app.Execute(clientset, config)
			return nil
		},
	}
	config = &app.Config{}
)

// Execute executes the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&config.Verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringArrayVarP(&config.Labels, "labels", "l", []string{"config-collector.io"}, "Labels that should be used for filtering")
	rootCmd.PersistentFlags().StringVarP(&config.FolderAnnotation, "folder-annotation", "a", "config-collector.io/folder", "The annotation the sidecar will look for in configmaps to override the destination folder for files, defaults to \"k8s-sidecar-target-directory\"")
	rootCmd.PersistentFlags().StringVarP(&config.Folder, "folder", "f", "/tmp", "Folder where the files should be placed")
	rootCmd.PersistentFlags().StringArrayVarP(&config.Namespaces, "namespaces", "n", []string{""}, "List of namespaces from there to collect resources from. Leave empty to look into all the namespaces")
}
