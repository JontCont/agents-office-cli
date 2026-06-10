## Why

The current `agent-office` CLI provides core infrastructure for multi-agent workforce management, but lacks three key capabilities that make it useful in practice: users cannot discover available AI models for their configured provider, cannot manage agents without editing YAML manually, and have no visibility into per-session token usage or cost after a run completes.

## What Changes

- `agent-office models` becomes provider-aware: lists models for the currently configured provider only. For OpenRouter, it fetches the live model list from the OpenRouter API. For OpenAI, Gemini, and Anthropic, it lists curated static model sets.
- A new `agent-office agent create` CLI command allows interactive creation of agents by prompting for name, role, and backstory, then appending the result to `agent-office.yaml`.
- The GUI gains an Agents management panel accessible via a tab or sidebar, backed by a new REST API (`GET /api/agents`, `POST /api/agents`) that reads and writes `agent-office.yaml`.
- The GUI gains a Session Usage panel showing real-time token accumulation and estimated cost (tokens times per-provider rate). At run completion, the session is written to `.agent-office-sessions/<date>-<runID>.json` for historical review.
- The Go HTTP server is restructured from a single wildcard handler to an explicit ServeMux that routes `/ws`, `/api/*`, and `/` (static files) independently.

## Capabilities

### New Capabilities

- `provider-model-listing`: Provider-aware model discovery — CLI fetches available models for the active provider (dynamic for OpenRouter, static curated lists for OpenAI/Gemini/Anthropic).
- `agent-management`: Create and list agents via both CLI (`agent create`) and GUI REST API (`GET/POST /api/agents`), persisting changes to `agent-office.yaml`.
- `session-usage-tracking`: Real-time token and cost telemetry in the GUI during a run, plus a persisted JSON session log written on run completion.

### Modified Capabilities

- `cli-auth-provider`: The `models` command behavior changes — from listing all providers to listing only the active provider's models.

## Impact

- Affected specs: `provider-model-listing` (new), `agent-management` (new), `session-usage-tracking` (new), `cli-auth-provider` (modified)
- Affected code:
  - New: `.agent-office-sessions/` (runtime output directory, not tracked in source)
  - Modified: `cmd/agent-office/main.go`
  - Modified: `pkg/workforce/provider.go`
  - Modified: `pkg/config/config.go`
  - Modified: `cmd/agent-office/gui/index.html`
  - Modified: `cmd/agent-office/gui/src/main.js`
  - Modified: `cmd/agent-office/gui/src/style.css`
