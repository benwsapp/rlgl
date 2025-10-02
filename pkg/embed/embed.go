package embed

import (
	"bytes"
	"html/template"
	"os"
	"sync"

	_ "embed"

	"gopkg.in/yaml.v3"
)

var (
	//go:embed templates/index.html
	indexTemplateSource string

	compileOnce sync.Once
	compiled    *template.Template
	compileErr  error
)

type Contributor struct {
	Active bool     `yaml:"active" json:"active"`
	Focus  string   `yaml:"focus" json:"focus"`
	Queue  []string `yaml:"queue" json:"queue"`
}

type SiteConfig struct {
	Name        string      `yaml:"name" json:"name"`
	Description string      `yaml:"description" json:"description"`
	User        string      `yaml:"user" json:"user"`
	Contributor Contributor `yaml:"contributor" json:"contributor"`
}

func getTemplate() (*template.Template, error) {
	compileOnce.Do(func() {
		compiled, compileErr = template.New("index").Parse(indexTemplateSource)
	})

	return compiled, compileErr
}

func LoadSiteConfig(path string) (SiteConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return SiteConfig{}, err
	}

	var cfg SiteConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return SiteConfig{}, err
	}

	return cfg, nil
}

func Index(configPath string) ([]byte, error) {
	tmpl, err := getTemplate()
	if err != nil {
		return nil, err
	}

	cfg, err := LoadSiteConfig(configPath)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, cfg); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
