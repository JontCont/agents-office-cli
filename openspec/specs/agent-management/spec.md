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
The system SHALL add an Agents tab to the GUI companion dashboard. The tab SHALL display a grid of cards of all configured agents (name, role, backstory, provider, model) with Edit and Delete actions for each card, and an Add Agent button that slides in a drawer-form from the right edge for agent creation/editing.
- The drawer-form MUST slide in and out smoothly and support an independent scrollbar (`overflow-y: auto`) to ensure action buttons are never cut off.
- Selecting a provider in the AI Provider dropdown SHALL automatically pre-populate a default model name in the Model Name field.
- During form submission or connection testing, the corresponding action button SHALL show visual loading feedback and be disabled to prevent duplicate submissions.
- Clicking Edit SHALL populate the drawer-form with the agent's current values and open it.
- Submitting the form in Edit mode SHALL send a `PUT /api/agents` request and refresh the grid on success.
- Clicking Delete SHALL trigger a confirmation and send a `DELETE /api/agents?name=<name>` request on approval, refreshing the grid on success.

#### Scenario: Agents tab displays configured agents
- **WHEN** the user opens the GUI and clicks the Agents tab
- **THEN** the GUI SHALL fetch `GET /api/agents` and display each agent's name, role, and backstory in a card grid with Edit and Delete icon buttons

#### Scenario: Add Agent drawer opens, pre-fills model, and submits successfully
- **WHEN** the user clicks Add Agent
- **THEN** the drawer SHALL slide in from the right
- **AND WHEN** the user selects "Anthropic" as AI Provider
- **THEN** the Model Name input field SHALL automatically be pre-filled with "claude-3-haiku"
- **AND WHEN** the user fills the rest of the form and clicks Submit
- **THEN** the Submit button SHALL show loading state
- **AND** the GUI SHALL POST to `/api/agents` and, on HTTP 201, slide the drawer out and refresh the agent grid to include the new agent

#### Scenario: Edit Agent drawer opens and submits changes
- **WHEN** the user clicks Edit on an agent card
- **THEN** the drawer SHALL slide in from the right populated with current agent values
- **AND WHEN** the user modifies fields and clicks Submit
- **THEN** the Submit button SHALL show loading state
- **AND** the GUI SHALL PUT to `/api/agents` and, on HTTP 200, slide the drawer out and refresh the agent grid to show the updated values

#### Scenario: Delete Agent triggers confirmation and refreshes list
- **WHEN** the user clicks Delete on an agent card and confirms
- **THEN** the GUI SHALL send `DELETE /api/agents?name=<name>` and, on HTTP 200, refresh the agent grid to remove the deleted agent


<!-- @trace
source: optimize-frontend-uiux
updated: 2026-06-19
code:
  - cmd/agent-office/gui/index.html
  - cmd/agent-office/gui/src/style.css
  - skills-lock.json
  - cmd/agent-office/gui/src/main.js
  - .agents/skills/frontend-design/SKILL.md
  - .autohand/skills/frontend-design/SKILL.md
  - .autohand/skills/frontend-design
  - go.mod
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

---
### Requirement: Smart auto-scrolling and responsive layouts
The companion GUI dashboard SHALL dynamically adapt its scroll behavior and grid layout depending on user state and screen dimensions.
- The thread view SHALL only automatically scroll to the bottom upon receiving new message events if the user's scroll position is already at the bottom (within a 100px buffer).
- If the user has scrolled up, the system SHALL NOT force scroll-to-bottom, and SHALL display a floating badge indicating new messages are available.
- The main application layout elements (run thread panel, telemetry dashboard, control panel) SHALL stack vertically in a single column when the viewport width is 1024px or less, ensuring usability on small screens.

#### Scenario: Auto-scroll is bypassed when user scrolls up
- **WHEN** the user is viewing older logs (scrolled up more than 100px from bottom) and a new `agent.speak` event is received
- **THEN** the scroll position of the thread view SHALL NOT change
- **AND** a floating badge saying "New messages below ?? SHALL be displayed

#### Scenario: Auto-scroll occurs when user is at the bottom
- **WHEN** the user is at the bottom of the log view and a new `agent.speak` event is received
- **THEN** the thread view SHALL scroll smoothly to the bottom to display the new message

#### Scenario: Responsive layout wraps on smaller screens
- **WHEN** the browser window width is resized to 1000px
- **THEN** the layout container SHALL stack its panels vertically

<!-- @trace
source: optimize-frontend-uiux
updated: 2026-06-19
code:
  - cmd/agent-office/gui/index.html
  - cmd/agent-office/gui/src/style.css
  - skills-lock.json
  - cmd/agent-office/gui/src/main.js
  - .agents/skills/frontend-design/SKILL.md
  - .autohand/skills/frontend-design/SKILL.md
  - .autohand/skills/frontend-design
  - go.mod
-->