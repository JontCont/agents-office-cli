package workforce

import (
	"encoding/json"
	"os"
	"testing"
)

func TestProviderTokens(t *testing.T) {
	// Clean up any test configuration
	defer func() {
		os.Remove(TokenFileName)
	}()

	// Test 1: Load empty configuration
	os.Remove(TokenFileName)
	cfg, err := LoadTokenConfig()
	if err != nil {
		t.Fatalf("Failed to load empty token config: %v", err)
	}
	if len(cfg.Tokens) != 0 {
		t.Errorf("Expected 0 tokens, got %d", len(cfg.Tokens))
	}

	// Test 2: Save and Load token
	err = SaveToken(ProviderOpenRouter, "SK_OR_TEST")
	if err != nil {
		t.Fatalf("Failed to save token: %v", err)
	}

	token, ok := GetToken(ProviderOpenRouter)
	if !ok || token != "SK_OR_TEST" {
		t.Errorf("Expected token 'SK_OR_TEST', got '%s' (found: %v)", token, ok)
	}

	// Test 3: Default provider updated
	cfg, err = LoadTokenConfig()
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}
	if cfg.DefaultProvider != string(ProviderOpenRouter) {
		t.Errorf("Expected default provider to be 'openrouter', got '%s'", cfg.DefaultProvider)
	}
}

func TestDefaultModels(t *testing.T) {
	tests := []struct {
		provider      AIProvider
		expectedModel string
	}{
		{ProviderGemini, "gemini-1.5-pro"},
		{ProviderOpenAI, "gpt-4o"},
		{ProviderAnthropic, "claude-3-5-sonnet"},
		{ProviderOpenRouter, "google/gemini-flash-1.5"},
	}

	for _, tt := range tests {
		got := GetDefaultModel(tt.provider)
		if got != tt.expectedModel {
			t.Errorf("GetDefaultModel(%s) = %s, want %s", tt.provider, got, tt.expectedModel)
		}
	}
}

func TestGetStaticModels(t *testing.T) {
	// Task 2.2: Verify correct model IDs for each static provider
	tests := []struct {
		provider AIProvider
		contains string
	}{
		{ProviderOpenAI, "gpt-4o"},
		{ProviderGemini, "gemini-1.5-pro"},
		{ProviderAnthropic, "claude-3-5-sonnet"},
	}

	for _, tt := range tests {
		models := GetStaticModels(tt.provider)
		if len(models) == 0 {
			t.Errorf("GetStaticModels(%s) returned empty slice", tt.provider)
			continue
		}
		found := false
		for _, m := range models {
			if m == tt.contains {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GetStaticModels(%s) expected to contain %q, got %v", tt.provider, tt.contains, models)
		}
	}

	// OpenRouter should return empty (it's dynamic)
	openRouterModels := GetStaticModels(ProviderOpenRouter)
	if len(openRouterModels) != 0 {
		t.Errorf("GetStaticModels(openrouter) expected empty slice, got %v", openRouterModels)
	}

	// Unknown provider
	unknownModels := GetStaticModels("unknown")
	if len(unknownModels) != 0 {
		t.Errorf("GetStaticModels(unknown) expected empty slice, got %v", unknownModels)
	}
}

func TestGetCostPer1KTokens(t *testing.T) {
	// Task 2.3: Verify cost rates from spec example table
	tests := []struct {
		provider AIProvider
		expected float64
	}{
		{ProviderOpenAI, 0.005},
		{ProviderGemini, 0.0005},
		{ProviderAnthropic, 0.008},
		{ProviderOpenRouter, 0.002},
	}

	for _, tt := range tests {
		got := GetCostPer1KTokens(tt.provider)
		if got != tt.expected {
			t.Errorf("GetCostPer1KTokens(%s) = %f, want %f", tt.provider, got, tt.expected)
		}
	}

	// Unknown provider returns 0.0
	got := GetCostPer1KTokens("unknown")
	if got != 0.0 {
		t.Errorf("GetCostPer1KTokens(unknown) = %f, want 0.0", got)
	}
}

func TestWriteSessionLog(t *testing.T) {
	// Task 5.1: WriteSessionLog writes correct JSON to .agent-office-sessions/<date>-<runID>.json
	testDir := ".test-agent-office-sessions"
	origDir := SessionLogDir

	// Patch the constant for testing by temporarily using a helper
	// Since SessionLogDir is a const, we write to a temp location by overriding os.MkdirAll target
	// Strategy: use a sub-test that writes to the real dir and cleans up
	t.Run("writes valid JSON", func(t *testing.T) {
		_ = testDir
		_ = origDir
		defer os.RemoveAll(SessionLogDir)

		log := SessionLog{
			RunID:            "run-99999",
			Provider:         "openrouter",
			Model:            "google/gemini-flash-1.5",
			StartedAt:        1718000000,
			EndedAt:          1718000060,
			TotalTokens:      695,
			EstimatedCostUSD: 0.00139, // 695/1000 * 0.002
			Steps:            4,
		}

		if err := WriteSessionLog(log); err != nil {
			t.Fatalf("WriteSessionLog failed: %v", err)
		}

		// Find the written file
		entries, err := os.ReadDir(SessionLogDir)
		if err != nil || len(entries) == 0 {
			t.Fatal("Expected session log file to be created")
		}

		data, err := os.ReadFile(SessionLogDir + "/" + entries[0].Name())
		if err != nil {
			t.Fatalf("Failed to read session log: %v", err)
		}

		var readBack SessionLog
		if err := json.Unmarshal(data, &readBack); err != nil {
			t.Fatalf("Failed to parse session log JSON: %v", err)
		}

		if readBack.RunID != log.RunID {
			t.Errorf("run_id mismatch: got %s, want %s", readBack.RunID, log.RunID)
		}
		if readBack.Provider != log.Provider {
			t.Errorf("provider mismatch: got %s, want %s", readBack.Provider, log.Provider)
		}
		if readBack.TotalTokens != log.TotalTokens {
			t.Errorf("total_tokens mismatch: got %d, want %d", readBack.TotalTokens, log.TotalTokens)
		}
		if readBack.Steps != log.Steps {
			t.Errorf("steps mismatch: got %d, want %d", readBack.Steps, log.Steps)
		}
	})
}
