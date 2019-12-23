package cmd

import (
	"os"

	"github.com/carlosjgp/kubernetes-config-collector/app"
	"github.com/carlosjgp/kubernetes-config-collector/client"
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
	rootCmd.PersistentFlags().StringArrayVarP(&config.Labels, "labels", "l", []string{"config-collector.k8s.io=true"}, "Labels that should be used for filtering")
	rootCmd.PersistentFlags().StringVar(&config.FolderAnnotation, "folder-annotation", "fa", "The annotation the sidecar will look for in configmaps to override the destination folder for files, defaults to \"k8s-sidecar-target-directory\"")
	rootCmd.PersistentFlags().StringVarP(&config.Folder, "folder", "f", "/tmp", "Folder where the files should be placed")
	rootCmd.PersistentFlags().StringArrayVarP(&config.Labels, "namespaces", "n", []string{}, "List of namespaces from there to collect resources from. Leave empty to look into all the namespaces")
	rootCmd.PersistentFlags().BoolVarP(&config.ConfigMaps, "config-maps", "c", true, "Enable to collect ConfigMaps")
	rootCmd.PersistentFlags().BoolVarP(&config.ConfigMaps, "secrets", "s", true, "Enable to collect ConfigMaps")
}
