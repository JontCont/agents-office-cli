package main

import (
	"bufio"
	"encoding/json"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"agent-office/pkg/config"
	"agent-office/pkg/workforce"

	"golang.org/x/term"
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

// Embed the GUI directory statically inside the executable
//go:embed all:gui
var guiFS embed.FS

const DefaultConfigContent = `version: "1.0"
agents: []
`

const configPath = "agent-office.yaml"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "init":
		handleInit()
	case "agent":
		if len(os.Args) < 3 {
			fmt.Println("Error: Invalid subcommand. Did you mean 'agent list', 'agent create', 'agent edit' or 'agent delete'?")
			os.Exit(1)
		}
		switch os.Args[2] {
		case "list":
			handleAgentList()
		case "create":
			handleAgentCreate()
		case "edit":
			handleAgentEdit()
		case "delete":
			handleAgentDelete()
		default:
			fmt.Printf("Error: Unknown agent subcommand '%s'\n", os.Args[2])
			os.Exit(1)
		}
	case "run":
		handleRun()
	case "gui":
		handleGUI()
	default:
		fmt.Printf("Error: Unknown command '%s'\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  agent-office init              Initialize workspace configuration")
	fmt.Println("  agent-office agent list        List configured agents")
	fmt.Println("  agent-office agent create      Interactively create a new agent")
	fmt.Println("  agent-office agent edit        Interactively edit a configured agent")
	fmt.Println("  agent-office agent delete      Interactively delete a configured agent")
	fmt.Println("  agent-office run               Start the multi-agent runtime server")
	fmt.Println("  agent-office gui               Start the websocket server and launch companion GUI")
}

func handleInit() {
	if _, err := os.Stat(configPath); err == nil {
		fmt.Println("Workspace is already initialized. agent-office.yaml already exists.")
		return
	}

	err := os.WriteFile(configPath, []byte(DefaultConfigContent), 0644)
	if err != nil {
		fmt.Printf("Error initializing workspace: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Initialized default agent-office configuration in agent-office.yaml")
}

func handleAgentList() {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Error loading configuration: %v. Please run 'agent-office init' first.\n", err)
		os.Exit(1)
	}

	fmt.Println("Configured Agents:")
	for _, agent := range cfg.Agents {
		fmt.Printf("- %s (%s)\n", agent.Name, agent.Role)
	}
}

func handleAgentCreate() {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Println("No configuration found. Run 'agent-office init' first.")
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Agent name: ")
	nameLine, _ := reader.ReadString('\n')
	name := strings.TrimSpace(nameLine)
	if name == "" {
		fmt.Println("Error: Agent name cannot be empty")
		os.Exit(1)
	}

	fmt.Print("Role: ")
	roleLine, _ := reader.ReadString('\n')
	role := strings.TrimSpace(roleLine)
	if role == "" {
		fmt.Println("Error: Role cannot be empty")
		os.Exit(1)
	}

	fmt.Print("Backstory: ")
	backstoryLine, _ := reader.ReadString('\n')
	backstory := strings.TrimSpace(backstoryLine)

	fmt.Print("AI Provider (openai/gemini/anthropic/openrouter): ")
	providerLine, _ := reader.ReadString('\n')
	provider := strings.TrimSpace(providerLine)

	fmt.Print("Model Name: ")
	modelLine, _ := reader.ReadString('\n')
	model := strings.TrimSpace(modelLine)

	fmt.Print("API Key / Token: ")
	byteToken, err := term.ReadPassword(int(os.Stdin.Fd()))
	var token string
	if err == nil {
		token = strings.TrimSpace(string(byteToken))
	}
	fmt.Println()

	cfg.Agents = append(cfg.Agents, config.AgentConfig{
		Name:      name,
		Role:      role,
		Backstory: backstory,
		Provider:  provider,
		Model:     model,
		Token:     token,
	})

	if token != "" && provider != "" {
		_ = workforce.SaveToken(workforce.AIProvider(provider), token)
	}

	if err := config.SaveConfig(configPath, cfg); err != nil {
		fmt.Printf("Error saving configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Agent '%s' created successfully.\n", name)
}

func handleAgentEdit() {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Println("No configuration found. Run 'agent-office init' first.")
		os.Exit(1)
	}

	if len(cfg.Agents) == 0 {
		fmt.Println("No agents configured to edit.")
		return
	}

	fmt.Println("Configured Agents:")
	for i, agent := range cfg.Agents {
		fmt.Printf("  %d) %s (%s)\n", i+1, agent.Name, agent.Role)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Select agent to edit (1-%d): ", len(cfg.Agents))
	selectionLine, _ := reader.ReadString('\n')
	selectionLine = strings.TrimSpace(selectionLine)

	var idx int
	if _, err := fmt.Sscanf(selectionLine, "%d", &idx); err != nil || idx < 1 || idx > len(cfg.Agents) {
		fmt.Println("Error: Invalid selection")
		os.Exit(1)
	}
	agentIdx := idx - 1
	agent := &cfg.Agents[agentIdx]

	fmt.Printf("\nEditing '%s' (press Enter to keep current value):\n", agent.Name)

	fmt.Printf("Agent name [%s]: ", agent.Name)
	nameLine, _ := reader.ReadString('\n')
	name := strings.TrimSpace(nameLine)
	if name != "" {
		for i, a := range cfg.Agents {
			if i != agentIdx && a.Name == name {
				fmt.Printf("Error: Agent with name '%s' already exists\n", name)
				os.Exit(1)
			}
		}
		agent.Name = name
	}

	fmt.Printf("Role [%s]: ", agent.Role)
	roleLine, _ := reader.ReadString('\n')
	role := strings.TrimSpace(roleLine)
	if role != "" {
		agent.Role = role
	}

	fmt.Printf("Backstory [%s]: ", agent.Backstory)
	backstoryLine, _ := reader.ReadString('\n')
	backstory := strings.TrimSpace(backstoryLine)
	if backstory != "" {
		agent.Backstory = backstory
	}

	fmt.Printf("Theme Color [%s]: ", agent.Color)
	colorLine, _ := reader.ReadString('\n')
	color := strings.TrimSpace(colorLine)
	if color != "" {
		agent.Color = color
	}

	fmt.Printf("Avatar [%s]: ", agent.Avatar)
	avatarLine, _ := reader.ReadString('\n')
	avatar := strings.TrimSpace(avatarLine)
	if avatar != "" {
		agent.Avatar = avatar
	}

	fmt.Printf("AI Provider [%s]: ", agent.Provider)
	providerLine, _ := reader.ReadString('\n')
	provider := strings.TrimSpace(providerLine)
	if provider != "" {
		agent.Provider = provider
	}

	fmt.Printf("Model Name [%s]: ", agent.Model)
	modelLine, _ := reader.ReadString('\n')
	model := strings.TrimSpace(modelLine)
	if model != "" {
		agent.Model = model
	}

	fmt.Print("API Key / Token: ")
	byteToken, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err == nil {
		token := strings.TrimSpace(string(byteToken))
		if token != "" {
			agent.Token = token
		}
	}

	if agent.Token != "" && agent.Provider != "" {
		_ = workforce.SaveToken(workforce.AIProvider(agent.Provider), agent.Token)
	}

	if err := config.SaveConfig(configPath, cfg); err != nil {
		fmt.Printf("Error saving configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Agent '%s' updated successfully.\n", agent.Name)
}

func handleAgentDelete() {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Println("No configuration found. Run 'agent-office init' first.")
		os.Exit(1)
	}

	if len(cfg.Agents) == 0 {
		fmt.Println("No agents configured to delete.")
		return
	}

	fmt.Println("Configured Agents:")
	for i, agent := range cfg.Agents {
		fmt.Printf("  %d) %s (%s)\n", i+1, agent.Name, agent.Role)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Select agent to delete (1-%d): ", len(cfg.Agents))
	selectionLine, _ := reader.ReadString('\n')
	selectionLine = strings.TrimSpace(selectionLine)

	var idx int
	if _, err := fmt.Sscanf(selectionLine, "%d", &idx); err != nil || idx < 1 || idx > len(cfg.Agents) {
		fmt.Println("Error: Invalid selection")
		os.Exit(1)
	}
	agentIdx := idx - 1
	agentName := cfg.Agents[agentIdx].Name

	fmt.Printf("Are you sure you want to delete agent '%s'? (y/n): ", agentName)
	confirmLine, _ := reader.ReadString('\n')
	confirm := strings.ToLower(strings.TrimSpace(confirmLine))

	if confirm == "y" || confirm == "yes" {
		cfg.Agents = append(cfg.Agents[:agentIdx], cfg.Agents[agentIdx+1:]...)
		if err := config.SaveConfig(configPath, cfg); err != nil {
			fmt.Printf("Error saving configuration: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Agent '%s' deleted successfully.\n", agentName)
	} else {
		fmt.Println("Deletion cancelled.")
	}
}

func handleRun() {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Error loading configuration: %v. Please run 'agent-office init' first.\n", err)
		os.Exit(1)
	}

	providerStr := cfg.Provider
	if providerStr == "" && len(cfg.Agents) > 0 {
		providerStr = cfg.Agents[0].Provider
	}
	if providerStr == "" {
		providerStr = "openai"
	}
	modelStr := cfg.Model
	if modelStr == "" && len(cfg.Agents) > 0 {
		modelStr = cfg.Agents[0].Model
	}
	if modelStr == "" {
		modelStr = workforce.GetDefaultModel(workforce.AIProvider(providerStr))
	}
	fmt.Printf("Starting workforce run with provider: %s, model: %s...\n", providerStr, modelStr)

	// Setup coordinator and event broad-casting
	var server *workforce.WSServer
	coord := workforce.NewCoordinator(
		func(state workforce.RunState) {
			if server != nil {
				server.BroadcastEvent(workforce.Event{
					ID:        fmt.Sprintf("state-%d", time.Now().UnixNano()),
					Type:      "state.change",
					Timestamp: time.Now().Unix(),
					Sender:    "system",
					Content:   string(state),
				})
			}
		},
		func(sender, content string) {
			if server != nil {
				server.BroadcastEvent(workforce.Event{
					ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
					Type:      "agent.speak",
					Timestamp: time.Now().Unix(),
					Sender:    sender,
					Content:   content,
					Metadata: map[string]interface{}{
						"stage": "Intervention",
					},
				})
			}
		},
	)

	server = workforce.NewWSServer(coord)

	// Start Mock Workforce Simulation
	startMockRun(coord, server, cfg)

	fmt.Println("Starting local WebSocket IPC server on :8080...")
	log.Fatal(server.Start(":8080"))
}

// latestSessionLog returns the path of the most recently written session log.
func latestSessionLog() string {
	entries, err := os.ReadDir(workforce.SessionLogDir)
	if err != nil || len(entries) == 0 {
		return ""
	}
	// Entries are sorted alphabetically; last one is latest date-prefixed file
	latest := entries[len(entries)-1]
	return workforce.SessionLogDir + "/" + latest.Name()
}

func handleGUI() {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Error loading configuration: %v. Please run 'agent-office init' first.\n", err)
		os.Exit(1)
	}

	providerStr := cfg.Provider
	if providerStr == "" && len(cfg.Agents) > 0 {
		providerStr = cfg.Agents[0].Provider
	}
	if providerStr == "" {
		providerStr = "openai"
	}
	modelStr := cfg.Model
	if modelStr == "" && len(cfg.Agents) > 0 {
		modelStr = cfg.Agents[0].Model
	}
	if modelStr == "" {
		modelStr = workforce.GetDefaultModel(workforce.AIProvider(providerStr))
	}
	fmt.Printf("Starting companion GUI with provider: %s, model: %s...\n", providerStr, modelStr)

	// Setup coordinator and event broad-casting
	var server *workforce.WSServer
	coord := workforce.NewCoordinator(
		func(state workforce.RunState) {
			if server != nil {
				server.BroadcastEvent(workforce.Event{
					ID:        fmt.Sprintf("state-%d", time.Now().UnixNano()),
					Type:      "state.change",
					Timestamp: time.Now().Unix(),
					Sender:    "system",
					Content:   string(state),
				})
			}
		},
		func(sender, content string) {
			if server != nil {
				server.BroadcastEvent(workforce.Event{
					ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
					Type:      "agent.speak",
					Timestamp: time.Now().Unix(),
					Sender:    sender,
					Content:   content,
					Metadata: map[string]interface{}{
						"stage": "Intervention",
					},
				})
			}
		},
	)

	server = workforce.NewWSServer(coord)

	// Start Mock Workforce Simulation
	startMockRun(coord, server, cfg)

	// Build explicit ServeMux
	mux := http.NewServeMux()

	// Register WebSocket handler on mux
	server.ServeOnMux(mux)

	// REST: GET /api/config
	mux.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		rate := workforce.GetCostPer1KTokens(workforce.AIProvider(providerStr))
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"provider":    providerStr,
			"cost_per_1k": rate,
		})
	})

	// REST: /api/agents (GET list, POST create, PUT update, DELETE delete)
	mux.HandleFunc("/api/agents", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetAgents(w, r)
		case http.MethodPost:
			handlePostAgents(w, r)
		case http.MethodPut:
			handlePutAgents(w, r)
		case http.MethodDelete:
			handleDeleteAgents(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// REST: POST /api/agents/test
	mux.HandleFunc("/api/agents/test", handleTestAgentConnection)

	// REST: GET /api/skills
	mux.HandleFunc("/api/skills", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		skillsDir := filepath.Join(".agents", "skills")
		var skills []string
		files, err := os.ReadDir(skillsDir)
		if err == nil {
			for _, file := range files {
				if file.IsDir() {
					skillPath := filepath.Join(skillsDir, file.Name(), "SKILL.md")
					if _, err := os.Stat(skillPath); err == nil {
						skills = append(skills, file.Name())
					}
				}
			}
		}
		if skills == nil {
			skills = []string{}
		}
		_ = json.NewEncoder(w).Encode(skills)
	})

	// REST: GET /api/session/latest
	mux.HandleFunc("/api/session/latest", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		path := latestSessionLog()
		_ = json.NewEncoder(w).Encode(map[string]string{"path": path})
	})

	// REST: GET /api/session/latest/content
	mux.HandleFunc("/api/session/latest/content", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		path := latestSessionLog()
		if path == "" {
			http.Error(w, "No session log found", http.StatusNotFound)
			return
		}
		data, err := os.ReadFile(path)
		if err != nil {
			http.Error(w, "Error reading session log", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	})

	// Static file server for embedded GUI
	guiSubFS, err := fs.Sub(guiFS, "gui")
	if err != nil {
		log.Fatalf("Failed to load embedded GUI assets: %v", err)
	}
	mux.Handle("/", http.FileServer(http.FS(guiSubFS)))

	// Check if port 8080 is available
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Port 8080 is already in use. Please terminate other runs before launching GUI: %v", err)
	}

	fmt.Println("Starting Companion Dashboard Server on http://localhost:8080...")

	// Launch browser in a background goroutine
	go func() {
		time.Sleep(500 * time.Millisecond)
		openBrowser("http://localhost:8080")
	}()

	// Serve with explicit mux
	log.Fatal(http.Serve(listener, mux))
}

func handleGetAgents(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "configuration not found"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"agents": cfg.Agents})
}

func handlePostAgents(w http.ResponseWriter, r *http.Request) {
	var body config.AgentConfig
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON body"})
		return
	}

	if strings.TrimSpace(body.Name) == "" || strings.TrimSpace(body.Role) == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "name and role are required"})
		return
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "configuration not found"})
		return
	}

	cfg.Agents = append(cfg.Agents, body)
	if body.Token != "" && body.Provider != "" {
		_ = workforce.SaveToken(workforce.AIProvider(body.Provider), body.Token)
	}
	if err := config.SaveConfig(configPath, cfg); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to save configuration"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(body)
}

type UpdateAgentRequest struct {
	OriginalName string `json:"originalName"`
	Name         string `json:"name"`
	Role         string `json:"role"`
	Backstory    string `json:"backstory"`
	Color        string `json:"color"`
	Avatar       string `json:"avatar"`
	Provider     string `json:"provider"`
	Model        string `json:"model"`
	Token        string `json:"token"`
}

func handlePutAgents(w http.ResponseWriter, r *http.Request) {
	var body UpdateAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON body"})
		return
	}

	body.OriginalName = strings.TrimSpace(body.OriginalName)
	body.Name = strings.TrimSpace(body.Name)
	body.Role = strings.TrimSpace(body.Role)
	body.Backstory = strings.TrimSpace(body.Backstory)

	if body.OriginalName == "" || body.Name == "" || body.Role == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "originalName, name, and role are required"})
		return
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "configuration not found"})
		return
	}

	foundIdx := -1
	for i, a := range cfg.Agents {
		if a.Name == body.OriginalName {
			foundIdx = i
		} else if a.Name == body.Name {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "agent with name '" + body.Name + "' already exists"})
			return
		}
	}

	if foundIdx == -1 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "agent not found"})
		return
	}

	cfg.Agents[foundIdx].Name = body.Name
	cfg.Agents[foundIdx].Role = body.Role
	cfg.Agents[foundIdx].Backstory = body.Backstory
	cfg.Agents[foundIdx].Color = body.Color
	cfg.Agents[foundIdx].Avatar = body.Avatar
	cfg.Agents[foundIdx].Provider = body.Provider
	cfg.Agents[foundIdx].Model = body.Model
	cfg.Agents[foundIdx].Token = body.Token

	if body.Token != "" && body.Provider != "" {
		_ = workforce.SaveToken(workforce.AIProvider(body.Provider), body.Token)
	}

	if err := config.SaveConfig(configPath, cfg); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to save configuration"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(cfg.Agents[foundIdx])
}

func handleDeleteAgents(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	if name == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "name parameter is required"})
		return
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "configuration not found"})
		return
	}

	foundIdx := -1
	for i, a := range cfg.Agents {
		if a.Name == name {
			foundIdx = i
			break
		}
	}

	if foundIdx == -1 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "agent not found"})
		return
	}

	// Remove from slice
	cfg.Agents = append(cfg.Agents[:foundIdx], cfg.Agents[foundIdx+1:]...)

	if err := config.SaveConfig(configPath, cfg); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to save configuration"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		fmt.Printf("Failed to open browser: %v\n", err)
	}
}

func startMockRun(coord *workforce.Coordinator, server *workforce.WSServer, cfg *config.Config) {
	go func() {
		for {
			prompt, ok := <-coord.GetStartPromptChan()
			if !ok {
				return
			}

			// Generate a unique run ID for this execution
			runID := "run-" + fmt.Sprintf("%d", time.Now().Unix()%100000)

			// Reload configuration from agent-office.yaml to get the latest agents, colors, tokens, etc.
			latestCfg, err := config.LoadConfig(configPath)
			if err != nil {
				// Fallback to the server start config if reloading fails
				latestCfg = cfg
			}

			// Run collaboration loop synchronously in this iteration of the loop
			runWorkforceExecution(coord, server, latestCfg, prompt, runID)
		}
	}()
}

func isSelfIntroductionRequest(history []ChatMessage) bool {
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].Role == "user" {
			content := strings.ToLower(history[i].Content)
			if strings.Contains(content, "自我介紹") ||
				strings.Contains(content, "介紹自己") ||
				strings.Contains(content, "介紹一下") ||
				strings.Contains(content, "introduce yourself") ||
				strings.Contains(content, "introduce yourselves") ||
				strings.Contains(content, "who are you") {
				return true
			}
		}
	}
	return false
}

func runWorkforceExecution(coord *workforce.Coordinator, server *workforce.WSServer, cfg *config.Config, prompt string, runID string) {
	providerStr := cfg.Provider
	if providerStr == "" && len(cfg.Agents) > 0 {
		providerStr = cfg.Agents[0].Provider
	}
	if providerStr == "" {
		providerStr = "openai"
	}
	modelStr := cfg.Model
	if modelStr == "" && len(cfg.Agents) > 0 {
		modelStr = cfg.Agents[0].Model
	}
	if modelStr == "" {
		modelStr = workforce.GetDefaultModel(workforce.AIProvider(providerStr))
	}

	startedAt := time.Now().Unix()
	totalTokens := 0
	totalSteps := 0

	// Small buffer before beginning the flow
	time.Sleep(1 * time.Second)

	if len(cfg.Agents) == 0 {
		_ = coord.Transition(workforce.StateFailed)
		server.BroadcastEvent(workforce.Event{
			ID:        "evt-fail-state",
			RunID:     runID,
			Type:      "state.change",
			Timestamp: time.Now().Unix(),
			Sender:    "system",
			Content:   string(workforce.StateFailed),
		})
		server.BroadcastEvent(workforce.Event{
			ID:        "evt-fail-log",
			RunID:     runID,
			Type:      "system.log",
			Timestamp: time.Now().Unix(),
			Sender:    "system",
			Content:   "Error: No agents configured in workspace. Please add agents first.",
		})
		return
	}

	// Transition to running
	_ = coord.Transition(workforce.StateRunning)
	server.BroadcastEvent(workforce.Event{
		ID:        "evt-0",
		RunID:     runID,
		Type:      "state.change",
		Timestamp: time.Now().Unix(),
		Sender:    "system",
		Content:   string(workforce.StateRunning),
	})

	// Broadcast system.log when task starts
	server.BroadcastEvent(workforce.Event{
		ID:        fmt.Sprintf("evt-start-log-%d", time.Now().UnixNano()),
		RunID:     runID,
		Type:      "system.log",
		Timestamp: time.Now().Unix(),
		Sender:    "system",
		Content:   fmt.Sprintf("Task started: %s", prompt),
	})

	// Inject the starting prompt as the initial message in the thread
	server.BroadcastEvent(workforce.Event{
		ID:        fmt.Sprintf("evt-prompt-%d", time.Now().UnixNano()),
		RunID:     runID,
		Type:      "agent.speak",
		Timestamp: time.Now().Unix(),
		Sender:    "User",
		Content:   prompt,
		Metadata: map[string]interface{}{
			"stage": "Planning",
		},
	})

	history := []ChatMessage{
		{Role: "user", Content: prompt, Name: "User"},
	}

	everyoneActive := false
	everyoneSpoken := make(map[string]bool)
	isUserMessage := true

	lastMessage := prompt
	for i := 0; i < 20; i++ {
		// Step boundary check (blocks when interrupted)
		if ok, _ := coord.CheckStepBoundary(); !ok {
			server.BroadcastEvent(workforce.Event{
				ID:        fmt.Sprintf("evt-abort-%d", i),
				RunID:     runID,
				Type:      "system.log",
				Timestamp: time.Now().Unix(),
				Sender:    "system",
				Content:   "Run aborted by supervisor.",
			})
			// Write session log for cancelled run
			endedAt := time.Now().Unix()
			rate := workforce.GetCostPer1KTokens(workforce.AIProvider(providerStr))
			_ = workforce.WriteSessionLog(workforce.SessionLog{
				RunID:            runID,
				Provider:         providerStr,
				Model:            modelStr,
				StartedAt:        startedAt,
				EndedAt:          endedAt,
				TotalTokens:      totalTokens,
				EstimatedCostUSD: float64(totalTokens) / 1000.0 * rate,
				Steps:            totalSteps,
			})
			return
		}

		// Convert config.AgentConfig to workforce.Agent for RouteTurn
		wfAgents := make([]workforce.Agent, len(cfg.Agents))
		for idx, a := range cfg.Agents {
			wfAgents[idx] = workforce.Agent{
				Name:      a.Name,
				Role:      a.Role,
				Backstory: a.Backstory,
				Skills:    a.Skills,
				Tools:     a.Tools,
				Hooks:     a.Hooks,
				Color:     a.Color,
				Avatar:    a.Avatar,
				Provider:  a.Provider,
				Model:     a.Model,
			}
		}

		// Determine speaker based on routing
		hasMention := false
		lowerMsg := strings.ToLower(lastMessage)
		if strings.Contains(lowerMsg, "@user") || strings.Contains(lowerMsg, "@supervisor") || strings.Contains(lowerMsg, "@human") {
			hasMention = true
		} else {
			for _, a := range cfg.Agents {
				name := strings.ToLower(a.Name)
				norm := strings.ReplaceAll(name, " ", "")
				hyphenated := strings.ReplaceAll(name, " ", "-")
				underscored := strings.ReplaceAll(name, " ", "_")
				if strings.Contains(lowerMsg, "@"+name) || strings.Contains(lowerMsg, "@"+norm) || strings.Contains(lowerMsg, "@"+hyphenated) || strings.Contains(lowerMsg, "@"+underscored) {
					hasMention = true
					break
				}
			}
		}

		// Check for @everyone activation
		if isUserMessage {
			containsEveryone := strings.Contains(lowerMsg, "@everyone")
			hasOtherExplicitAgentMention := false
			for _, a := range cfg.Agents {
				name := strings.ToLower(a.Name)
				norm := strings.ReplaceAll(name, " ", "")
				hyphenated := strings.ReplaceAll(name, " ", "-")
				underscored := strings.ReplaceAll(name, " ", "_")
				if strings.Contains(lowerMsg, "@"+name) || strings.Contains(lowerMsg, "@"+norm) || strings.Contains(lowerMsg, "@"+hyphenated) || strings.Contains(lowerMsg, "@"+underscored) {
					hasOtherExplicitAgentMention = true
					break
				}
			}

			if containsEveryone && !hasOtherExplicitAgentMention {
				everyoneActive = true
				allSpoken := true
				for _, a := range cfg.Agents {
					if !everyoneSpoken[a.Name] {
						allSpoken = false
						break
					}
				}
				if allSpoken {
					everyoneSpoken = make(map[string]bool)
				}
			} else if hasOtherExplicitAgentMention {
				everyoneActive = false
			}
		}

		speaker := ""
		// 1. If @everyone is active, strictly route to the next agent who hasn't spoken yet
		if everyoneActive {
			var nextAgentName string
			for _, a := range cfg.Agents {
				if !everyoneSpoken[a.Name] {
					nextAgentName = a.Name
					break
				}
			}

			if nextAgentName != "" {
				speaker = nextAgentName
				everyoneSpoken[nextAgentName] = true
			}

			// Check if all agents have spoken
			allSpoken := true
			for _, a := range cfg.Agents {
				if !everyoneSpoken[a.Name] {
					allSpoken = false
					break
				}
			}
			if allSpoken {
				everyoneActive = false
			}
		}

		// 2. If not @everyone, check if there is an explicit agent mention first
		if speaker == "" {
			var mentionedAgent string
			for _, a := range cfg.Agents {
				name := strings.ToLower(a.Name)
				norm := strings.ReplaceAll(name, " ", "")
				hyphenated := strings.ReplaceAll(name, " ", "-")
				underscored := strings.ReplaceAll(name, " ", "_")
				if strings.Contains(lowerMsg, "@"+name) || strings.Contains(lowerMsg, "@"+norm) || strings.Contains(lowerMsg, "@"+hyphenated) || strings.Contains(lowerMsg, "@"+underscored) {
					mentionedAgent = a.Name
					break
				}
			}
			if mentionedAgent != "" {
				speaker = mentionedAgent
			}
		}

		// 3. If no agent mention, check for human supervisor handoff
		if speaker == "" {
			if strings.Contains(lowerMsg, "@user") || strings.Contains(lowerMsg, "@supervisor") || strings.Contains(lowerMsg, "@human") {
				speaker = "User"
			}
		}

		// 4. Check for self-introductions fallback
		if speaker == "" {
			if !hasMention && isSelfIntroductionRequest(history) {
				spoken := make(map[string]bool)
				for _, msg := range history {
					if msg.Role == "assistant" {
						spoken[strings.ToLower(msg.Name)] = true
					}
				}
				for _, a := range cfg.Agents {
					normalizedName := strings.ReplaceAll(strings.ToLower(a.Name), " ", "_")
					if !spoken[normalizedName] {
						speaker = a.Name
						break
					}
				}
			}
		}

		// 5. Fallback routing via RouteTurn
		if speaker == "" {
			speaker = workforce.RouteTurn(lastMessage, "Planning", wfAgents, cfg.Agents[0].Name)
		}

		if speaker == "User" {
			_ = coord.AskHuman()
			ok, feedback := coord.CheckStepBoundary()
			if !ok {
				server.BroadcastEvent(workforce.Event{
					ID:        fmt.Sprintf("evt-abort-%d", i),
					RunID:     runID,
					Type:      "system.log",
					Timestamp: time.Now().Unix(),
					Sender:    "system",
					Content:   "Run aborted by supervisor.",
				})
				// Write session log for cancelled run
				endedAt := time.Now().Unix()
				rate := workforce.GetCostPer1KTokens(workforce.AIProvider(providerStr))
				_ = workforce.WriteSessionLog(workforce.SessionLog{
					RunID:            runID,
					Provider:         providerStr,
					Model:            modelStr,
					StartedAt:        startedAt,
					EndedAt:          endedAt,
					TotalTokens:      totalTokens,
					EstimatedCostUSD: float64(totalTokens) / 1000.0 * rate,
					Steps:            totalSteps,
				})
				return
			}
			lastMessage = feedback
			history = append(history, ChatMessage{
				Role:    "user",
				Content: feedback,
				Name:    "Supervisor",
			})
			isUserMessage = true
			i--
			continue
		}

		// Find active agent details
		var activeAgent *config.AgentConfig
		for idx := range cfg.Agents {
			if strings.EqualFold(cfg.Agents[idx].Name, speaker) {
				activeAgent = &cfg.Agents[idx]
				break
			}
		}
		if activeAgent == nil {
			activeAgent = &cfg.Agents[0]
			speaker = activeAgent.Name
		}

		activeColor := activeAgent.Color
		activeAvatar := activeAgent.Avatar
		activeProvider := providerStr
		if activeAgent.Provider != "" {
			activeProvider = activeAgent.Provider
		}
		activeModel := modelStr
		if activeAgent.Model != "" {
			activeModel = activeAgent.Model
		}

		// Build system instruction
		var otherAgentNames []string
		for _, a := range cfg.Agents {
			if a.Name != activeAgent.Name {
				otherAgentNames = append(otherAgentNames, "@"+a.Name)
			}
		}
		var systemPrompt string
		globalPromptPath := filepath.Join(".agents", "system_prompt.md")
		if data, err := os.ReadFile(globalPromptPath); err == nil {
			tmpl := string(data)
			tmpl = strings.ReplaceAll(tmpl, "{{.Name}}", activeAgent.Name)
			tmpl = strings.ReplaceAll(tmpl, "{{.Role}}", activeAgent.Role)
			tmpl = strings.ReplaceAll(tmpl, "{{.Backstory}}", activeAgent.Backstory)
			tmpl = strings.ReplaceAll(tmpl, "{{.OtherAgents}}", strings.Join(otherAgentNames, ", "))
			systemPrompt = tmpl
		} else {
			systemPrompt = fmt.Sprintf("You are an AI agent named %s with the role '%s' and backstory: %s. "+
				"You are collaborating with other agents in a workforce to solve the user's task. Respond to the conversation history in character. "+
				"Keep your response concise and focused on solving the user's task. "+
				"To tag or hand off to another agent, mention them with @AgentName (available other agents: %s). "+
				"To request feedback or ask a question to the human supervisor, mention @User. Tagging @User will automatically pause execution to wait for their input. "+
				"If the task is complete and nothing else needs to be done, summarize the results and state that the work is finalized.",
				activeAgent.Name, activeAgent.Role, activeAgent.Backstory, strings.Join(otherAgentNames, ", "))
		}

		// Load static skills from configuration
		var staticSkillsPrompt string
		for _, sk := range activeAgent.Skills {
			if skContent := loadSkillPrompt(sk); skContent != "" {
				staticSkillsPrompt += fmt.Sprintf("\n\n=== Skill: %s ===\n%s", sk, skContent)
			}
		}

		// Load dynamic skills parsed from the last message
		var dynamicSkillsPrompt string
		dynamicSkills := extractDynamicSkills(lastMessage, speaker, cfg.Agents)
		for _, sk := range dynamicSkills {
			if skContent := loadSkillPrompt(sk); skContent != "" {
				dynamicSkillsPrompt += fmt.Sprintf("\n\n=== Dynamic Skill: %s ===\n%s", sk, skContent)
			}
		}

		systemPrompt += staticSkillsPrompt + dynamicSkillsPrompt

		// Call LLM
		var content string
		var tokens int
		var latency float64
		var err error

		startTime := time.Now()
		// Fetch the key for testing or execution
		agentToken := activeAgent.Token
		if agentToken == "" && activeProvider != "" {
			if gToken, ok := workforce.GetToken(workforce.AIProvider(activeProvider)); ok {
				agentToken = gToken
			}
		}

		if activeProvider != "" && agentToken != "" {
			content, tokens, err = callLLMDirectly(activeProvider, agentToken, activeModel, systemPrompt, history)
			latency = time.Since(startTime).Seconds()
		}

		// Fallback if real call fails or not configured
		if err != nil || activeProvider == "" || agentToken == "" {
			mockTexts := map[string]string{
				"lead strategist":    "Requirements loaded. Handing off to @Technical Architect to begin implementation.",
				"technical architect": "Scaffolding complete. Ready for review @Critical Reviewer.",
				"critical reviewer":   "Code looks solid. Running unit tests. @Lead Strategist, please finalize.",
			}
			lowerSpeaker := strings.ToLower(speaker)
			content = mockTexts[lowerSpeaker]
			if content == "" {
				content = "Hello, I am " + speaker + " working on the task."
			}
			if err != nil {
				content = fmt.Sprintf("[LLM Connection Failed: %v] ", err) + content
			}
			tokens = 120
			latency = 1.0
		}

		// Perform step
		totalTokens += tokens
		totalSteps++

		totalAgents := len(cfg.Agents)
		activeAgents := 1
		idleAgents := totalAgents - activeAgents

		server.BroadcastEvent(workforce.Event{
			ID:        fmt.Sprintf("evt-step-%d", i),
			RunID:     runID,
			Type:      "agent.speak",
			Timestamp: time.Now().Unix(),
			Sender:    speaker,
			Content:   content,
			Metadata: map[string]interface{}{
				"tokens":        tokens,
				"stage":         "Execution",
				"latency":       latency,
				"active_agents": activeAgents,
				"idle_agents":   idleAgents,
				"total_agents":  totalAgents,
				"color":         activeColor,
				"avatar":        activeAvatar,
				"provider":      activeProvider,
				"model":         activeModel,
			},
		})

		lastMessage = content
		isUserMessage = false
		history = append(history, ChatMessage{
			Role:    "assistant",
			Content: content,
			Name:    strings.ReplaceAll(speaker, " ", "_"),
		})

		// Stop if finalized
		if strings.Contains(strings.ToLower(content), "finalized") || strings.Contains(strings.ToLower(content), "完結") || strings.Contains(strings.ToLower(content), "conclude") {
			break
		}

		// Delay between steps to simulate processing
		time.Sleep(3 * time.Second)
	}

	// Final boundary check before completing
	if ok, _ := coord.CheckStepBoundary(); !ok {
		server.BroadcastEvent(workforce.Event{
			ID:        "evt-abort-final",
			RunID:     runID,
			Type:      "system.log",
			Timestamp: time.Now().Unix(),
			Sender:    "system",
			Content:   "Run aborted by supervisor.",
		})
		// Write session log for cancelled run
		endedAt := time.Now().Unix()
		rate := workforce.GetCostPer1KTokens(workforce.AIProvider(providerStr))
		_ = workforce.WriteSessionLog(workforce.SessionLog{
			RunID:            runID,
			Provider:         providerStr,
			Model:            modelStr,
			StartedAt:        startedAt,
			EndedAt:          endedAt,
			TotalTokens:      totalTokens,
			EstimatedCostUSD: float64(totalTokens) / 1000.0 * rate,
			Steps:            totalSteps,
		})
		return
	}

	// Complete run
	if err := coord.Transition(workforce.StateCompleted); err != nil {
		return
	}
	endedAt := time.Now().Unix()
	server.BroadcastEvent(workforce.Event{
		ID:        "evt-final",
		RunID:     runID,
		Type:      "state.change",
		Timestamp: time.Now().Unix(),
		Sender:    "system",
		Content:   string(workforce.StateCompleted),
	})

	// Write session log
	rate := workforce.GetCostPer1KTokens(workforce.AIProvider(providerStr))
	_ = workforce.WriteSessionLog(workforce.SessionLog{
		RunID:            runID,
		Provider:         providerStr,
		Model:            modelStr,
		StartedAt:        startedAt,
		EndedAt:          endedAt,
		TotalTokens:      totalTokens,
		EstimatedCostUSD: float64(totalTokens) / 1000.0 * rate,
		Steps:            totalSteps,
	})
}

func callLLMDirectly(providerStr, token, model, systemPrompt string, history interface{}) (string, int, error) {
	histSource := history.([]ChatMessage)
	hist := make([]ChatMessage, len(histSource))
	for i, msg := range histSource {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		hist[i] = ChatMessage{
			Role:    role,
			Content: msg.Content,
			Name:    msg.Name,
		}
	}
	if len(hist)%2 == 0 {
		hist = append(hist, ChatMessage{
			Role:    "user",
			Content: "[System] It is your turn to speak. Please respond in character matching your system prompt.",
			Name:    "System",
		})
	}

	client := &http.Client{Timeout: 30 * time.Second}
	var url string
	var reqBody []byte
	var err error

	switch providerStr {
	case "openrouter", "openai":
		if providerStr == "openrouter" {
			url = "https://openrouter.ai/api/v1/chat/completions"
		} else {
			url = "https://api.openai.com/v1/chat/completions"
		}
		if model == "" {
			if providerStr == "openrouter" {
				model = "google/gemini-flash-1.5"
			} else {
				model = "gpt-4o"
			}
		}

		messages := []ChatMessage{}
		if systemPrompt != "" {
			messages = append(messages, ChatMessage{Role: "system", Content: systemPrompt})
		}
		messages = append(messages, hist...)

		payload := map[string]interface{}{
			"model":    model,
			"messages": messages,
		}
		reqBody, _ = json.Marshal(payload)

	case "gemini":
		if model == "" {
			model = "gemini-1.5-pro"
		}
		url = fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, token)
		
		type GeminiPart struct {
			Text string `json:"text"`
		}
		type GeminiContent struct {
			Role  string       `json:"role"`
			Parts []GeminiPart `json:"parts"`
		}
		type GeminiPayload struct {
			Contents         []GeminiContent `json:"contents"`
			SystemInstruction *struct {
				Parts []GeminiPart `json:"parts"`
			} `json:"systemInstruction,omitempty"`
		}

		contents := []GeminiContent{}
		for _, h := range hist {
			role := h.Role
			if role == "assistant" {
				role = "model"
			}
			content := h.Content
			if h.Name != "" && h.Name != "User" && h.Name != "System" {
				content = fmt.Sprintf("[%s] %s", h.Name, h.Content)
			}
			contents = append(contents, GeminiContent{
				Role:  role,
				Parts: []GeminiPart{{Text: content}},
			})
		}

		payload := GeminiPayload{
			Contents: contents,
		}
		if systemPrompt != "" {
			payload.SystemInstruction = &struct {
				Parts []GeminiPart `json:"parts"`
			}{
				Parts: []GeminiPart{{Text: systemPrompt}},
			}
		}
		reqBody, _ = json.Marshal(payload)

	case "anthropic":
		if model == "" {
			model = "claude-3-5-sonnet-20241022"
		}
		url = "https://api.anthropic.com/v1/messages"

		type AnthropicMessage struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}
		anthropicHist := make([]AnthropicMessage, len(hist))
		for idx, h := range hist {
			content := h.Content
			if h.Name != "" && h.Name != "User" && h.Name != "System" {
				content = fmt.Sprintf("[%s] %s", h.Name, h.Content)
			}
			anthropicHist[idx] = AnthropicMessage{
				Role:    h.Role,
				Content: content,
			}
		}
		payload := map[string]interface{}{
			"model":      model,
			"max_tokens": 4096,
			"messages":   anthropicHist,
		}
		if systemPrompt != "" {
			payload["system"] = systemPrompt
		}
		reqBody, _ = json.Marshal(payload)

	default:
		return "", 0, fmt.Errorf("unsupported provider: %s", providerStr)
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(reqBody)))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	if providerStr == "openai" || providerStr == "openrouter" {
		req.Header.Set("Authorization", "Bearer "+token)
	} else if providerStr == "anthropic" {
		req.Header.Set("x-api-key", token)
		req.Header.Set("anthropic-version", "2023-06-01")
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("API status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var responseText string
	var tokensUsed int

	switch providerStr {
	case "openrouter", "openai":
		var result struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
			Usage struct {
				TotalTokens int `json:"total_tokens"`
			} `json:"usage"`
		}
		if err := json.Unmarshal(bodyBytes, &result); err != nil {
			return "", 0, err
		}
		if len(result.Choices) > 0 {
			responseText = result.Choices[0].Message.Content
		}
		tokensUsed = result.Usage.TotalTokens
		if tokensUsed == 0 {
			tokensUsed = len(responseText) / 4
		}

	case "gemini":
		var result struct {
			Candidates []struct {
				Content struct {
					Parts []struct {
						Text string `json:"text"`
					} `json:"parts"`
				} `json:"content"`
			} `json:"candidates"`
		}
		if err := json.Unmarshal(bodyBytes, &result); err != nil {
			return "", 0, err
		}
		if len(result.Candidates) > 0 && len(result.Candidates[0].Content.Parts) > 0 {
			responseText = result.Candidates[0].Content.Parts[0].Text
		}
		tokensUsed = len(responseText) / 4

	case "anthropic":
		var result struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
			Usage struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			} `json:"usage"`
		}
		if err := json.Unmarshal(bodyBytes, &result); err != nil {
			return "", 0, err
		}
		if len(result.Content) > 0 {
			responseText = result.Content[0].Text
		}
		tokensUsed = result.Usage.InputTokens + result.Usage.OutputTokens
	}

	return responseText, tokensUsed, nil
}

// findLatestSessionFile returns the path of the most recently modified session log file.
func findLatestSessionFile() string {
	entries, err := os.ReadDir(workforce.SessionLogDir)
	if err != nil || len(entries) == 0 {
		return ""
	}
	// Find the most recent by ModTime
	var latestPath string
	var latestTime time.Time
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().After(latestTime) {
			latestTime = info.ModTime()
			latestPath = filepath.Join(workforce.SessionLogDir, e.Name())
		}
	}
	return latestPath
}

func handleTestAgentConnection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Provider string `json:"provider"`
		Token    string `json:"token"`
		Model    string `json:"model"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON body"})
		return
	}

	if req.Provider == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "provider is required"})
		return
	}

	if req.Token == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "API Key / Token is required to test connection"})
		return
	}

	var valid bool
	var errMsg string
	var err error

	client := &http.Client{Timeout: 8 * time.Second}

	switch req.Provider {
	case "openai":
		apiReq, _ := http.NewRequest("GET", "https://api.openai.com/v1/models", nil)
		apiReq.Header.Set("Authorization", "Bearer "+req.Token)
		resp, dErr := client.Do(apiReq)
		if dErr != nil {
			err = dErr
		} else {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				valid = true
			} else {
				valid = false
				errMsg = fmt.Sprintf("OpenAI API returned status %d", resp.StatusCode)
			}
		}
	case "gemini":
		url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models?key=%s", req.Token)
		resp, dErr := client.Get(url)
		if dErr != nil {
			err = dErr
		} else {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				valid = true
			} else {
				valid = false
				errMsg = fmt.Sprintf("Gemini API returned status %d", resp.StatusCode)
			}
		}
	case "anthropic":
		payload := strings.NewReader(`{"model":"claude-3-haiku-20240307","max_tokens":1,"messages":[{"role":"user","content":"Hi"}]}`)
		apiReq, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", payload)
		apiReq.Header.Set("x-api-key", req.Token)
		apiReq.Header.Set("anthropic-version", "2023-06-01")
		apiReq.Header.Set("content-type", "application/json")
		resp, dErr := client.Do(apiReq)
		if dErr != nil {
			err = dErr
		} else {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusBadRequest {
				valid = true
			} else {
				valid = false
				errMsg = fmt.Sprintf("Anthropic API returned status %d", resp.StatusCode)
			}
		}
	case "openrouter":
		apiReq, _ := http.NewRequest("GET", "https://openrouter.ai/api/v1/auth/key", nil)
		apiReq.Header.Set("Authorization", "Bearer "+req.Token)
		resp, dErr := client.Do(apiReq)
		if dErr != nil {
			err = dErr
		} else {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				valid = true
			} else {
				valid = false
				errMsg = fmt.Sprintf("OpenRouter API returned status %d", resp.StatusCode)
			}
		}
	default:
		errMsg = "unsupported provider"
	}

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("network error: %v", err),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if valid {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   errMsg,
		})
	}
}

func loadSkillPrompt(skillName string) string {
	skillName = filepath.Clean(skillName)
	skillName = strings.ReplaceAll(skillName, "..", "")
	path := filepath.Join(".agents", "skills", skillName, "SKILL.md")
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(content)
}

func extractDynamicSkills(msg string, speaker string, agents []config.AgentConfig) []string {
	var skills []string
	words := strings.Fields(msg)
	for _, word := range words {
		if strings.HasPrefix(word, "/") {
			skillName := strings.TrimPrefix(word, "/")
			skillName = strings.TrimFunc(skillName, func(r rune) bool {
				return r == '.' || r == '?' || r == ',' || r == '!' || r == ';' || r == ':'
			})
			if skillName == "" {
				continue
			}

			// Verify if this skill exists
			skillPath := filepath.Join(".agents", "skills", skillName, "SKILL.md")
			if _, err := os.Stat(skillPath); err != nil {
				continue
			}

			// Check if this message targets speaker
			targetMatch := false
			lowerMsg := strings.ToLower(msg)

			if strings.Contains(lowerMsg, "@everyone") {
				targetMatch = true
			} else {
				speakerLower := strings.ToLower(speaker)
				speakerNorm := strings.ReplaceAll(speakerLower, " ", "")
				speakerHyphen := strings.ReplaceAll(speakerLower, " ", "-")
				speakerUnder := strings.ReplaceAll(speakerLower, " ", "_")

				if strings.Contains(lowerMsg, "@"+speakerLower) ||
					strings.Contains(lowerMsg, "@"+speakerNorm) ||
					strings.Contains(lowerMsg, "@"+speakerHyphen) ||
					strings.Contains(lowerMsg, "@"+speakerUnder) {
					targetMatch = true
				} else {
					otherAgentMentioned := false
					for _, a := range agents {
						if strings.EqualFold(a.Name, speaker) {
							continue
						}
						aLower := strings.ToLower(a.Name)
						aNorm := strings.ReplaceAll(aLower, " ", "")
						aHyphen := strings.ReplaceAll(aLower, " ", "-")
						aUnder := strings.ReplaceAll(aLower, " ", "_")
						if strings.Contains(lowerMsg, "@"+aLower) ||
							strings.Contains(lowerMsg, "@"+aNorm) ||
							strings.Contains(lowerMsg, "@"+aHyphen) ||
							strings.Contains(lowerMsg, "@"+aUnder) {
							otherAgentMentioned = true
							break
						}
					}
					if !otherAgentMentioned {
						targetMatch = true
					}
				}
			}

			if targetMatch {
				skills = append(skills, skillName)
			}
		}
	}
	return skills
}

