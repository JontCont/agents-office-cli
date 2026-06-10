# Agent Office CLI & Companion GUI

`agent-office` is a high-performance, local-first multi-agent orchestration runtime built in Go. It enables developers to coordinate agent workforces deterministically and monitor executions through a decoupled companion GUI dashboard communicating over local WebSockets.

---

## Features

- **Deterministic Turn Coordination**: Avoids unpredictable and costly LLM-driven turn selection by resolving the next active agent via explicit handoffs (`@agent` / `@name`), active project stage rules, or a fallback planner agent.
- **Safe Step-Boundary Interruption**: Implements a robust state-transition engine (`QUEUED`, `RUNNING`, `INTERRUPTING`, `INTERRUPTED`, `RESUMING`, `COMPLETED`, `FAILED`, `CANCELLED`). Execution halts safely before sending LLM prompts or initiating tool calls, guaranteeing state consistency and preventing corrupted sessions.
- **Embedded Web Dashboard**: Compiles all static web assets (HTML, CSS, JavaScript) directly into the Go executable via `go:embed`. Serves the companion app on port `8080` and upgrades client connections to WebSockets over a single network port.
- **Tauri Companion GUI Scaffolding**: Provides structured configurations inside the frontend directory, pre-configured to build a native cross-platform desktop application using Tauri.
- **CLI Workspace Commands**: Quick initialization commands (`init`) to scaffold default configurations and list (`agent list`) the defined agent capabilities.

---

## Architecture

The Go backend serves as the single source of truth for the workspace, coordinating agent routing and checking step boundaries. The GUI is a detached monitor client that connects to the Go backend via WebSocket.

```mermaid
graph TD
    subgraph Go Backend Runtime (agent-office)
        CLI[Go CLI Entrypoint] --> Server[WebSocket Server :8080]
        CLI --> Config[Config Loader]
        CLI --> Coord[Deterministic Turn Coordinator]
        CLI --> Flow[Step-Boundary Interruption Engine]
    end

    subgraph Client Companion GUI
        Browser[System Browser / Tauri Window] <-->|WebSocket IPC| Server
        Browser --> Thread[Run Thread Viewer]
        Browser --> Telemetry[Metrics Dashboard]
        Browser --> Controls[Interruption / Resume Panel]
    end
```

---

## Project Structure

```
├── cmd/
│   └── agent-office/
│       ├── main.go               # Go entry point & CLI subcommands
│       └── gui/                  # Web Dashboard Frontend
│           ├── index.html        # Dashboard panel layout
│           ├── src/
│           │   ├── main.js       # Reconnect loops, event parsing, rendering
│           │   └── style.css     # Premium dark-mode glassmorphic styles
│           └── src-tauri/        # Tauri workspace Rust templates & configurations
├── pkg/
│   ├── config/
│   │   └── config.go             # YAML parser & agent configuration loader
│   └── workforce/
│       ├── types.go              # Shared RunState, Event, and Command definitions
│       ├── coordinator.go        # Deterministic turn-routing logic
│       ├── interruption.go       # Step-boundary checking & transition manager
│       └── server.go             # Upgrader & broadcast WebSocket hub
├── go.mod                        # Go module manifest
└── agent-office.yaml             # Workspace config (generated on init)
```

---

## Getting Started

### Prerequisites
- [Go 1.22+](https://go.dev/dl/)

### Installation & Compilation
Clone the repository, download dependencies, and compile the CLI executable:
```powershell
# Clean and install dependencies
go mod tidy

# Build the executable
go build -o agent-office.exe ./cmd/agent-office
```

---

## CLI Commands

### 1. Workspace Initialization
Initialize a new workspace configuration in the current working directory. This creates a default `agent-office.yaml` file:
```powershell
.\agent-office.exe init
```

### 2. Listing Configured Agents
Verify the active configuration and view the list of configured agents and roles:
```powershell
.\agent-office.exe agent list
```

### 3. Launching the GUI Dashboard Companion
Starts the WebSocket backend, serves the dashboard assets, and automatically launches your default web browser to view the real-time simulation:
```powershell
.\agent-office.exe gui
```
*Note: If you run `gui`, the console will start streaming mock agent discussion thread turns. You can use the buttons on the dashboard to **Interrupt**, **Abort**, or **Resume** (with feedback message) the active execution.*

---

## Running Tests

All core coordination logic, state transitions, and IPC packet serializations are guarded by unit tests:

```powershell
# Run the complete test suite
go test ./... -v
```
- `coordinator_test.go`: Asserts explicit keyword routing, stage routing, and planner fallbacks.
- `interruption_test.go`: Verifies thread-safe boundary blocking, resume injections, and aborts.
- `event_test.go`: Tests event serialization compliance.
