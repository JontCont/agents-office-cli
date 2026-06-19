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
The system SHALL provide a multi-tab GUI companion dashboard supporting a top navigation tab bar with exactly three tabs: Run Thread, Logs, and Agents.
- The Agents tab SHALL display a grid of cards of all configured agents (name, role, backstory, provider, model) with Edit and Delete actions for each card, and an Add Agent button that slides in a drawer-form from the right edge for agent creation/editing.
- The drawer-form MUST slide in and out smoothly using CSS transitions, and when hidden, its visibility MUST be set to hidden to prevent layout shift or off-screen focus issues.
- The drawer-form MUST place the action buttons (Create/Cancel) in a fixed footer block positioned absolutely at the bottom, while the input fields wrapper in the middle has an independent scrollbar (overflow-y: auto) with a custom modern scrollbar to ensure action buttons are never cut off.
- The right telemetry dashboard panel SHALL support category-based collapsible sections and a global collapse button to fold the entire sidebar out of view.

#### Scenario: Agents tab displays configured agents
- **WHEN** the user opens the GUI and clicks the Agents tab
- **THEN** the GUI SHALL fetch GET /api/agents and display each agent's name, role, and backstory in a card grid with Edit and Delete icon buttons

#### Scenario: Add Agent drawer opens, pre-fills model, and submits successfully
- **WHEN** the user clicks Add Agent
- **THEN** the drawer SHALL slide in from the right
- **AND WHEN** the user selects "Anthropic" as AI Provider
- **THEN** the Model Name input field SHALL automatically be pre-filled with "claude-3-haiku"
- **AND WHEN** the user fills the rest of the form and clicks Submit
- **THEN** the Submit button SHALL show loading state
- **AND** the GUI SHALL POST to /api/agents and, on HTTP 201, slide the drawer out and refresh the agent grid to include the new agent

#### Scenario: Edit Agent drawer opens and submits changes
- **WHEN** the user clicks Edit on an agent card
- **THEN** the drawer SHALL slide in from the right populated with current agent values
- **AND WHEN** the user modifies fields and clicks Submit
- **THEN** the Submit button SHALL show loading state
- **AND** the GUI SHALL PUT to /api/agents and, on HTTP 200, slide the drawer out and refresh the agent grid to show the updated values

#### Scenario: Delete Agent triggers confirmation and refreshes list
- **WHEN** the user clicks Delete on an agent card and confirms
- **THEN** the GUI SHALL send DELETE /api/agents?name=<name> and, on HTTP 200, refresh the agent grid to remove the deleted agent


<!-- @trace
source: restructure-dashboard-layout
updated: 2026-06-19
code:
  - cmd/agent-office/gui/index.html
  - .autohand/skills/frontend-design/SKILL.md
  - cmd/agent-office/gui/src/main.js
  - cmd/agent-office/gui/src/style.css
  - .autohand/skills/frontend-design
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
The companion GUI dashboard SHALL start in a resting IDLE state upon launching. The input textarea and action buttons SHALL be located at the bottom of the Run Thread panel, unlocked, allowing the user to input a starting task description like a chatroom interface.
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
source: restructure-dashboard-layout
updated: 2026-06-19
code:
  - cmd/agent-office/gui/index.html
  - .autohand/skills/frontend-design/SKILL.md
  - cmd/agent-office/gui/src/main.js
  - cmd/agent-office/gui/src/style.css
  - .autohand/skills/frontend-design
-->

---
### Requirement: Tag-based agent activation and token control
The turn routing mechanism SHALL support explicit agent tagging (e.g. @agent-name). During a workforce run:
- Only the agent whose name matches the tag in the previous message SHALL be activated to make LLM calls and consume token input.
- All other agents SHALL remain in an idle state and MUST NOT execute or calculate token inputs for that turn.
- Mentions of @User, @Supervisor, or @Human (case-insensitive) SHALL activate human handoff routing, transitioning the run state to INTERRUPTED.

#### Scenario: Turn routed only to tagged agent
- **WHEN** an agent sends a message containing a tag @CriticalReviewer
- **THEN** the coordinator SHALL route the next turn only to the agent named CriticalReviewer
- **AND** all other configured agents SHALL remain idle and MUST NOT invoke LLM calls

#### Scenario: Turn routed to human handoff
- **WHEN** an agent sends a message containing a tag @User
- **THEN** the coordinator SHALL transition the run state to INTERRUPTED
- **AND** await feedback input from the human supervisor


<!-- @trace
source: fix-agent-looping-and-feedback-sync
updated: 2026-06-19
code:
  - .autohand/skills/frontend-design
  - .agents/critical_reviewer/system_prompt.md
  - .agents/technical_architect/system_prompt.md
  - .agents/system_prompt.md
  - cmd/agent-office/main.go
  - pkg/workforce/interruption.go
  - .autohand/skills/frontend-design/SKILL.md
  - pkg/workforce/coordinator.go
  - .agents/lead_strategist/system_prompt.md
  - cmd/agent-office/gui/src/main.js
  - pkg/config/config.go
tests:
  - pkg/workforce/coordinator_test.go
  - pkg/config/config_test.go
  - pkg/workforce/interruption_test.go
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

---
### Requirement: Supervisor feedback injection in execution loop
During a workforce run, if the execution is interrupted, the supervisor SHALL be able to provide feedback text. The system MUST inject this supervisor feedback back into the conversation history as a user message in the next LLM call turn.

#### Scenario: Supervisor feedback is injected into history
- **WHEN** the execution is resumed with supervisor feedback
- **THEN** the system SHALL append the feedback content as a message with role "user" and name "Supervisor" to the conversation history
- **AND** pass the updated history in subsequent LLM calls


<!-- @trace
source: fix-agent-looping-and-feedback-sync
updated: 2026-06-19
code:
  - .autohand/skills/frontend-design
  - .agents/critical_reviewer/system_prompt.md
  - .agents/technical_architect/system_prompt.md
  - .agents/system_prompt.md
  - cmd/agent-office/main.go
  - pkg/workforce/interruption.go
  - .autohand/skills/frontend-design/SKILL.md
  - pkg/workforce/coordinator.go
  - .agents/lead_strategist/system_prompt.md
  - cmd/agent-office/gui/src/main.js
  - pkg/config/config.go
tests:
  - pkg/workforce/coordinator_test.go
  - pkg/config/config_test.go
  - pkg/workforce/interruption_test.go
-->

---
### Requirement: Filesystem-based prompt loading
The system SHALL support loading system prompt templates from the filesystem under the .agents/ folder.
- The global collaboration template SHALL be loaded from .agents/system_prompt.md.
- Each agent's backstory template SHALL be loaded from .agents/[agent_name]/system_prompt.md where the agent name is normalized to lowercase and spaces are replaced with underscores.
- If these files do not exist, the system MUST fall back to using default config settings or hardcoded string fallbacks.

#### Scenario: Global and agent-specific prompt templates loaded successfully
- **WHEN** a workforce run is started and .agents/system_prompt.md and .agents/lead_strategist/system_prompt.md exist on disk
- **THEN** the system SHALL parse the templates, replace {{.Name}}, {{.Role}}, {{.Backstory}}, and {{.OtherAgents}} fields with active values, and call the LLM with the generated prompt content

<!-- @trace
source: fix-agent-looping-and-feedback-sync
updated: 2026-06-19
code:
  - .autohand/skills/frontend-design
  - .agents/critical_reviewer/system_prompt.md
  - .agents/technical_architect/system_prompt.md
  - .agents/system_prompt.md
  - cmd/agent-office/main.go
  - pkg/workforce/interruption.go
  - .autohand/skills/frontend-design/SKILL.md
  - pkg/workforce/coordinator.go
  - .agents/lead_strategist/system_prompt.md
  - cmd/agent-office/gui/src/main.js
  - pkg/config/config.go
tests:
  - pkg/workforce/coordinator_test.go
  - pkg/config/config_test.go
  - pkg/workforce/interruption_test.go
-->

---
### Requirement: Autocomplete popup in guidance input
The companion GUI guidance input textarea SHALL support an autocomplete popup.
- Typing `@` MUST display a dropdown containing configured agent names and the `@everyone` tag.
- Typing `/` MUST display a dropdown containing local UI commands (`/clear`, `/help`, `/agents`, `/logs`) and dynamic skill options loaded from `.agents/skills/`.
- Pressing `Enter` or clicking an item in the autocomplete popup SHALL select the item and insert it into the textarea.
- Pressing `ArrowUp` or `ArrowDown` SHALL navigate through the list.

#### Scenario: Autocomplete popup displays on trigger character
- **WHEN** the user types "@" in the supervisor guidance input textarea
- **THEN** the autocomplete popup SHALL appear displaying the configured agents and @everyone
- **AND WHEN** the user selects "Lead Strategist" and presses Enter
- **THEN** the popup SHALL close and "@Lead Strategist " SHALL be inserted at the cursor position

#### Scenario: Slash command autocomplete list loaded dynamically
- **WHEN** the user types "/" in the supervisor guidance input textarea
- **THEN** the GUI SHALL fetch the available skills from the server
- **AND** the autocomplete popup SHALL display local UI commands and the fetched skill names as slash options


<!-- @trace
source: add-autocomplete-and-skills-routing
updated: 2026-06-19
code:
  - cmd/agent-office/gui/src/style.css
  - .autohand/skills/frontend-design/SKILL.md
  - .autohand/skills/frontend-design
  - cmd/agent-office/gui/src/main.js
  - cmd/agent-office/main.go
-->

---
### Requirement: Sequential routing for everyone tag
If the starting prompt or supervisor guidance contains the `@everyone` tag and does not contain any other explicit agent mentions:
- The coordinator SHALL execute a sequential loop where every configured agent is activated to speak exactly once in their defined order.
- This sequential routing SHALL persist across intermediate human interruptions until all configured agents have spoken.

#### Scenario: Everyone tag triggers sequential turn routing
- **WHEN** the user starts a run with the prompt "@everyone please introduce yourselves"
- **THEN** the coordinator SHALL route the first turn to the first configured agent
- **AND WHEN** that agent completes their turn and tags @User
- **AND** the user resumes with "please continue"
- **THEN** the coordinator SHALL route the next turn to the second configured agent who has not spoken yet


<!-- @trace
source: add-autocomplete-and-skills-routing
updated: 2026-06-19
code:
  - cmd/agent-office/gui/src/style.css
  - .autohand/skills/frontend-design/SKILL.md
  - .autohand/skills/frontend-design
  - cmd/agent-office/gui/src/main.js
  - cmd/agent-office/main.go
-->

---
### Requirement: Agent skill prompt injection
The workforce runtime SHALL support both static and dynamic skill prompt loading:
- **Static Skill loading**: If an agent configuration contains a skill name in its `skills` array, the system SHALL read the skill template from `.agents/skills/[skill_name]/SKILL.md` and append its instructions to the agent's system prompt for all turns.
- **Dynamic Skill loading**: If a routed message contains a skill command (e.g. `@AgentName /SkillName`), the system SHALL load the skill template from `.agents/skills/[skill_name]/SKILL.md` and dynamically append its instructions to the target agent's system prompt for that turn only.

#### Scenario: Dynamic skill prompt injection on tag
- **WHEN** the message history contains "@Technical Architect /frontend-design please style this button"
- **THEN** the system SHALL load the content of .agents/skills/frontend-design/SKILL.md
- **AND** append it to the system prompt of Technical Architect for the subsequent LLM call

<!-- @trace
source: add-autocomplete-and-skills-routing
updated: 2026-06-19
code:
  - cmd/agent-office/gui/src/style.css
  - .autohand/skills/frontend-design/SKILL.md
  - .autohand/skills/frontend-design
  - cmd/agent-office/gui/src/main.js
  - cmd/agent-office/main.go
-->