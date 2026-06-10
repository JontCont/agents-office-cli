package workforce

import (
	"encoding/json"
	"testing"
)

func TestEventSerialization(t *testing.T) {
	evt := Event{
		ID:        "evt-123",
		RunID:     "run-456",
		Type:      "agent.speak",
		Timestamp: 1625097600,
		Sender:    "planner",
		Content:   "Hello from planner",
		Metadata: map[string]interface{}{
			"tokens": 42,
		},
	}

	data, err := json.Marshal(evt)
	if err != nil {
		t.Fatalf("Failed to marshal Event: %v", err)
	}

	var parsed Event
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		t.Fatalf("Failed to unmarshal Event: %v", err)
	}

	if parsed.ID != evt.ID || parsed.Sender != evt.Sender || parsed.Content != evt.Content {
		t.Errorf("Mismatch in serialized/deserialized event fields. Got %+v, want %+v", parsed, evt)
	}

	if tokens, ok := parsed.Metadata["tokens"].(float64); !ok || int(tokens) != 42 {
		t.Errorf("Expected metadata 'tokens' to be 42, got %v", parsed.Metadata["tokens"])
	}
}

func TestCommandSerialization(t *testing.T) {
	cmd := Command{
		Type:    "run.resume",
		RunID:   "run-456",
		Message: "Proceed",
	}

	data, err := json.Marshal(cmd)
	if err != nil {
		t.Fatalf("Failed to marshal Command: %v", err)
	}

	var parsed Command
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		t.Fatalf("Failed to unmarshal Command: %v", err)
	}

	if parsed.Type != cmd.Type || parsed.RunID != cmd.RunID || parsed.Message != cmd.Message {
		t.Errorf("Mismatch in serialized/deserialized command fields. Got %+v, want %+v", parsed, cmd)
	}
}
