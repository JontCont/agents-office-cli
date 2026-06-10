package workforce

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type AIProvider string

const (
	ProviderGemini     AIProvider = "gemini"
	ProviderOpenAI     AIProvider = "openai"
	ProviderAnthropic  AIProvider = "anthropic"
	ProviderOpenRouter AIProvider = "openrouter"
)

type TokenConfig struct {
	DefaultProvider string            `json:"default_provider"`
	Tokens          map[string]string `json:"tokens"`
}

const TokenFileName = ".\\environment\\.agent-office-token"

// GetTokenConfigPath returns the path of the token file in the workspace root.
func GetTokenConfigPath() string {
	return TokenFileName
}

// LoadTokenConfig loads the token config from disk.
func LoadTokenConfig() (*TokenConfig, error) {
	path := GetTokenConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &TokenConfig{
				Tokens: make(map[string]string),
			}, nil
		}
		return nil, err
	}

	var cfg TokenConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Tokens == nil {
		cfg.Tokens = make(map[string]string)
	}
	return &cfg, nil
}

// SaveTokenConfig saves the token config to disk.
func SaveTokenConfig(cfg *TokenConfig) error {
	path := GetTokenConfigPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// SaveToken sets the token for a specific provider and saves it.
func SaveToken(provider AIProvider, token string) error {
	cfg, err := LoadTokenConfig()
	if err != nil {
		return err
	}
	cfg.Tokens[string(provider)] = token
	cfg.DefaultProvider = string(provider)
	return SaveTokenConfig(cfg)
}

// GetToken returns the token for a specific provider.
func GetToken(provider AIProvider) (string, bool) {
	cfg, err := LoadTokenConfig()
	if err != nil {
		return "", false
	}
	token, ok := cfg.Tokens[string(provider)]
	return token, ok
}

// GetDefaultModel returns the default model name for the given provider.
func GetDefaultModel(provider AIProvider) string {
	switch provider {
	case ProviderGemini:
		return "gemini-1.5-pro"
	case ProviderOpenAI:
		return "gpt-4o"
	case ProviderAnthropic:
		return "claude-3-5-sonnet"
	case ProviderOpenRouter:
		return "google/gemini-flash-1.5"
	default:
		return ""
	}
}

// ProviderStaticModels holds a curated list of models for each static provider.
var ProviderStaticModels = map[AIProvider][]string{
	ProviderOpenAI: {
		"gpt-4o",
		"gpt-4-turbo",
		"gpt-4",
		"gpt-3.5-turbo",
	},
	ProviderGemini: {
		"gemini-1.5-pro",
		"gemini-1.5-flash",
		"gemini-1.0-pro",
	},
	ProviderAnthropic: {
		"claude-3-5-sonnet",
		"claude-3-opus",
		"claude-3-haiku",
	},
}

// GetStaticModels returns the curated model list for a static provider.
// Returns an empty slice for unknown providers or OpenRouter (which is dynamic).
func GetStaticModels(provider AIProvider) []string {
	models, ok := ProviderStaticModels[provider]
	if !ok {
		return []string{}
	}
	return models
}

// ProviderCostPer1KTokens holds blended input+output cost estimates per 1K tokens (USD).
var ProviderCostPer1KTokens = map[AIProvider]float64{
	ProviderOpenAI:     0.005,
	ProviderGemini:     0.0005,
	ProviderAnthropic:  0.008,
	ProviderOpenRouter: 0.002,
}

// GetCostPer1KTokens returns the cost per 1K tokens for the given provider.
// Returns 0.0 for unknown providers.
func GetCostPer1KTokens(provider AIProvider) float64 {
	rate, ok := ProviderCostPer1KTokens[provider]
	if !ok {
		return 0.0
	}
	return rate
}

// SessionLog represents a completed run's usage and cost summary.
type SessionLog struct {
	RunID            string  `json:"run_id"`
	Provider         string  `json:"provider"`
	Model            string  `json:"model"`
	StartedAt        int64   `json:"started_at"`
	EndedAt          int64   `json:"ended_at"`
	TotalTokens      int     `json:"total_tokens"`
	EstimatedCostUSD float64 `json:"estimated_cost_usd"`
	Steps            int     `json:"steps"`
}

const SessionLogDir = "sessions"

// WriteSessionLog creates the sessions directory if needed and writes the log
// as <YYYY-MM-DD>-<runID>.json using the EndedAt Unix timestamp for the date.
func WriteSessionLog(log SessionLog) error {
	if err := os.MkdirAll(SessionLogDir, 0755); err != nil {
		return err
	}
	t := time.Unix(log.EndedAt, 0)
	filename := t.Format("2006-01-02") + "-" + log.RunID + ".json"
	path := SessionLogDir + "/" + filename

	data, err := json.MarshalIndent(log, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
