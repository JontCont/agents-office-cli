package config

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type AgentConfig struct {
	Name      string   `yaml:"name" json:"name"`
	Role      string   `yaml:"role" json:"role"`
	Backstory string   `yaml:"backstory" json:"backstory"`
	Skills    []string `yaml:"skills" json:"skills"`
	Tools     []string `yaml:"tools" json:"tools"`
	Hooks     []string `yaml:"hooks" json:"hooks"`
	Color           string   `yaml:"color" json:"color"`
	Avatar          string   `yaml:"avatar" json:"avatar"`
	VisualCharacter string   `yaml:"visual_character,omitempty" json:"visual_character,omitempty"`
	Provider        string   `yaml:"provider" json:"provider"`
	Model           string   `yaml:"model" json:"model"`
	Token           string   `yaml:"token,omitempty" json:"token,omitempty"`
}

type Config struct {
	Version  string        `yaml:"version"`
	Username string        `yaml:"username,omitempty"`
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

	// Load backstory from filesystem if exists
	for i, a := range cfg.Agents {
		normalized := strings.ToLower(a.Name)
		normalized = strings.ReplaceAll(normalized, " ", "_")
		promptPath := filepath.Join(".agents", normalized, "system_prompt.md")
		if content, err := os.ReadFile(promptPath); err == nil {
			cfg.Agents[i].Backstory = strings.TrimSpace(string(content))
		}
	}

	return &cfg, nil
}

// SaveConfig marshals cfg to YAML and writes it to path with 0644 permissions.
func SaveConfig(path string, cfg *Config) error {
	// Preserve original backstories to restore them afterwards
	backstories := make([]string, len(cfg.Agents))
	for i, a := range cfg.Agents {
		backstories[i] = a.Backstory
		if a.Backstory != "" {
			normalized := strings.ToLower(a.Name)
			normalized = strings.ReplaceAll(normalized, " ", "_")
			dirPath := filepath.Join(".agents", normalized)
			if err := os.MkdirAll(dirPath, 0755); err != nil {
				return err
			}
			filePath := filepath.Join(dirPath, "system_prompt.md")
			if err := os.WriteFile(filePath, []byte(a.Backstory), 0644); err != nil {
				return err
			}
		}
		cfg.Agents[i].Backstory = ""
	}

	defer func() {
		for i := range cfg.Agents {
			cfg.Agents[i].Backstory = backstories[i]
		}
	}()

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
