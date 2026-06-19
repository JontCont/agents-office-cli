package workforce

import (
	"testing"
)

func TestDeterministicTurnCoordination(t *testing.T) {
	agents := []Agent{
		{Name: "planner", Role: "Planner"},
		{Name: "coder", Role: "Developer"},
		{Name: "reviewer", Role: "Reviewer"},
		{Name: "architect", Role: "Architect"},
	}

	tests := []struct {
		name            string
		lastMessage     string
		stage           string
		fallbackPlanner string
		expectedNext    string
	}{
		{
			name:         "Explicit handoff with @agent syntax",
			lastMessage:  "@agent coder, please review this",
			stage:        "Coding",
			expectedNext: "coder",
		},
		{
			name:         "Explicit handoff with @name syntax",
			lastMessage:  "Let me hand off to @reviewer",
			stage:        "Coding",
			expectedNext: "reviewer",
		},
		{
			name:         "Stage rule: Planning stage",
			lastMessage:  "Ready for next step",
			stage:        "Planning",
			expectedNext: "planner",
		},
		{
			name:         "Stage rule: Design stage",
			lastMessage:  "Starting design phase",
			stage:        "Design",
			expectedNext: "architect",
		},
		{
			name:            "Fallback to default planner agent",
			lastMessage:     "Hello world with no mentions or matching stages",
			stage:           "UnknownStage",
			fallbackPlanner: "planner",
			expectedNext:    "planner",
		},
		{
			name:         "Human handoff with @user syntax",
			lastMessage:  "Wait, let's ask @user",
			stage:        "Coding",
			expectedNext: "User",
		},
		{
			name:         "Human handoff with @supervisor syntax",
			lastMessage:  "@supervisor, what do you think?",
			stage:        "Coding",
			expectedNext: "User",
		},
		{
			name:         "Human handoff with @human syntax",
			lastMessage:  "Paging @human",
			stage:        "Coding",
			expectedNext: "User",
		},
		{
			name:         "Agent mention takes priority over human handoff",
			lastMessage:  "@reviewer, please review this. @user, what do you think?",
			stage:        "Coding",
			expectedNext: "reviewer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RouteTurn(tt.lastMessage, tt.stage, agents, tt.fallbackPlanner)
			if got != tt.expectedNext {
				t.Errorf("RouteTurn() = %v, want %v (test: %s)", got, tt.expectedNext, tt.name)
			}
		})
	}
}
