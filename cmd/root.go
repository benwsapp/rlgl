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
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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
