## 1. CLI Login Subcommand and Provider Selection Menu

- [x] 1.1 Implement the **Interactive provider login** and **Decision 2: Interactive CLI Login Menu fallback** in `cmd/agent-office/main.go`. Behavior: Executing `agent-office login` without arguments prompts the user to select from a list of providers (1: Gemini, 2: OpenAI, 3: Anthropic, 4: OpenRouter) and prompts for a token. Verification: Run `agent-office login` manually and verify the interactive prompts and options.
- [x] 1.2 Implement the **Flag-based provider login** in `cmd/agent-office/main.go` and `pkg/workforce/provider.go`. Behavior: Running `agent-office login --provider <provider> <token>` configures the credential directly without prompts. Verification: Run `agent-office login --provider openrouter MY_TEST_TOKEN` and inspect the saved configuration.
- [x] 1.3 Implement the **Decision 1: Plain-text Local Token Configuration** credentials store in `pkg/workforce/provider.go`. Behavior: Credentials are saved/updated as plain-text JSON in `.agent-office-token`. Verification: Confirm `.agent-office-token` has correct JSON structure after running `agent-office login`.

## 2. Multi-Provider Validation and HTTP Routing

- [x] 2.1 Implement the **Token loading on execution** logic in `cmd/agent-office/main.go` and `pkg/workforce/provider.go`. Behavior: The CLI runner (`run` or `gui`) loads `.agent-office-token` and fails if the token for the active provider is missing. Verification: Run `agent-office gui` with a missing token and assert it exits with authentication error output.
- [x] 2.2 Implement the **Decision 3: Multi-Provider LLM Authentication Handling** client configuration in `pkg/workforce/provider.go` and update `pkg/config/config.go` with the `provider` configuration field. Behavior: The active provider token is loaded and applied to LLM client requests using correct endpoint (e.g., OpenRouter endpoint `https://openrouter.ai/api/v1` and Bearer authentication header). Verification: Run unit tests checking client initialization and provider mapping.

## 3. Default Model Configuration and Resolution

- [x] 3.1 Implement the **Default model configuration** and **Decision 4: Default Model Resolution and models Command** in `pkg/workforce/provider.go` and `cmd/agent-office/main.go`. Behavior: Executing `agent-office models` lists all providers and default models (with `gpt-4o` as the default). Verification: Run `agent-office models` and verify the terminal output.
- [x] 3.2 Update config loader and runners in `pkg/config/config.go` and `cmd/agent-office/main.go` to support default values. Behavior: If no provider is set in `agent-office.yaml`, default to `openai` (and `gpt-4o`), loading the corresponding token. Verification: Run `agent-office run` with no provider set and confirm it outputs errors for `openai` token missing (or runs if configured).
