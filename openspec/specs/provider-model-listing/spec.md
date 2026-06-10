# provider-model-listing Specification

## Purpose

TBD - created by archiving change 'agent-office-feature-completion'. Update Purpose after archive.

## Requirements

### Requirement: Provider-aware model listing
The system SHALL make `agent-office models` display models only for the currently configured provider. The active provider is determined by reading the `default_provider` field from `.agent-office-token`. If no token file exists or the provider is empty, the command MUST exit with code 1 and print an error instructing the user to run `agent-office login` first.

#### Scenario: Models listed for OpenRouter provider
- **WHEN** the user runs `agent-office models` and the active provider is `openrouter`
- **THEN** the system SHALL call `GET https://openrouter.ai/api/v1/models` with `Authorization: Bearer <token>`
- **AND** print each model ID from the `data[].id` array, one per line, prefixed with `- `
- **AND** SHALL NOT print models from OpenAI, Gemini, or Anthropic

#### Scenario: Models listed for static provider
- **WHEN** the user runs `agent-office models` and the active provider is `openai`, `gemini`, or `anthropic`
- **THEN** the system SHALL print only the curated static model list for that provider
- **AND** SHALL NOT print models from other providers

##### Example: Provider isolation

| Active Provider | Expected Output Contains | Must NOT Contain |
| --- | --- | --- |
| openai | `gpt-4o`, `gpt-4-turbo`, `gpt-3.5-turbo` | gemini, claude, openrouter |
| gemini | `gemini-1.5-pro`, `gemini-1.5-flash` | gpt, claude, openrouter |
| anthropic | `claude-3-5-sonnet`, `claude-3-opus` | gpt, gemini, openrouter |
| openrouter | live list from API | (static models from other providers) |

#### Scenario: No provider configured
- **WHEN** the user runs `agent-office models` and no `.agent-office-token` exists or `default_provider` is empty
- **THEN** the system MUST print `"No provider configured. Run 'agent-office login' first."` and exit with code 1

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