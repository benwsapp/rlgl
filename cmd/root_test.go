package cmd_test

import (
	"testing"

	"github.com/benwsapp/rlgl/cmd"
)

func TestRootCommand(t *testing.T) {
	t.Parallel()

	if cmd.RootCmd == nil {
		t.Fatal("rootCmd should not be nil")
	}

	if cmd.RootCmd.Use != "rlgl" {
		t.Errorf("expected Use 'rlgl', got %s", cmd.RootCmd.Use)
	}

	if cmd.RootCmd.Short == "" {
		t.Error("expected non-empty Short description")
	}

	if cmd.RootCmd.Long == "" {
		t.Error("expected non-empty Long description")
	}
}

func TestRootCommandFlags(t *testing.T) {
	t.Parallel()

	flag := cmd.RootCmd.Flags().Lookup("toggle")
	if flag == nil {
		t.Fatal("expected 'toggle' flag to exist")
	}

	if flag.Shorthand != "t" {
		t.Errorf("expected shorthand 't', got %s", flag.Shorthand)
	}

	if flag.DefValue != "false" {
		t.Errorf("expected default value 'false', got %s", flag.DefValue)
	}
}

func TestExecute(t *testing.T) {
	t.Parallel()

	if cmd.RootCmd == nil {
		t.Error("rootCmd should not be nil")
	}
}
