## Context

This design outlines the architecture of a CLI-first, multi-agent workforce system. The system decouples the core agent runtime from the user interface. The backend is implemented as a high-performance Go CLI executable (`agent-office`) that handles orchestration, tool execution, and LLM communication. The frontend is a standalone Tauri desktop companion application that provides visual monitoring, telemetry dashboards, and human supervisor intervention capabilities. The two processes communicate locally via WebSockets.

## Goals / Non-Goals

**Goals:**
- Provide a robust, local-first agent orchestration runtime in Go.
- Support deterministic turn-routing using explicit handoffs, role-based rules, and fallback coordination.
- Build a companion Tauri GUI that displays real-time agent messages, token telemetry, and execution state.
- Establish a reliable supervisor interruption loop that pauses execution at safe step boundaries.

**Non-Goals:**
- No remote cloud hosting or multi-tenant user accounts; database and execution are strictly local.
- No peer-to-peer private messages between agents; all communications must flow through the central Go Run Thread.
- No LLM-based turn coordination for the initial release to ensure deterministic execution.
- No sandbox container isolation (e.g., Docker) for agent tools in this MVP.

## Decisions

### Decision 1: Decoupled Tauri GUI and Go CLI with WebSocket IPC
- **Approach**: The Go CLI starts a local WebSocket server on a default port (e.g., 8080). The Tauri companion application launches separately and connects to the CLI via WS.
- **Alternatives**:
  - *Option A (Embedded HTTP/Web Server)*: Served SPA in browser. Rejected because the browser-tab UX feels less like a premium desktop dashboard.
  - *Option B (Wails Framework)*: Tight compilation of Go and UI. Rejected because it couples the CLI runtime directly to GUI windows, increasing startup latency and CLI binary size.
- **Rationale**: Decoupling preserves maximum CLI execution speed, simplifies CLI-only scripting, and allows Tauri to be built, updated, and launched independently as a dashboard client.

### Decision 2: Deterministic Go Runtime Coordinator
- **Approach**: The Go controller runs the shared discussion thread. Instead of using an LLM coordinator, turn-routing resolves sequentially: (1) check for explicit agent handoff keywords, (2) apply role-based task stage routing, (3) fallback to a default planner agent.
- **Alternatives**: LLM-based routing where an LLM is prompted to choose the next speaker. Rejected due to cost, latency, and routing unpredictability.
- **Rationale**: A deterministic controller ensures predictable execution, eliminates routing-loop token costs, and makes testing and debugging agent interactions straightforward.

### Decision 3: Safe Step-Boundary Interruption Flow
- **Approach**: Maintain a central state machine on the Go server tracking the `RunState`. When a supervisor requests an interruption, state changes to `INTERRUPTING`. The execution loop checks this state before sending any LLM request or initiating a tool call. If true, the system writes the current state, suspends execution, and moves to `INTERRUPTED`.
- **Alternatives**: Instant thread termination (kills the process but corrupts active states) or immediate pause (causes half-completed tool calls or cut-off LLM tokens).
- **Rationale**: Halting only on step boundaries avoids incomplete API responses, guarantees state consistency, and ensures that when execution resumes, agents have a complete history of the run.

## Implementation Contract

#### Process Ownership
- **Execution Owner**: The Go CLI runtime process is the sole owner of the multi-agent execution thread and event loop.
- **CLI Independence**: The Go backend (`agent-office run`) operates fully without the GUI.
- **GUI Attachment**: The Tauri application (`agent-office gui`) connects to the Go local WebSocket server to attach to the runtime, displaying the run thread and telemetry. It remains a monitor client and never owns the execution loop.

#### Protocol Model
WebSocket payloads are strictly divided into downstream **Events** and upstream **Commands**:
1. **Events** (Server ➔ Client):
   - `agent.speak`: Broadcast of agent message output.
   - `tool.call` / `tool.return`: Logs of tool executions.
   - `state.change`: Lifecycle state updates.
   - `system.log`: Execution errors and warnings.
2. **Commands** (Client ➔ Server):
   - `run.interrupt`: Supervisor request to pause the run thread.
   - `run.resume`: Supervisor instruction to resume execution with feedback.
   - `run.abort`: Supervisor instruction to terminate the run.

```go
// Event schema for downstream WebSocket streaming
type Event struct {
    ID        string                 `json:"id"`
    RunID     string                 `json:"run_id"`
    Type      string                 `json:"type"` // "agent.speak" | "tool.call" | "tool.return" | "state.change" | "system.log"
    Timestamp int64                  `json:"timestamp"`
    Sender    string                 `json:"sender"`
    Content   string                 `json:"content"`
    Metadata  map[string]interface{} `json:"metadata"`
}

// Command schema for upstream WebSocket control
type Command struct {
    Type    string `json:"type"` // "run.interrupt" | "run.resume" | "run.abort"
    RunID   string `json:"run_id"`
    Message string `json:"message,omitempty"` // Supervisor feedback for resume
}
```

#### RunState Transition Table
| Source State | Destination State | Trigger / Event | Description |
| :--- | :--- | :--- | :--- |
| `QUEUED` | `RUNNING` | Run execution starts | Initial state of a task before execution begins. |
| `RUNNING` | `INTERRUPTING` | `run.interrupt` command received | Supervisor requests pause. Go core waits for the next step boundary. |
| `RUNNING` | `COMPLETED` | Execution ends successfully | Run finished all planner phases. |
| `RUNNING` | `FAILED` | Critical runtime error or timeout | Execution halts on LLM/tool failure. |
| `RUNNING` | `INTERRUPTED` | `ask_human` tool called | Agent explicitly escalates to human supervisor. |
| `INTERRUPTING` | `INTERRUPTED` | Step boundary reached | Go core completes current turn and commits the pause. |
| `INTERRUPTED` | `RESUMING` | `run.resume` command received | Supervisor submits feedback string and signals resume. |
| `RESUMING` | `RUNNING` | Supervisor message injected | Feedback is broadcast to the thread, runtime starts next turn. |
| `INTERRUPTED` | `CANCELLED` | `run.abort` command | Supervisor aborts the run. |
| `INTERRUPTING` | `CANCELLED` | `run.abort` command | Supervisor aborts the run. |

#### Observable Behavior
- The CLI commands print logs directly to stdout. The WebSocket server listens on a dynamically allocated local port.
- The GUI attaches to this port, streams telemetry, and provides controls for pause, resume, and abort.

#### Failure Modes
- **Port Collision**: Solved via dynamic binding and discovery file writing.
- **WebSocket Drops**: Solved via client-side automatic reconnection retries.
- **Execution Failures**: LLM timeouts abort the workflow and set state to `FAILED`.

#### Acceptance Criteria
- Go runtime core runs workflows deterministically.
- WebSocket server correctly distinguishes downstream Events and upstream Commands.
- Supervisor controls halt and resume run workflows at step boundaries according to the RunState transition rules.

#### Scope Boundaries
- **In-Scope**: Local CLI execution, Tauri GUI event monitoring, local WebSocket IPC protocol, and single-user control flows.
- **Out-of-Scope**: Cloud synchronization, multi-tenant databases, GUI config builders, and sandboxed runtimes.

## Risks / Trade-offs

- **[Risk] WebSocket Port Conflict** → [Mitigation] Go server dynamically binds ports and writes the assigned port to a local configuration file for Tauri to auto-discover.
- **[Risk] State Desynchronization** → [Mitigation] The Go CLI serves as the single source of truth for execution states. The GUI only visualizes and emits command payloads.
