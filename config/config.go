package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration
type Config struct {
	App      AppConfig      `yaml:"app"`
	Database DatabaseConfig `yaml:"database"`
	LLM      LLMConfig     `yaml:"llm"`
	SRS      SRSConfig     `yaml:"srs"`
}

type AppConfig struct {
	Name string `yaml:"name"`
	Port int    `yaml:"port"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type LLMConfig struct {
	Provider string `yaml:"provider"` // openai or deepseek
	APIKey   string `yaml:"api_key"`
	Model    string `yaml:"model"`
}

type SRSConfig struct {
	InitialEF float64 `yaml:"initial_ef"`
	MinEF     float64 `yaml:"min_ef"`
}

// Load loads configuration from a YAML file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Load API key from environment variable if not set
	if cfg.LLM.APIKey == "" {
		cfg.LLM.APIKey = os.Getenv("DEEPSEEK_API_KEY")
		if cfg.LLM.APIKey == "" {
			cfg.LLM.APIKey = os.Getenv("OPENAI_API_KEY")
		}
	}

	// Set defaults
	if cfg.App.Port == 0 {
		cfg.App.Port = 8081
	}
	if cfg.SRS.InitialEF == 0 {
		cfg.SRS.InitialEF = 2.5
	}
	if cfg.SRS.MinEF == 0 {
		cfg.SRS.MinEF = 1.3
	}

	return &cfg, nil
}
