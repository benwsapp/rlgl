package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/benwsapp/rlgl/cmd"
	"github.com/spf13/viper"
)

func TestRunCommand(t *testing.T) {
	t.Parallel()

	runCmd := cmd.RootCmd.Commands()[0]
	if runCmd == nil {
		t.Fatal("runCmd should not be nil")
	}

	if runCmd.Use != "run" {
		t.Errorf("expected Use 'run', got %s", runCmd.Use)
	}

	if runCmd.Short == "" {
		t.Error("expected non-empty Short description")
	}

	if runCmd.RunE == nil {
		t.Error("expected RunE to be defined")
	}
}

func TestRunCommandFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		defaultValue string
	}{
		{"addr", ":8080"},
		{"config", ""},
	}

	runCmd := cmd.RootCmd.Commands()[0]

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			flag := runCmd.Flags().Lookup(testCase.name)
			if flag == nil {
				t.Fatalf("expected '%s' flag to exist", testCase.name)
			}

			if flag.DefValue != testCase.defaultValue {
				t.Errorf("expected default value '%s', got %s", testCase.defaultValue, flag.DefValue)
			}
		})
	}
}

func TestRunCommandTrustedOriginsFlag(t *testing.T) {
	t.Parallel()

	runCmd := cmd.RootCmd.Commands()[0]

	flag := runCmd.Flags().Lookup("trusted-origins")
	if flag == nil {
		t.Fatal("expected 'trusted-origins' flag to exist")
	}

	if flag.Value.Type() != "stringSlice" {
		t.Errorf("expected type 'stringSlice', got %s", flag.Value.Type())
	}
}

//nolint:paralleltest // Cannot run in parallel - uses viper.Reset()
func TestInitConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "site.yaml")
	configContent := `name: "Test Site"
description: "Test"
user: "test"
contributor:
  active: true
  focus: "test"
  queue: []
`

	err := os.WriteFile(configPath, []byte(configContent), 0o600)
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	viper.Reset()

	runCmd := cmd.RootCmd.Commands()[0]

	err = runCmd.Flags().Set("config", configPath)
	if err != nil {
		t.Fatalf("failed to set config flag: %v", err)
	}

	result, err := cmd.InitConfig(runCmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != configPath {
		t.Errorf("expected config path %s, got %s", configPath, result)
	}
}

//nolint:paralleltest // Cannot run in parallel - uses viper.Reset()
func TestInitConfigNotFound(t *testing.T) {
	viper.Reset()

	runCmd := cmd.RootCmd.Commands()[0]

	err := runCmd.Flags().Set("config", "/nonexistent/config.yaml")
	if err != nil {
		t.Fatalf("failed to set config flag: %v", err)
	}

	_, err = cmd.InitConfig(runCmd)
	if err == nil {
		t.Error("expected error for nonexistent config, got nil")
	}
}

//nolint:paralleltest // Cannot run in parallel - uses viper.Reset() and t.Chdir()
func TestInitConfigDefaultPath(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "site.yaml")
	configContent := `name: "Test"
description: "Test"
user: "test"
contributor:
  active: true
  focus: "test"
  queue: []
`

	err := os.WriteFile(configPath, []byte(configContent), 0o600)
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	viper.Reset()
	t.Chdir(tmpDir)

	runCmd := cmd.RootCmd.Commands()[0]

	err = runCmd.Flags().Set("config", "")
	if err != nil {
		t.Fatalf("failed to set config flag: %v", err)
	}

	result, err := cmd.InitConfig(runCmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result == "" {
		t.Error("expected non-empty config path")
	}
}

func TestViperBindings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		envVar string
		key    string
	}{
		{"RLGL_ADDR", "addr"},
		{"RLGL_CONFIG", "config"},
		{"RLGL_TRUSTED_ORIGINS", "trusted-origins"},
	}

	runCmd := cmd.RootCmd.Commands()[0]

	for _, testCase := range tests {
		t.Run(testCase.envVar, func(t *testing.T) {
			t.Parallel()

			flag := runCmd.Flags().Lookup(testCase.key)
			if flag == nil {
				t.Errorf("expected flag '%s' to exist for env var %s", testCase.key, testCase.envVar)
			}
		})
	}
}
