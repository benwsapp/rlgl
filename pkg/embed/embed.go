package embed

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sync"

	_ "embed"

	"gopkg.in/yaml.v3"
)

var (
	//go:embed templates/index.html
	indexTemplateSource string

	compileOnce sync.Once
	compiled    *template.Template
	errCompile  error
)

type Contributor struct {
	Active bool     `json:"active" yaml:"active"`
	Focus  string   `json:"focus"  yaml:"focus"`
	Queue  []string `json:"queue"  yaml:"queue"`
}

type SlackConfig struct {
	Enabled             bool   `json:"enabled"               yaml:"enabled"`
	UserToken           string `json:"user_token"            yaml:"user_token"`            //nolint:tagliatelle
	StatusEmojiActive   string `json:"status_emoji_active"   yaml:"status_emoji_active"`   //nolint:tagliatelle
	StatusEmojiInactive string `json:"status_emoji_inactive" yaml:"status_emoji_inactive"` //nolint:tagliatelle
	TTLSeconds          int    `json:"ttl_seconds"           yaml:"ttl_seconds"`           //nolint:tagliatelle
}

type SiteConfig struct {
	Name        string      `json:"name"        yaml:"name"`
	Description string      `json:"description" yaml:"description"`
	User        string      `json:"user"        yaml:"user"`
	Contributor Contributor `json:"contributor" yaml:"contributor"`
	Slack       SlackConfig `json:"slack"       yaml:"slack"`
}

// GetTemplate returns the compiled index template, using sync.Once for caching.
func GetTemplate() (*template.Template, error) {
	compileOnce.Do(func() {
		compiled, errCompile = template.New("index").Parse(indexTemplateSource)
	})

	if errCompile != nil {
		return nil, fmt.Errorf("failed to compile template: %w", errCompile)
	}

	return compiled, nil
}

func LoadSiteConfig(path string) (SiteConfig, error) {
	// #nosec G304 - Path is controlled by caller and validated
	cleanPath := filepath.Clean(path)

	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return SiteConfig{}, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg SiteConfig

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return SiteConfig{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}

func Index(configPath string) ([]byte, error) {
	tmpl, err := GetTemplate()
	if err != nil {
		return nil, err
	}

	cfg, err := LoadSiteConfig(configPath)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer

	err = tmpl.Execute(&buf, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}
