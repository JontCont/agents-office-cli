# session-usage-tracking Specification

## Purpose

TBD - created by archiving change 'agent-office-feature-completion'. Update Purpose after archive.

## Requirements

### Requirement: Real-time session cost estimation in GUI
The system SHALL extend the GUI telemetry panel to display an estimated cost field that updates in real-time as `agent.speak` events are received. The cost SHALL be calculated as `(totalTokens / 1000) * ratePerProvider`. The rate per provider SHALL be obtained by fetching `GET /api/config` which returns the active provider name and its cost-per-1K-tokens rate.

#### Scenario: Cost estimation updates on each agent event
- **WHEN** the GUI receives an `agent.speak` event with a `tokens` metadata field
- **THEN** the telemetry panel SHALL add those tokens to `totalTokens` and recalculate `Est. Cost (USD)` as `(totalTokens / 1000) * ratePerProvider`
- **AND** display the result formatted to 4 decimal places (e.g., `$0.0014`)

##### Example: Cost calculation by provider

| Provider | Rate (USD/1K tokens) | Total Tokens | Expected Est. Cost |
| --- | --- | --- | --- |
| openai | 0.005 | 695 | $0.0035 |
| openrouter | 0.002 | 695 | $0.0014 |
| gemini | 0.0005 | 695 | $0.0003 |
| anthropic | 0.008 | 695 | $0.0056 |


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
### Requirement: Session log persisted on run completion
On run completion (state transitions to `COMPLETED` or `CANCELLED`), the system SHALL write a JSON session log file to `.agent-office-sessions/<YYYY-MM-DD>-<runID>.json`. The directory SHALL be created if it does not exist. The JSON file SHALL contain: `run_id`, `provider`, `model`, `started_at` (Unix timestamp), `ended_at` (Unix timestamp), `total_tokens`, `estimated_cost_usd`, and `steps`.

#### Scenario: Session log written on COMPLETED
- **WHEN** the run transitions to state `COMPLETED`
- **THEN** the system SHALL write `.agent-office-sessions/<date>-<runID>.json` with correct fields
- **AND** the file SHALL be valid JSON parseable without error

#### Scenario: Session log written on CANCELLED
- **WHEN** the run transitions to state `CANCELLED`
- **THEN** the system SHALL write the session log with `ended_at` set to the cancellation timestamp

#### Scenario: Session log directory auto-created
- **WHEN** `.agent-office-sessions/` does not exist and a run completes
- **THEN** the system SHALL create the directory and write the log file without error

##### Example: Session log JSON shape
```json
{
  "run_id": "run-12345",
  "provider": "openrouter",
  "model": "google/gemini-flash-1.5",
  "started_at": 1718000000,
  "ended_at": 1718000060,
  "total_tokens": 695,
  "estimated_cost_usd": 0.0014,
  "steps": 4
}
```


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
### Requirement: Session log path displayed in GUI after completion
After a run completes, the GUI SHALL display the path of the saved session log file in the telemetry panel. The system SHALL also load and render the full session log content in the dedicated Logs tab, replacing the modal popup log viewer.

#### Scenario: Session log path shown in GUI
- **WHEN** the GUI receives a state.change event with content COMPLETED or CANCELLED
- **THEN** the GUI SHALL fetch GET /api/session/latest and display the returned log file path in the telemetry panel
- **AND** when the user clicks the Logs tab, the GUI SHALL fetch GET /api/session/latest/content and display the log JSON in the tab's content panel

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