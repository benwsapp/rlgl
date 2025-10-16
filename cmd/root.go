/*
Copyright © 2025 Ben Sapp ya.bsapp.ru
*/

package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// Version information, set during build via ldflags.
var (
	Version   = "dev"
	CommitSHA = "unknown"
	BuildTime = "unknown"
)

// RootCmd is the root command for rlgl.
var RootCmd = &cobra.Command{
	Use:   "rlgl",
	Short: "Current work-in-progress status dashboard",
	Long: `Status dashboard that shows what's currently
in progress as well as what upcoming items are in the user's
personal backlog.

rlgl reads from a local config file and syncs the status to a
remote web page.`,
}

func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
