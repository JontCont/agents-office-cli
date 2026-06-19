package config

import (
	"encoding/json"
	"os"
	"testing"
)

func TestAgentConfigJSONTags(t *testing.T) {
	agent := AgentConfig{
		Name:      "test-agent",
		Role:      "Tester",
		Backstory: "Tests stuff",
	}

	data, err := json.Marshal(agent)
	if err != nil {
		t.Fatalf("Failed to marshal AgentConfig: %v", err)
	}

	// Verify that the JSON contains lowercase keys
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal AgentConfig JSON: %v", err)
	}

	if _, ok := result["name"]; !ok {
		t.Errorf("Expected key 'name' in JSON, got: %s", string(data))
	}
	if _, ok := result["role"]; !ok {
		t.Errorf("Expected key 'role' in JSON, got: %s", string(data))
	}
	if _, ok := result["backstory"]; !ok {
		t.Errorf("Expected key 'backstory' in JSON, got: %s", string(data))
	}
}

func TestConfigLoadSave(t *testing.T) {
	tmpFile := "test-agent-office.yaml"
	defer func() {
		os.Remove(tmpFile)
		os.RemoveAll(".agents/agent-1")
	}()

	cfg := &Config{
		Version:  "1.0",
		Provider: "openai",
		Model:    "gpt-4o",
		Agents: []AgentConfig{
			{
				Name:      "agent-1",
				Role:      "Role 1",
				Backstory: "Backstory 1",
				Skills:    []string{"planning", "coding"},
				Tools:     []string{"search_web"},
				Hooks:     []string{"pre-commit"},
			},
		},
	}

	if err := SaveConfig(tmpFile, cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	loaded, err := LoadConfig(tmpFile)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if loaded.Version != cfg.Version || loaded.Provider != cfg.Provider || loaded.Model != cfg.Model {
		t.Errorf("Loaded config mismatch. Expected %+v, got %+v", cfg, loaded)
	}

	if len(loaded.Agents) != 1 || loaded.Agents[0].Name != "agent-1" {
		t.Errorf("Loaded agents mismatch. Expected %+v, got %+v", cfg.Agents, loaded.Agents)
	}

	agent := loaded.Agents[0]
	if len(agent.Skills) != 2 || agent.Skills[0] != "planning" || agent.Skills[1] != "coding" {
		t.Errorf("Expected skills ['planning', 'coding'], got %v", agent.Skills)
	}
	if len(agent.Tools) != 1 || agent.Tools[0] != "search_web" {
		t.Errorf("Expected tools ['search_web'], got %v", agent.Tools)
	}
	if len(agent.Hooks) != 1 || agent.Hooks[0] != "pre-commit" {
		t.Errorf("Expected hooks ['pre-commit'], got %v", agent.Hooks)
	}
}
