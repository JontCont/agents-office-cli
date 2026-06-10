## Context

Currently, the workforce runtime relies on hardcoded configurations and has no authentication mechanism to communicate with external APIs (like OpenRouter). We need a CLI-first, multi-provider credentials loader to store developer API keys securely on their local disk without committing them to the repository or displaying them in the companion GUI dashboard.

## Goals / Non-Goals

**Goals:**
- Implement a CLI login flow to save API keys to a local `.agent-office-token` configuration file in JSON format.
- Support `gemini`, `openai`, `anthropic`, and `openrouter` as distinct AI providers using an `AIProvider` enum.
- Ensure CLI execution checks for the active provider's token and fails early if not authenticated.
- Establish default model names for each provider, using `gpt-4o` (OpenAI) as the global fallback.

**Non-Goals:**
- No graphical login form in the companion GUI; the GUI remains a read-only management panel.
- No OS keychain encryption or platform-specific keyring integration for this phase; the token file is plain-text JSON.
- No automatic LLM API token verification during registration; token validity is tested only when running workflows.

## Decisions

### Decision 1: Plain-text Local Token Configuration
- **Approach**: The credentials are saved in `.agent-office-token` (in the workspace or user home directory) as plain-text JSON.
- **Alternatives**: Cryptographic encryption (DPAPI or AES-256). Rejected for this phase to maximize simplicity, debuggability, and direct user editability.
- **Rationale**: Keeps implementation lightweight and transparent. Developers can inspect, edit, or copy their keys directly.

### Decision 2: Interactive CLI Login Menu fallback
- **Approach**: Running `agent-office login` without arguments prompts the user interactively in the console to choose a provider (1 to 4) and type their key. Running with `--provider <name> <token>` configures the credential directly.
- **Alternatives**: Flag-only input. Rejected because interactive prompts provide a self-guiding user experience.
- **Rationale**: Interactive terminal menus allow quick credential setups without looking up documentation.

### Decision 3: Multi-Provider LLM Authentication Handling
- **Approach**: Introduce an `AIProvider` enum. The CLI runner reads the configured provider, fetches the matching token from `.agent-office-token`, and applies provider-specific headers (e.g. Bearer authorization and API base URL for OpenRouter).
- **Alternatives**: Relying solely on standard environment variables (like `GEMINI_API_KEY`). Rejected because we need to support concurrent configurations for multiple providers.
- **Rationale**: Allows developers to switch between Gemini, OpenAI, Claude, or OpenRouter in their workspace configuration file, while the backend takes care of routing requests correctly.

### Decision 4: Default Model Resolution and models Command
- **Approach**: Define a default model mapping inside Go. If no model is configured in `agent-office.yaml`, fallback to the provider's default model (Gemini: `gemini-1.5-pro`, OpenAI: `gpt-4o`, Anthropic: `claude-3-5-sonnet`, OpenRouter: `google/gemini-flash-1.5`). If no provider is specified at all, default to `openai` with `gpt-4o`. Implement `agent-office models` subcommand to list these defaults.
- **Alternatives**: Let LLM clients fail if no model is explicitly passed. Rejected because providing default settings ensures an out-of-the-box working experience.
- **Rationale**: Selecting `gpt-4o` as the default OpenAI model and providing defaults for other services speeds up workspace bootstrapping.

## Implementation Contract

#### CLI Command Contract
- Command: `agent-office login`
- Flags: `--provider <name>` (optional), takes an argument `<token>` (optional).
- Interactive flow: If no arguments/flags are passed, it displays:
  ```
  Select AI Provider:
    1) Gemini
    2) OpenAI
    3) Anthropic
    4) OpenRouter
  Enter selection (1-4):
  ```
  And then prompts for the Token securely.
- Result: Saves/updates the `.agent-office-token` JSON file:
  ```json
  {
    "tokens": {
      "openrouter": "SK_OR_123"
    }
  }
  ```

- Command: `agent-office models`
- Behavior: Lists all supported providers and their default models:
  ```
  Supported Models:
  - openai: gpt-4o (Default)
  - gemini: gemini-1.5-pro
  - anthropic: claude-3-5-sonnet
  - openrouter: google/gemini-flash-1.5
  ```

#### Configuration Contract
- Add `provider` string field in `agent-office.yaml`. Values: `gemini`, `openai`, `anthropic`, `openrouter`.
- Add `model` string field in `agent-office.yaml` (optional). Overrides the default model name.

#### Execution Contract
- When `agent-office run` or `agent-office gui` starts:
  - It loads the workspace `agent-office.yaml` configuration to find the active provider.
  - If the provider is not set, default to `openai`.
  - It loads `.agent-office-token`.
  - If the token for the active provider is missing, it exits with exit code 1 and prints: `Error: Not logged in for provider '<active-provider>'. Please run 'agent-office login' first.`
  - If the token exists, it initializes the corresponding LLM API client. For `openrouter`, it sets the authorization Bearer header and base URL to `https://openrouter.ai/api/v1`.

## Risks / Trade-offs

- **[Risk] Plain-text exposure** → [Mitigation] Developers must secure their local directories. We recommend adding `.agent-office-token` to global `.gitignore` patterns if saved inside the workspace.
