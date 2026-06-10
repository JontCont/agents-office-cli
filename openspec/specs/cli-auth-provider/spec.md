## ADDED Requirements

### Requirement: Interactive provider login
The system SHALL provide a CLI command `agent-office login` that configures credentials for AI providers. When run without arguments, the system SHALL display a numbered interactive terminal list of supported providers: 1) Gemini, 2) OpenAI, 3) Anthropic, 4) OpenRouter, and prompt the user to input the token for the selected provider.

#### Scenario: Interactive OpenRouter login
- **WHEN** the user executes `agent-office login` without arguments
- **THEN** the system SHALL display the menu of providers
- **AND** the user SHALL select option 4 (OpenRouter) and input token `SK_OR_123`
- **AND** the system SHALL write `openrouter: SK_OR_123` into `.agent-office-token`.

##### Example: Selection list
| User Input Menu | User Input Token | Expected Token Configuration |
| --- | --- | --- |
| 1 | KEY_GEM_999 | gemini: KEY_GEM_999 |
| 4 | KEY_OR_888 | openrouter: KEY_OR_888 |


<!-- @trace
source: cli-login-multi-provider
updated: 2026-06-10
code:
  - README.md
  - agent-office.exe
  - pkg/workforce/provider.go
  - .agent-office-token
  - cmd/agent-office/main.go
  - pkg/config/config.go
  - agent-office.yaml
tests:
  - pkg/workforce/provider_test.go
-->


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

### Requirement: Flag-based provider login
The system SHALL support flag-based credentials setup via `agent-office login --provider <provider> <token>` where `<provider>` is one of `gemini`, `openai`, `anthropic`, or `openrouter`.

#### Scenario: Login with OpenRouter provider flag
- **WHEN** the user runs `agent-office login --provider openrouter SK_OR_555`
- **THEN** the system SHALL write `openrouter: SK_OR_555` into `.agent-office-token` without displaying interactive prompts.


<!-- @trace
source: cli-login-multi-provider
updated: 2026-06-10
code:
  - README.md
  - agent-office.exe
  - pkg/workforce/provider.go
  - .agent-office-token
  - cmd/agent-office/main.go
  - pkg/config/config.go
  - agent-office.yaml
tests:
  - pkg/workforce/provider_test.go
-->


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

### Requirement: Token loading on execution
Before executing any multi-agent workforce loop, the system MUST read `.agent-office-token`. If the token for the active provider is missing, the execution MUST terminate with an authentication error.

#### Scenario: Running without logged in credentials
- **GIVEN** `.agent-office-token` does not contain credentials for `openrouter`
- **AND** the active workspace provider is set to `openrouter`
- **WHEN** the user executes `agent-office run` or `agent-office gui`
- **THEN** the execution MUST fail and output an error message instructing the user to login first.


<!-- @trace
source: cli-login-multi-provider
updated: 2026-06-10
code:
  - README.md
  - agent-office.exe
  - pkg/workforce/provider.go
  - .agent-office-token
  - cmd/agent-office/main.go
  - pkg/config/config.go
  - agent-office.yaml
tests:
  - pkg/workforce/provider_test.go
-->


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

### Requirement: Default model configuration
The system SHALL define default model names for each provider: `gpt-4o` for `openai`, `gemini-1.5-pro` for `gemini`, `claude-3-5-sonnet` for `anthropic`, and `google/gemini-flash-1.5` for `openrouter`. If no provider is specified in the configuration, the system SHALL default to `openai` (and thus `gpt-4o`). The system SHALL support a CLI subcommand `agent-office models` which lists all supported providers and their default models.

#### Scenario: Listing models via CLI
- **WHEN** the user runs `agent-office models`
- **THEN** the system SHALL output the list of supported providers and their default models.

##### Example: Models list output
```
Supported Models:
- openai: gpt-4o (Default)
- gemini: gemini-1.5-pro
- anthropic: claude-3-5-sonnet
- openrouter: google/gemini-flash-1.5
```

## Requirements

<!-- @trace
source: cli-login-multi-provider
updated: 2026-06-10
code:
  - README.md
  - agent-office.exe
  - pkg/workforce/provider.go
  - .agent-office-token
  - cmd/agent-office/main.go
  - pkg/config/config.go
  - agent-office.yaml
tests:
  - pkg/workforce/provider_test.go
-->


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

### Requirement: Interactive provider login
The system SHALL provide a CLI command `agent-office login` that configures credentials for AI providers. When run without arguments, the system SHALL display a numbered interactive terminal list of supported providers: 1) Gemini, 2) OpenAI, 3) Anthropic, 4) OpenRouter, and prompt the user to input the token for the selected provider.

#### Scenario: Interactive OpenRouter login
- **WHEN** the user executes `agent-office login` without arguments
- **THEN** the system SHALL display the menu of providers
- **AND** the user SHALL select option 4 (OpenRouter) and input token `SK_OR_123`
- **AND** the system SHALL write `openrouter: SK_OR_123` into `.agent-office-token`.

##### Example: Selection list
| User Input Menu | User Input Token | Expected Token Configuration |
| --- | --- | --- |
| 1 | KEY_GEM_999 | gemini: KEY_GEM_999 |
| 4 | KEY_OR_888 | openrouter: KEY_OR_888 |

---
### Requirement: Flag-based provider login
The system SHALL support flag-based credentials setup via `agent-office login --provider <provider> <token>` where `<provider>` is one of `gemini`, `openai`, `anthropic`, or `openrouter`.

#### Scenario: Login with OpenRouter provider flag
- **WHEN** the user runs `agent-office login --provider openrouter SK_OR_555`
- **THEN** the system SHALL write `openrouter: SK_OR_555` into `.agent-office-token` without displaying interactive prompts.

---
### Requirement: Token loading on execution
Before executing any multi-agent workforce loop, the system MUST read `.agent-office-token`. If the token for the active provider is missing, the execution MUST terminate with an authentication error.

#### Scenario: Running without logged in credentials
- **GIVEN** `.agent-office-token` does not contain credentials for `openrouter`
- **AND** the active workspace provider is set to `openrouter`
- **WHEN** the user executes `agent-office run` or `agent-office gui`
- **THEN** the execution MUST fail and output an error message instructing the user to login first.

---
### Requirement: Default model configuration
The system SHALL define default model names for each provider: `gpt-4o` for `openai`, `gemini-1.5-pro` for `gemini`, `claude-3-5-sonnet` for `anthropic`, and `google/gemini-flash-1.5` for `openrouter`. If no provider is specified in the configuration, the system SHALL default to `openai` (and thus `gpt-4o`). The system SHALL support a CLI subcommand `agent-office models` which lists all supported providers and their default models when a provider is configured.

#### Scenario: Listing models via CLI shows only active provider
- **WHEN** the user runs `agent-office models` with an active provider configured
- **THEN** the system SHALL output only the models available for that provider, not models for other providers.

##### Example: Models list output (active provider: openrouter)
The output SHALL contain only OpenRouter model IDs fetched from the live API, not gpt-4o, gemini-1.5-pro, or claude-3-5-sonnet.