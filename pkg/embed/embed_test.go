package embed_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/benwsapp/rlgl/pkg/embed"
)

func TestLoadSiteConfig(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "site.yaml")

	configContent := `name: "Test Site"
description: "Test Description"
user: "testuser"
contributor:
  active: true
  focus: "testing feature"
  queue:
    - "task 1"
    - "task 2"
    - "task 3"
`

	err := os.WriteFile(configPath, []byte(configContent), 0o600)
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	cfg, err2 := embed.LoadSiteConfig(configPath)
	if err2 != nil {
		t.Fatalf("expected no error, got %v", err2)
	}

	if cfg.Name != "Test Site" {
		t.Errorf("expected name 'Test Site', got %s", cfg.Name)
	}

	if cfg.Description != "Test Description" {
		t.Errorf("expected description 'Test Description', got %s", cfg.Description)
	}

	if cfg.User != "testuser" {
		t.Errorf("expected user 'testuser', got %s", cfg.User)
	}

	if !cfg.Contributor.Active {
		t.Error("expected contributor.active to be true")
	}

	if cfg.Contributor.Focus != "testing feature" {
		t.Errorf("expected focus 'testing feature', got %s", cfg.Contributor.Focus)
	}

	if len(cfg.Contributor.Queue) != 3 {
		t.Errorf("expected 3 queue items, got %d", len(cfg.Contributor.Queue))
	}

	if cfg.Contributor.Queue[0] != "task 1" {
		t.Errorf("expected first queue item 'task 1', got %s", cfg.Contributor.Queue[0])
	}
}

func TestLoadSiteConfigFileNotFound(t *testing.T) {
	t.Parallel()

	_, err := embed.LoadSiteConfig("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestLoadSiteConfigInvalidYAML(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")

	invalidContent := `name: "Test
this is not valid yaml
	- broken
`

	err := os.WriteFile(configPath, []byte(invalidContent), 0o600)
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	_, err = embed.LoadSiteConfig(configPath)
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}

func TestLoadSiteConfigEmptyFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	configPath := filepath.Join(tmpDir, "empty.yaml")

	err := os.WriteFile(configPath, []byte(""), 0o600)
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	cfg, err := embed.LoadSiteConfig(configPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Name != "" {
		t.Errorf("expected empty name, got %s", cfg.Name)
	}
}

func TestIndex(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "site.yaml")

	configContent := `name: "Test Site"
description: "Test Description"
user: "testuser"
contributor:
  active: true
  focus: "testing"
  queue:
    - "task 1"
`

	err := os.WriteFile(configPath, []byte(configContent), 0o600)
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	content, err := embed.Index(configPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(content) == 0 {
		t.Error("expected non-empty content, got empty")
	}
}

func TestIndexInvalidConfig(t *testing.T) {
	t.Parallel()

	_, err := embed.Index("/nonexistent/config.yaml")
	if err == nil {
		t.Error("expected error for nonexistent config, got nil")
	}
}

func TestGetTemplate(t *testing.T) {
	t.Parallel()

	tmpl1, err := embed.GetTemplate()
	if err != nil {
		t.Fatalf("expected no error on first call, got %v", err)
	}

	if tmpl1 == nil {
		t.Fatal("expected non-nil template")
	}

	tmpl2, err := embed.GetTemplate()
	if err != nil {
		t.Fatalf("expected no error on second call, got %v", err)
	}

	if tmpl2 == nil {
		t.Fatal("expected non-nil template on second call")
	}

	if tmpl1 != tmpl2 {
		t.Error("expected same template instance from cache")
	}
}

func TestSiteConfigStruct(t *testing.T) {
	t.Parallel()

	cfg := embed.SiteConfig{
		Name:        "Test",
		Description: "Description",
		User:        "user",
		Contributor: embed.Contributor{
			Active: true,
			Focus:  "focus",
			Queue:  []string{"task1", "task2"},
		},
	}

	if cfg.Name != "Test" {
		t.Errorf("expected name 'Test', got %s", cfg.Name)
	}

	if cfg.Description != "Description" {
		t.Errorf("expected description 'Description', got %s", cfg.Description)
	}

	if cfg.User != "user" {
		t.Errorf("expected user 'user', got %s", cfg.User)
	}

	if !cfg.Contributor.Active {
		t.Error("expected active to be true")
	}

	if len(cfg.Contributor.Queue) != 2 {
		t.Errorf("expected 2 queue items, got %d", len(cfg.Contributor.Queue))
	}
}

func TestContributorStruct(t *testing.T) {
	t.Parallel()

	contrib := embed.Contributor{
		Active: false,
		Focus:  "test focus",
		Queue:  []string{"a", "b", "c"},
	}

	if contrib.Active {
		t.Error("expected active to be false")
	}

	if contrib.Focus != "test focus" {
		t.Errorf("expected focus 'test focus', got %s", contrib.Focus)
	}

	if len(contrib.Queue) != 3 {
		t.Errorf("expected 3 items, got %d", len(contrib.Queue))
	}
}
