package workforce

import (
	"strings"
)

// RouteTurn resolves the next agent name based on the last message, active stage, configured agents, and fallback planner.
func RouteTurn(lastMessage string, stage string, agents []Agent, fallbackPlanner string) string {
	normalize := func(s string) string {
		s = strings.ToLower(s)
		s = strings.ReplaceAll(s, " ", "")
		s = strings.ReplaceAll(s, "-", "")
		s = strings.ReplaceAll(s, "_", "")
		return s
	}

	// 1. Explicit Handoff Check
	lowerMsg := strings.ToLower(lastMessage)
	for _, a := range agents {
		name := strings.ToLower(a.Name)
		norm := normalize(a.Name)
		hyphenated := strings.ReplaceAll(name, " ", "-")
		underscored := strings.ReplaceAll(name, " ", "_")

		if strings.Contains(lowerMsg, "@agent "+name) ||
			strings.Contains(lowerMsg, "@agent "+norm) ||
			strings.Contains(lowerMsg, "@agent "+hyphenated) ||
			strings.Contains(lowerMsg, "@agent "+underscored) ||
			strings.Contains(lowerMsg, "@"+name) ||
			strings.Contains(lowerMsg, "@"+norm) ||
			strings.Contains(lowerMsg, "@"+hyphenated) ||
			strings.Contains(lowerMsg, "@"+underscored) {
			return a.Name
		}
	}

	// Build exact map for stage/fallback lookups
	agentMapExact := make(map[string]bool)
	for _, a := range agents {
		agentMapExact[strings.ToLower(a.Name)] = true
	}

	// 2. Stage rules (role-based routing)
	switch strings.ToLower(stage) {
	case "planning":
		if agentMapExact["planner"] {
			return "planner"
		}
	case "design":
		if agentMapExact["architect"] {
			return "architect"
		}
	case "coding":
		if agentMapExact["reviewer"] {
			return "reviewer"
		}
	}

	// 3. Fallback Planner
	if fallbackPlanner != "" && agentMapExact[strings.ToLower(fallbackPlanner)] {
		return fallbackPlanner
	}

	// Default fallback
	if len(agents) > 0 {
		return agents[0].Name
	}

	return ""
}
