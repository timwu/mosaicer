package cmd

import "github.com/spf13/cobra"

var (
	rootCmd = &cobra.Command{
		Use:   "mosaicer",
		Short: "Generate photo mosaics",
	}
)

func Execute() error {
	return rootCmd.Execute()
}
