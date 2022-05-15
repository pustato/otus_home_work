package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var resultCode = 0

var rootCmd = &cobra.Command{
	Use:   "calendar",
	Short: "Calendar application",
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		os.Exit(resultCode)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("config", "configs/config.yml", "Path to configuration file")
}
