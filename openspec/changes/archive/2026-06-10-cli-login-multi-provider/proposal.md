## Why

Currently, the workforce runtime relies on hardcoded agent lists and lacks native integration with external LLM APIs like OpenRouter, Gemini, or OpenAI. Furthermore, developers need a secure and flexible way to configure their AI API credentials via the CLI (e.g. `agent-office login`) without exposing them inside the GUI dashboard or version-controlled workspace files, and support default models for each provider.

## What Changes

- Implement a CLI login flow (`agent-office login`) to securely capture AI API Tokens.
- Introduce an `AIProvider` enum to coordinate different connection parameters for `gemini`, `openai`, `anthropic`, and `openrouter` (OpenRouter API).
- Store configured API tokens as plain-text JSON locally in `.agent-office-token` in the workspace.
- Modify CLI runners (`agent-office run` and `agent-office gui`) to automatically read the credential token for the selected provider, failing if not logged in.
- Add default models for each provider (defaulting to `gpt-4o` for OpenAI) and a CLI subcommand (`agent-office models`) to list the supported models and their defaults.

## Capabilities

### New Capabilities

- `cli-auth-provider`: CLI login subcommand, local token storage, and default model configuration for multiple AI providers (Gemini, OpenAI, Anthropic, OpenRouter).

### Modified Capabilities

(none)

## Impact

- Affected specs: `cli-auth-provider`
- Affected code:
  - New:
    - `pkg/workforce/provider.go`
  - Modified:
    - `cmd/agent-office/main.go`
    - `pkg/config/config.go`
    - `pkg/workforce/types.go`
