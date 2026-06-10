package config

import (
	"os"
	"gopkg.in/yaml.v3"
)

type AgentConfig struct {
	Name      string   `yaml:"name" json:"name"`
	Role      string   `yaml:"role" json:"role"`
	Backstory string   `yaml:"backstory" json:"backstory"`
	Skills    []string `yaml:"skills" json:"skills"`
	Tools     []string `yaml:"tools" json:"tools"`
	Hooks     []string `yaml:"hooks" json:"hooks"`
	Color     string   `yaml:"color" json:"color"`
	Avatar    string   `yaml:"avatar" json:"avatar"`
	Provider  string   `yaml:"provider" json:"provider"`
	Model     string   `yaml:"model" json:"model"`
	Token     string   `yaml:"token,omitempty" json:"token,omitempty"`
}

type Config struct {
	Version  string        `yaml:"version"`
	Provider string        `yaml:"provider"`
	Model    string        `yaml:"model"`
	Agents   []AgentConfig `yaml:"agents"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// SaveConfig marshals cfg to YAML and writes it to path with 0644 permissions.
func SaveConfig(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
