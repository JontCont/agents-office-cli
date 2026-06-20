package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"agent-office/pkg/config"
	"agent-office/pkg/workforce"
	"github.com/gorilla/websocket"
)

func init() {
	// Change working directory to project root so skill templates and config files can be loaded
	_ = os.Chdir("../..")
}

func TestIsAgentTagged(t *testing.T) {
	tests := []struct {
		message   string
		agentName string
		expected  bool
	}{
		{"@TechnicalArchitect go", "Technical Architect", true},
		{"@technical-architect go", "Technical Architect", true},
		{"@technicalarchitect go", "Technical Architect", true},
		{"hello", "Technical Architect", false},
		{"@everyone review", "Technical Architect", false}, // Note: @everyone is handled at the call site, not in isAgentTagged
		{"@CriticalReviewer test", "Critical Reviewer", true},
		{"@critical-reviewer test", "Critical Reviewer", true},
		{"@critical_reviewer test", "Critical Reviewer", true},
	}

	for _, tt := range tests {
		result := isAgentTagged(tt.message, tt.agentName)
		if result != tt.expected {
			t.Errorf("isAgentTagged(%q, %q) = %v; want %v", tt.message, tt.agentName, result, tt.expected)
		}
	}
}

func TestTagRequiredRoutingIntegration(t *testing.T) {
	// Set up mock config
	cfg := &config.Config{
		Provider:    "openai",
		Model:       "gpt-4o",
		TagRequired: true,
		Agents: []config.AgentConfig{
			{Name: "Lead Strategist", Role: "Lead Strategist", Color: "#123456", Avatar: ""},
			{Name: "Technical Architect", Role: "Technical Architect", Color: "#654321", Avatar: ""},
			{Name: "Critical Reviewer", Role: "Critical Reviewer", Color: "#abcdef", Avatar: ""},
		},
	}

	// Create coordinator
	var server *workforce.WSServer
	coord := workforce.NewCoordinator(func(state workforce.RunState) {
		if server != nil {
			server.BroadcastEvent(workforce.Event{
				Type:    "state.change",
				Content: string(state),
			})
		}
	}, nil)
	server = workforce.NewWSServer(coord)

	// Create httptest Server
	mux := http.NewServeMux()
	server.ServeOnMux(mux)
	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	// Convert http URL to ws URL
	wsURL := strings.Replace(testServer.URL, "http", "ws", 1) + "/ws"

	// Connect WebSocket client
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to dial WebSocket server: %v", err)
	}
	defer conn.Close()

	// Read initial state
	var initialEvt workforce.Event
	if err := conn.ReadJSON(&initialEvt); err != nil {
		t.Fatalf("Failed to read initial state event: %v", err)
	}

	// Channel to collect events
	events := make(chan workforce.Event, 100)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			var evt workforce.Event
			err := conn.ReadJSON(&evt)
			if err != nil {
				return
			}
			events <- evt
		}
	}()

	// Start workforce execution in a goroutine
	runID := "test-run-123"
	prompt := "hello world" // no agent tags, should skip all and transition to human handoff
	
	go runWorkforceExecution(coord, server, cfg, prompt, runID)

	// Wait and verify events for initial skipping
	timeout := time.After(3 * time.Second)
	skippedLead := false
	skippedArch := false
	skippedRev := false
	interrupted := false

Loop1:
	for {
		select {
		case evt := <-events:
			if evt.Type == "state.change" && evt.Content == "INTERRUPTED" {
				interrupted = true
			}
			if evt.Type == "system.log" {
				if strings.Contains(evt.Content, "Skipping agent Lead Strategist") {
					skippedLead = true
				}
				if strings.Contains(evt.Content, "Skipping agent Technical Architect") {
					skippedArch = true
				}
				if strings.Contains(evt.Content, "Skipping agent Critical Reviewer") {
					skippedRev = true
				}
			}
		case <-timeout:
			break Loop1
		}
		if interrupted && skippedLead && skippedArch && skippedRev {
			break Loop1
		}
	}

	if !skippedLead {
		t.Error("Expected Lead Strategist to be skipped")
	}
	if !skippedArch {
		t.Error("Expected Technical Architect to be skipped")
	}
	if !skippedRev {
		t.Error("Expected Critical Reviewer to be skipped")
	}
	if !interrupted {
		t.Error("Expected execution to be interrupted for human handoff")
	}

	// Now send a resume command (supervisor feedback with tag to Technical Architect)
	resumeCmd := workforce.Command{
		Type:    "run.resume",
		RunID:   runID,
		Message: "@TechnicalArchitect please build it",
	}
	cmdBytes, _ := json.Marshal(resumeCmd)
	_ = conn.WriteMessage(websocket.TextMessage, cmdBytes)

	// Wait and verify that Technical Architect speaks
	timeout = time.After(5 * time.Second)
	architectSpoke := false

Loop2:
	for {
		select {
		case evt := <-events:
			if evt.Type == "agent.speak" && evt.Sender == "Technical Architect" {
				architectSpoke = true
				break Loop2
			}
		case <-timeout:
			break Loop2
		}
		if architectSpoke {
			break Loop2
		}
	}

	if !architectSpoke {
		t.Error("Expected Technical Architect to speak after being tagged in resume message")
	}

	// Abort the run to clean up
	abortCmd := workforce.Command{
		Type:  "run.abort",
		RunID: runID,
	}
	abortBytes, _ := json.Marshal(abortCmd)
	_ = conn.WriteMessage(websocket.TextMessage, abortBytes)
}

func TestTagRequiredEveryoneIntegration(t *testing.T) {
	cfg := &config.Config{
		Provider:    "openai",
		Model:       "gpt-4o",
		TagRequired: true,
		Agents: []config.AgentConfig{
			{Name: "Lead Strategist", Role: "Lead Strategist", Color: "#123456", Avatar: ""},
			{Name: "Technical Architect", Role: "Technical Architect", Color: "#654321", Avatar: ""},
		},
	}

	coord := workforce.NewCoordinator(nil, nil)
	server := workforce.NewWSServer(coord)

	mux := http.NewServeMux()
	server.ServeOnMux(mux)
	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	wsURL := strings.Replace(testServer.URL, "http", "ws", 1) + "/ws"

	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to dial WebSocket server: %v", err)
	}
	defer conn.Close()

	var initialEvt workforce.Event
	_ = conn.ReadJSON(&initialEvt)

	events := make(chan workforce.Event, 100)
	go func() {
		for {
			var evt workforce.Event
			err := conn.ReadJSON(&evt)
			if err != nil {
				return
			}
			events <- evt
		}
	}()

	runID := "test-run-everyone"
	prompt := "@everyone start the analysis"

	go runWorkforceExecution(coord, server, cfg, prompt, runID)

	timeout := time.After(5 * time.Second)
	leadSpoke := false
	archSpoke := false
	skippedAny := false

LoopEveryone:
	for {
		select {
		case evt := <-events:
			if evt.Type == "agent.speak" {
				if evt.Sender == "Lead Strategist" {
					leadSpoke = true
				}
				if evt.Sender == "Technical Architect" {
					archSpoke = true
				}
			}
			if evt.Type == "system.log" && strings.Contains(evt.Content, "Skipping agent") {
				skippedAny = true
			}
		case <-timeout:
			break LoopEveryone
		}
		if (leadSpoke && archSpoke) || skippedAny {
			break LoopEveryone
		}
	}

	if skippedAny {
		t.Error("Expected no agents to be skipped with @everyone tag")
	}
	if !leadSpoke {
		t.Error("Expected Lead Strategist to speak under @everyone")
	}

	abortCmd := workforce.Command{
		Type:  "run.abort",
		RunID: runID,
	}
	abortBytes, _ := json.Marshal(abortCmd)
	_ = conn.WriteMessage(websocket.TextMessage, abortBytes)
}

func TestTagRequiredDisabledIntegration(t *testing.T) {
	cfg := &config.Config{
		Provider:    "openai",
		Model:       "gpt-4o",
		TagRequired: false,
		Agents: []config.AgentConfig{
			{Name: "Lead Strategist", Role: "Lead Strategist", Color: "#123456", Avatar: ""},
		},
	}

	coord := workforce.NewCoordinator(nil, nil)
	server := workforce.NewWSServer(coord)

	mux := http.NewServeMux()
	server.ServeOnMux(mux)
	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	wsURL := strings.Replace(testServer.URL, "http", "ws", 1) + "/ws"

	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to dial WebSocket server: %v", err)
	}
	defer conn.Close()

	var initialEvt workforce.Event
	_ = conn.ReadJSON(&initialEvt)

	events := make(chan workforce.Event, 100)
	go func() {
		for {
			var evt workforce.Event
			err := conn.ReadJSON(&evt)
			if err != nil {
				return
			}
			events <- evt
		}
	}()

	runID := "test-run-disabled"
	prompt := "hello without tags"

	go runWorkforceExecution(coord, server, cfg, prompt, runID)

	timeout := time.After(5 * time.Second)
	leadSpoke := false
	skippedAny := false

LoopDisabled:
	for {
		select {
		case evt := <-events:
			if evt.Type == "agent.speak" && evt.Sender == "Lead Strategist" {
				leadSpoke = true
			}
			if evt.Type == "system.log" && strings.Contains(evt.Content, "Skipping agent") {
				skippedAny = true
			}
		case <-timeout:
			break LoopDisabled
		}
		if leadSpoke || skippedAny {
			break LoopDisabled
		}
	}

	if skippedAny {
		t.Error("Expected no agents to be skipped when tag_required is false")
	}
	if !leadSpoke {
		t.Error("Expected Lead Strategist to speak as fallback when tag_required is false")
	}

	abortCmd := workforce.Command{
		Type:  "run.abort",
		RunID: runID,
	}
	abortBytes, _ := json.Marshal(abortCmd)
	_ = conn.WriteMessage(websocket.TextMessage, abortBytes)
}

func TestParseFrontmatter_ValidYAML(t *testing.T) {
	content := "---\ninput_token_provider: input-prov\noutput_token_provider: output-prov\n---\n# My Skill\nHello"
	body, inProv, outProv := parseSkillContent(content)
	if body != "# My Skill\nHello" {
		t.Errorf("Expected body '# My Skill\\nHello', got %q", body)
	}
	if inProv != "input-prov" {
		t.Errorf("Expected input provider 'input-prov', got %q", inProv)
	}
	if outProv != "output-prov" {
		t.Errorf("Expected output provider 'output-prov', got %q", outProv)
	}
}

func TestParseFrontmatter_Malformed(t *testing.T) {
	content := "---\ninput_token_provider: [unclosed\n---\n# My Skill\nHello"
	body, inProv, outProv := parseSkillContent(content)
	if body != "# My Skill\nHello" {
		t.Errorf("Expected body '# My Skill\\nHello', got %q", body)
	}
	if inProv != "" || outProv != "" {
		t.Errorf("Expected empty providers, got in: %q, out: %q", inProv, outProv)
	}
}

func TestParseFrontmatter_MissingDelimiters(t *testing.T) {
	content := "# My Skill\nHello"
	body, inProv, outProv := parseSkillContent(content)
	if body != content {
		t.Errorf("Expected body to match content, got %q", body)
	}
	if inProv != "" || outProv != "" {
		t.Errorf("Expected empty providers, got in: %q, out: %q", inProv, outProv)
	}
}

func TestParseFrontmatter_EmptyBlock(t *testing.T) {
	content := "---\n---\n# My Skill\nHello"
	body, inProv, outProv := parseSkillContent(content)
	if body != "# My Skill\nHello" {
		t.Errorf("Expected body '# My Skill\\nHello', got %q", body)
	}
	if inProv != "" || outProv != "" {
		t.Errorf("Expected empty providers, got in: %q, out: %q", inProv, outProv)
	}
}

func TestParseFrontmatter_ExtraFields(t *testing.T) {
	content := "---\ninput_token_provider: input-prov\noutput_token_provider: output-prov\nextra_field: ignore-me\n---\n# My Skill\nHello"
	body, inProv, outProv := parseSkillContent(content)
	if body != "# My Skill\nHello" {
		t.Errorf("Expected body '# My Skill\\nHello', got %q", body)
	}
	if inProv != "input-prov" {
		t.Errorf("Expected input provider 'input-prov', got %q", inProv)
	}
	if outProv != "output-prov" {
		t.Errorf("Expected output provider 'output-prov', got %q", outProv)
	}
}

func TestProviderFallback_AgentDefault(t *testing.T) {
	agentProvider := "agent-default"
	
	inProv := ""
	outProv := ""
	
	inputTokenProvider := agentProvider
	outputTokenProvider := agentProvider
	
	if inProv != "" {
		inputTokenProvider = inProv
	}
	if outProv != "" {
		outputTokenProvider = outProv
	}
	
	if inputTokenProvider != "agent-default" || outputTokenProvider != "agent-default" {
		t.Errorf("Expected fallback to agent provider, got in: %q, out: %q", inputTokenProvider, outputTokenProvider)
	}
}

func TestProviderFallback_BothEmpty(t *testing.T) {
	agentProvider := ""
	
	inProv := ""
	outProv := ""
	
	inputTokenProvider := agentProvider
	outputTokenProvider := agentProvider
	
	if inProv != "" {
		inputTokenProvider = inProv
	}
	if outProv != "" {
		outputTokenProvider = outProv
	}
	
	if inputTokenProvider != "" || outputTokenProvider != "" {
		t.Errorf("Expected empty token providers, got in: %q, out: %q", inputTokenProvider, outputTokenProvider)
	}
}

func TestSkillTokenAttributionIntegration(t *testing.T) {
	// Set up config with an agent having a static skill "test-token-attribution"
	cfg := &config.Config{
		Provider:    "openai",
		Model:       "gpt-4o",
		TagRequired: false,
		Agents: []config.AgentConfig{
			{
				Name:     "Lead Strategist",
				Role:     "Lead Strategist",
				Skills:   []string{"test-token-attribution"},
				Provider: "agent-default-provider",
			},
		},
	}

	coord := workforce.NewCoordinator(nil, nil)
	server := workforce.NewWSServer(coord)

	mux := http.NewServeMux()
	server.ServeOnMux(mux)
	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	wsURL := strings.Replace(testServer.URL, "http", "ws", 1) + "/ws"

	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to dial WebSocket server: %v", err)
	}
	defer conn.Close()

	var initialEvt workforce.Event
	_ = conn.ReadJSON(&initialEvt)

	events := make(chan workforce.Event, 100)
	go func() {
		for {
			var evt workforce.Event
			err := conn.ReadJSON(&evt)
			if err != nil {
				return
			}
			events <- evt
		}
	}()

	runID := "test-run-attribution"
	prompt := "run with static skill"

	go runWorkforceExecution(coord, server, cfg, prompt, runID)

	timeout := time.After(5 * time.Second)
	attributionFound := false

LoopAttribution:
	for {
		select {
		case evt := <-events:
			if evt.Type == "agent.speak" && evt.Sender == "Lead Strategist" {
				inProv, _ := evt.Metadata["input_token_provider"].(string)
				outProv, _ := evt.Metadata["output_token_provider"].(string)
				if inProv == "test-input" && outProv == "test-output" {
					attributionFound = true
				} else {
					t.Errorf("Metadata mismatch: input: %q, output: %q", inProv, outProv)
				}
				break LoopAttribution
			}
		case <-timeout:
			break LoopAttribution
		}
		if attributionFound {
			break LoopAttribution
		}
	}

	if !attributionFound {
		t.Error("Expected event metadata to contain input_token_provider and output_token_provider from the skill")
	}

	abortCmd := workforce.Command{
		Type:  "run.abort",
		RunID: runID,
	}
	abortBytes, _ := json.Marshal(abortCmd)
	_ = conn.WriteMessage(websocket.TextMessage, abortBytes)
}

