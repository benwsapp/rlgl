package cmd

import (
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, _ []string) {
		cmd.Println("rlgl", Version)
		cmd.Println("Commit:", CommitSHA)
		cmd.Println("Built:", BuildTime)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
