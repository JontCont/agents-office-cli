# agent-management Specification

## Purpose

TBD - created by archiving change 'agent-office-feature-completion'. Update Purpose after archive.

## Requirements

### Requirement: Agent creation via CLI
The system SHALL provide a CLI command `agent-office agent create` that interactively prompts the user for agent name, role, and backstory, then appends the new agent to `agent-office.yaml`. If `agent-office.yaml` does not exist, the command MUST print `"No configuration found. Run 'agent-office init' first."` and exit with code 1.

#### Scenario: Interactive agent creation succeeds
- **WHEN** the user runs `agent-office agent create` and `agent-office.yaml` exists
- **THEN** the system SHALL prompt `Agent name:`, `Role:`, and `Backstory:` in sequence
- **AND** append the new agent to the `agents` list in `agent-office.yaml`
- **AND** print `Agent '<name>' created successfully.`

#### Scenario: Agent creation fails when config missing
- **WHEN** the user runs `agent-office agent create` and `agent-office.yaml` does not exist
- **THEN** the system MUST print `"No configuration found. Run 'agent-office init' first."` and exit with code 1


<!-- @trace
source: agent-office-feature-completion
updated: 2026-06-10
code:
  - pkg/workforce/provider.go
  - pkg/config/config.go
  - .agent-office-sessions/2026-06-10-run-75549.json
  - .agent-office-sessions/2026-06-10-run-79450.json
  - pkg/workforce/server.go
  - agent-office.yaml
  - cmd/agent-office/gui/src/style.css
  - agent-office.exe
  - .agent-office-token
  - cmd/agent-office/gui/index.html
  - cmd/agent-office/main.go
  - agent-office.exe~
  - cmd/agent-office/gui/src/main.js
tests:
  - pkg/workforce/provider_test.go
  - pkg/config/config_test.go
-->

---
### Requirement: Agent listing, creation, updating, and deletion via REST API
The system SHALL expose REST endpoints for agent management when running in `gui` mode.
- `GET /api/agents` returns all configured agents as JSON.
- `POST /api/agents` accepts a JSON body with `name`, `role`, and `backstory` fields, appends the agent to `agent-office.yaml`, and returns the created agent with HTTP 201.
- `PUT /api/agents` accepts a JSON body with `originalName`, `name`, `role`, and `backstory` fields, updates the agent in `agent-office.yaml` matching `originalName`, and returns the updated agent with HTTP 200.
- `DELETE /api/agents` accepts a query parameter `name`, deletes the matching agent from `agent-office.yaml`, and returns HTTP 200.
All creation and update operations MUST require `name` and `role` to be non-empty.

#### Scenario: GET /api/agents returns agent list
- **WHEN** a client sends `GET /api/agents`
- **THEN** the server SHALL respond with HTTP 200 and body `{"agents": [{"name":"...","role":"...","backstory":"..."}]}`

#### Scenario: POST /api/agents creates agent
- **WHEN** a client sends `POST /api/agents` with body `{"name":"qa","role":"QA Engineer","backstory":"Runs tests."}`
- **THEN** the server SHALL append the agent to `agent-office.yaml` and respond with HTTP 201 and the created agent as JSON

#### Scenario: POST /api/agents rejects missing required fields
- **WHEN** a client sends `POST /api/agents` with `name` or `role` empty or missing
- **THEN** the server SHALL respond with HTTP 400 and body `{"error":"name and role are required"}`

#### Scenario: PUT /api/agents updates agent
- **WHEN** a client sends `PUT /api/agents` with body `{"originalName":"qa","name":"qa-senior","role":"Senior QA Engineer","backstory":"Runs complex tests."}`
- **THEN** the server SHALL update the matching agent in `agent-office.yaml` and respond with HTTP 200 and the updated agent as JSON

#### Scenario: DELETE /api/agents deletes agent
- **WHEN** a client sends `DELETE /api/agents?name=qa`
- **THEN** the server SHALL delete the matching agent from `agent-office.yaml` and respond with HTTP 200


<!-- @trace
source: agent-office-feature-completion
updated: 2026-06-10
code:
  - pkg/workforce/provider.go
  - pkg/config/config.go
  - .agent-office-sessions/2026-06-10-run-75549.json
  - .agent-office-sessions/2026-06-10-run-79450.json
  - pkg/workforce/server.go
  - agent-office.yaml
  - cmd/agent-office/gui/src/style.css
  - agent-office.exe
  - .agent-office-token
  - cmd/agent-office/gui/index.html
  - cmd/agent-office/main.go
  - agent-office.exe~
  - cmd/agent-office/gui/src/main.js
tests:
  - pkg/workforce/provider_test.go
  - pkg/config/config_test.go
-->

---
### Requirement: Agents management panel in GUI
The system SHALL add an Agents tab to the GUI companion dashboard. The tab SHALL display a table of all configured agents (name, role, backstory) with Edit and Delete actions for each row, and an Add Agent button that reveals an inline form for agent creation.
- Clicking Edit SHALL populate the inline form with the agent's current values and switch the form to Edit mode.
- Submitting the form in Edit mode SHALL send a `PUT /api/agents` request and refresh the table on success.
- Clicking Delete SHALL trigger a confirmation and send a `DELETE /api/agents?name=<name>` request on approval, refreshing the table on success.

#### Scenario: Agents tab displays configured agents
- **WHEN** the user opens the GUI and clicks the Agents tab
- **THEN** the GUI SHALL fetch `GET /api/agents` and display each agent's name, role, and backstory in a table with Edit and Delete icon buttons

#### Scenario: Add Agent form submits and refreshes list
- **WHEN** the user fills the Add Agent form and clicks Submit
- **THEN** the GUI SHALL POST to `/api/agents` and, on HTTP 201, refresh the agent table to include the new agent

#### Scenario: Edit Agent form submits and refreshes list
- **WHEN** the user clicks Edit on an agent, modifies fields, and clicks Submit
- **THEN** the GUI SHALL PUT to `/api/agents` and, on HTTP 200, refresh the agent table to show the updated values

#### Scenario: Delete Agent triggers confirmation and refreshes list
- **WHEN** the user clicks Delete on an agent and confirms
- **THEN** the GUI SHALL send `DELETE /api/agents?name=<name>` and, on HTTP 200, refresh the agent table to remove the deleted agent

<!-- @trace
source: agent-office-feature-completion
updated: 2026-06-10
code:
  - pkg/workforce/provider.go
  - pkg/config/config.go
  - .agent-office-sessions/2026-06-10-run-75549.json
  - .agent-office-sessions/2026-06-10-run-79450.json
  - pkg/workforce/server.go
  - agent-office.yaml
  - cmd/agent-office/gui/src/style.css
  - agent-office.exe
  - .agent-office-token
  - cmd/agent-office/gui/index.html
  - cmd/agent-office/main.go
  - agent-office.exe~
  - cmd/agent-office/gui/src/main.js
tests:
  - pkg/workforce/provider_test.go
  - pkg/config/config_test.go
-->

---
### Requirement: Agent Config Extension with Skills, Tools, and Hooks
The system SHALL extend the agent configuration model to support skills, tools, and hooks attributes. These fields SHALL be stored as arrays of strings in agent-office.yaml and mapped to Go structs.

#### Scenario: Parse agent config with skills, tools, and hooks
- **WHEN** the system loads agent-office.yaml containing skills, tools, or hooks for an agent
- **THEN** the system SHALL successfully deserialize them into the configuration structs


<!-- @trace
source: interactive-workforce-orchestration
updated: 2026-06-11
code:
  - .agent-office-sessions/2026-06-10-run-79450.json
  - .agent-office-sessions/2026-06-10-run-75549.json
  - cmd/agent-office/gui/index.html
  - pkg/workforce/coordinator.go
  - pkg/workforce/server.go
  - pkg/workforce/provider.go
  - cmd/agent-office/gui/src/style.css
  - pkg/workforce/types.go
  - agent-office.yaml
  - pkg/config/config.go
  - agent-office.exe
  - .agent-office-sessions/2026-06-10-run-79509.json
  - pkg/workforce/interruption.go
  - .agent-office-token
  - agent-office.exe~
  - cmd/agent-office/gui/src/main.js
  - cmd/agent-office/main.go
tests:
  - pkg/config/config_test.go
-->

---
### Requirement: Idle mode and dynamic task initiation
The companion GUI dashboard SHALL start in a resting IDLE state upon launching. The input textarea and action button SHALL be unlocked, allowing the user to input a starting task description.
- Submitting the task SHALL send a WebSocket command run.start containing the starting prompt.
- The coordinator SHALL receive run.start, transition state to RUNNING, and initialize the workforce loop.

#### Scenario: Dashboard starts in idle state
- **WHEN** the user launches agent-office gui and opens the dashboard
- **THEN** the status badge SHALL show QUEUED or IDLE
- **AND** the input text area and the submit button SHALL be enabled

#### Scenario: User initiates task from idle dashboard
- **WHEN** the user enters a prompt in the input area and clicks Start Task
- **THEN** the GUI SHALL send a run.start command with the prompt to the backend WebSocket server
- **AND** the backend SHALL transition the run state to RUNNING


<!-- @trace
source: interactive-workforce-orchestration
updated: 2026-06-11
code:
  - .agent-office-sessions/2026-06-10-run-79450.json
  - .agent-office-sessions/2026-06-10-run-75549.json
  - cmd/agent-office/gui/index.html
  - pkg/workforce/coordinator.go
  - pkg/workforce/server.go
  - pkg/workforce/provider.go
  - cmd/agent-office/gui/src/style.css
  - pkg/workforce/types.go
  - agent-office.yaml
  - pkg/config/config.go
  - agent-office.exe
  - .agent-office-sessions/2026-06-10-run-79509.json
  - pkg/workforce/interruption.go
  - .agent-office-token
  - agent-office.exe~
  - cmd/agent-office/gui/src/main.js
  - cmd/agent-office/main.go
tests:
  - pkg/config/config_test.go
-->

---
### Requirement: Tag-based agent activation and token control
The turn routing mechanism SHALL support explicit agent tagging (e.g. @agent-name). During a workforce run:
- Only the agent whose name matches the tag in the previous message SHALL be activated to make LLM calls and consume token input.
- All other agents SHALL remain in an idle state and MUST NOT execute or calculate token inputs for that turn.

#### Scenario: Turn routed only to tagged agent
- **WHEN** an agent sends a message containing a tag @CriticalReviewer
- **THEN** the coordinator SHALL route the next turn only to the agent named CriticalReviewer
- **AND** all other configured agents SHALL remain idle and MUST NOT invoke LLM calls


<!-- @trace
source: interactive-workforce-orchestration
updated: 2026-06-11
code:
  - .agent-office-sessions/2026-06-10-run-79450.json
  - .agent-office-sessions/2026-06-10-run-75549.json
  - cmd/agent-office/gui/index.html
  - pkg/workforce/coordinator.go
  - pkg/workforce/server.go
  - pkg/workforce/provider.go
  - cmd/agent-office/gui/src/style.css
  - pkg/workforce/types.go
  - agent-office.yaml
  - pkg/config/config.go
  - agent-office.exe
  - .agent-office-sessions/2026-06-10-run-79509.json
  - pkg/workforce/interruption.go
  - .agent-office-token
  - agent-office.exe~
  - cmd/agent-office/gui/src/main.js
  - cmd/agent-office/main.go
tests:
  - pkg/config/config_test.go
-->

---
### Requirement: Agent status statistics in Telemetry Dashboard
The Telemetry Dashboard SHALL display three dynamic counters for configured agents:
- Active: Count of agents currently executing or speaking.
- Idle: Count of configured agents not currently active.
- Total: Total count of configured agents loaded from agent-office.yaml.

#### Scenario: Agent status counters update dynamically
- **WHEN** a workforce run executes a step with an active agent
- **THEN** the Telemetry Dashboard SHALL update the Active count to 1 and the Idle count to total minus 1


<!-- @trace
source: interactive-workforce-orchestration
updated: 2026-06-11
code:
  - .agent-office-sessions/2026-06-10-run-79450.json
  - .agent-office-sessions/2026-06-10-run-75549.json
  - cmd/agent-office/gui/index.html
  - pkg/workforce/coordinator.go
  - pkg/workforce/server.go
  - pkg/workforce/provider.go
  - cmd/agent-office/gui/src/style.css
  - pkg/workforce/types.go
  - agent-office.yaml
  - pkg/config/config.go
  - agent-office.exe
  - .agent-office-sessions/2026-06-10-run-79509.json
  - pkg/workforce/interruption.go
  - .agent-office-token
  - agent-office.exe~
  - cmd/agent-office/gui/src/main.js
  - cmd/agent-office/main.go
tests:
  - pkg/config/config_test.go
-->

---
### Requirement: Multilingual GUI (i18n)
The companion GUI dashboard SHALL support a language toggle button (English / 中文) in the header. Clicking the toggle SHALL dynamically translate all static label text, description boxes, placeholders, and tooltips between English and Traditional Chinese without requiring a browser reload.

#### Scenario: Language toggle updates UI translations
- **WHEN** the user clicks the language toggle button
- **THEN** the GUI SHALL update the text content of all elements dynamically using translation key mappings

<!-- @trace
source: interactive-workforce-orchestration
updated: 2026-06-11
code:
  - .agent-office-sessions/2026-06-10-run-79450.json
  - .agent-office-sessions/2026-06-10-run-75549.json
  - cmd/agent-office/gui/index.html
  - pkg/workforce/coordinator.go
  - pkg/workforce/server.go
  - pkg/workforce/provider.go
  - cmd/agent-office/gui/src/style.css
  - pkg/workforce/types.go
  - agent-office.yaml
  - pkg/config/config.go
  - agent-office.exe
  - .agent-office-sessions/2026-06-10-run-79509.json
  - pkg/workforce/interruption.go
  - .agent-office-token
  - agent-office.exe~
  - cmd/agent-office/gui/src/main.js
  - cmd/agent-office/main.go
tests:
  - pkg/config/config_test.go
-->