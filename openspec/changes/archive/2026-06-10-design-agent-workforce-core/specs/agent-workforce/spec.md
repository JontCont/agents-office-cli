## ADDED Requirements

### Requirement: Workspace initialization
The system SHALL provide a CLI command to initialize a local workspace configuration for the agent workforce runtime.

#### Scenario: Initialize a new workspace
- **GIVEN** the current directory does not contain an agent-office workspace config
- **WHEN** the user runs `agent-office init`
- **THEN** the system SHALL create the default workspace configuration files
- **AND** SHALL make the workspace ready for local agent configuration.

### Requirement: GUI attachment to local runtime
The system SHALL allow the companion GUI to connect to a local runtime instance and display active run information without becoming the execution owner.

#### Scenario: Attach GUI to existing runtime
- **GIVEN** a local runtime instance is already running
- **WHEN** the user launches `agent-office gui`
- **THEN** the GUI SHALL connect to the active runtime
- **AND** SHALL display the active run thread and telemetry.

#### Scenario: Reconnect after socket interruption
- **GIVEN** the GUI is connected to a local runtime
- **WHEN** the WebSocket connection is interrupted
- **THEN** the GUI SHALL surface the disconnected state
- **AND** SHALL retry connecting until the runtime becomes available.

### Requirement: Deterministic turn coordination
The system SHALL coordinate agent participation in a shared discussion thread using deterministic routing. For each turn, the runtime SHALL resolve the next speaker in the following order of precedence: explicit handoff, role-based routing, and fallback planner routing.

#### Scenario: Resolve speaker via explicit handoff
- **WHEN** Agent A ends its message with "@agent B, please review this"
- **THEN** Go controller MUST route the next turn to Agent B.

#### Scenario: Fallback to default planner agent
- **WHEN** no explicit handoff is found in the message and no stage rules match
- **THEN** Go controller MUST route the next turn to the default planner agent.

##### Example: Speaker routing table
| Last Speaker Message | Active Stage | Expected Next Speaker |
| --- | --- | --- |
| "Let me hand off to @reviewer" | Coding | reviewer |
| "Ready for next step" | Planning | planner |
| "Starting design phase" | Design | architect |

### Requirement: On-demand supervisor interruption
The Go server MUST support transitions of RunState requested by the supervisor. Upon receiving an interrupt signal, the Go runtime SHALL pause execution before sending the next LLM API request or triggering a tool call. The Go runtime SHALL then transition the run state to INTERRUPTED and wait for supervisor input.

#### Scenario: Graceful pause at step boundary
- **WHEN** supervisor clicks "Interrupt" while run is in state RUNNING
- **THEN** runtime state SHALL transition to INTERRUPTING and halt execution before the next API request.
- **AND** runtime state SHALL transition to INTERRUPTED and wait for supervisor input.

#### Scenario: Agent-initiated breakpoint
- **WHEN** an agent calls the ask_human tool
- **THEN** runtime state SHALL transition to INTERRUPTED and wait for supervisor input.

#### Scenario: Resuming execution with supervisor feedback
- **WHEN** run state is INTERRUPTED and supervisor submits guidance text
- **THEN** runtime state SHALL transition to RESUMING
- **AND** the supervisor text SHALL be injected as a high-priority system-level event visible to all agents.
- **AND** the runtime state SHALL return to RUNNING.

### Requirement: Real-time event streaming
The system SHALL stream discussion, tool, and run-state events to the companion GUI in real time over the local IPC channel.

#### Scenario: Update telemetry from streamed event
- **GIVEN** the GUI is connected to an active runtime
- **WHEN** the GUI receives an `agent.speak` event containing token and latency metadata
- **THEN** the GUI SHALL update the displayed run telemetry
- **AND** SHALL render the message in the run thread view.
