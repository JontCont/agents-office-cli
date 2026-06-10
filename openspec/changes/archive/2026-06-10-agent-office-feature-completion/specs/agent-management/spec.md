## ADDED Requirements

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
