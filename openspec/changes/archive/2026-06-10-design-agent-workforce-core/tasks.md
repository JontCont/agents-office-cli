## 1. CLI Core and Agent Go Runtime

- [x] 1.1 Implement the Go command line entry point, config loader, and **Workspace initialization** commands. Behavior: `agent-office init` initializes a default workforce config file, and `agent-office agent list` outputs configured agents. Verification: Run `agent-office init` and inspect the generated config.
- [x] 1.2 Implement **Deterministic turn coordination** in Go according to **Decision 2: Deterministic Go Runtime Coordinator**. Behavior: The Go workforce runtime routes messages based on explicit handoffs, stage rules, or planner fallback. Verification: Run `go test ./pkg/workforce -run TestDeterministicTurnCoordination`.
- [x] 1.3 Implement the state transition engine for **On-demand supervisor interruption** as described in **Decision 3: Safe Step-Boundary Interruption Flow**. Behavior: Execution pauses at the next step boundary (before next LLM or tool call) when state transitions to `INTERRUPTING`. Verification: Run `go test ./pkg/workforce -run TestInterruptionFlow`.

## 2. WebSocket IPC Communication

- [x] 2.1 Implement the Go WebSocket communication server and JSON message serialization for **Real-time event streaming**. Behavior: A websocket server starts on launch and streams `Event` logs to client connections. Verification: Run `go run cmd/agent-office/main.go gui` and verify port 8080 accepts websocket handshakes using a websocket test utility.
- [x] 2.2 Define and implement client/server event structures in Go as defined in **Decision 1: Decoupled Tauri GUI and Go CLI with WebSocket IPC**. Behavior: All lifecycle events (agent speaks, tool calls, and human interruption) map to standard JSON structures. Verification: Review serialization unit tests in `pkg/workforce/event_test.go`.

## 3. Tauri Companion GUI Dashboard

- [x] 3.1 Scaffold the Tauri companion GUI application and implement **GUI attachment to local runtime** to connect to the Go process. Behavior: Running `agent-office gui` launches the desktop window and auto-connects to the active runtime, automatically retrying connection if interrupted. Verification: Run `agent-office gui` and verify connection state indicator in UI.
- [x] 3.2 Implement client-side WebSocket client and message rendering for **Real-time event streaming** in Tauri as part of **Decision 1: Decoupled Tauri GUI and Go CLI with WebSocket IPC**. Behavior: The GUI connects to localhost WebSocket, updates the thread viewer with incoming messages, and aggregates token/latency metrics in the telemetry panel. Verification: Start Go mock server, send mock event payloads, and assert GUI updates.
- [x] 3.3 Implement UI interruption panel and resume controls for **On-demand supervisor interruption** and **Decision 3: Safe Step-Boundary Interruption Flow**. Behavior: Clicking "Interrupt" in Tauri GUI pauses CLI execution; submitting guidance resumes the flow. Verification: End-to-end integration test by running a mock workforce workflow in Go CLI and interrupting it via GUI.
