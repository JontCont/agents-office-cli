## Context

`agent-office` is a CLI-first multi-agent workflow tool. The current server in `cmd/agent-office/main.go` uses a single `http.Handle("/", fileServer)` that intercepts all HTTP paths, making it impossible to add REST API routes alongside WebSocket and static file serving. The `models` command lists all four providers unconditionally, regardless of which provider the user has configured. There is no mechanism for creating agents outside of manually editing `agent-office.yaml`. Session token data is displayed in the GUI telemetry panel but never persisted to disk.

## Goals / Non-Goals

**Goals:**

- Make `agent-office models` show models only for the active provider (read from `.agent-office-token`)
- Fetch OpenRouter models dynamically from `GET https://openrouter.ai/api/v1/models` using the stored token
- Add `agent-office agent create` interactive CLI command that appends a new agent to `agent-office.yaml`
- Add REST endpoints `GET /api/agents`, `POST /api/agents`, `PUT /api/agents`, and `DELETE /api/agents` for the GUI
- Add a GUI Agents panel (tab-based layout) for listing, creating, editing, and deleting agents
- Fix Go struct `AgentConfig` in `pkg/config/config.go` with explicit `json` tags to prevent fields rendering as `undefined` in GUI
- Extend GUI telemetry with per-provider cost estimation (token * rate)
- Write session JSON log to `.agent-office-sessions/<YYYY-MM-DD>-<runID>.json` on run completion
- Restructure the HTTP server to use an explicit `http.ServeMux`

**Non-Goals:**

- Real OpenRouter streaming LLM calls (still mock workforce in this change)
- Persistent cost rate configuration (rates are hardcoded constants per provider)
- Authentication for the REST API (local-only server, no auth required)
- Pagination for the OpenRouter model list (show all returned models)

## Decisions

### Provider-aware models via ServeMux and REST

The HTTP server must serve three concerns on a single port: WebSocket (`/ws`), REST API (`/api/...`), and static files (`/`). Using `http.DefaultServeMux` with explicit pattern registration solves this without adding a router dependency. The existing `http.Handle("/", fileServer)` catches everything ? switching to a named `mux := http.NewServeMux()` and registering patterns in specificity order (`/api/`, `/ws`, `/`) gives correct routing.

Alternative considered: introducing `gorilla/mux` or `chi`. Rejected ? adds an external dependency for functionality the stdlib covers.

### OpenRouter model list via live API call

OpenRouter exposes `GET https://openrouter.ai/api/v1/models` (Bearer token, JSON response with `data[].id`). The `agent-office models` command reads the active provider from `.agent-office-token`, and if it is `openrouter`, performs an HTTP GET with the stored token and prints `data[].id` values. For other providers, the command prints a static curated list.

Alternative considered: caching the model list locally. Rejected ? unnecessary complexity; the list is small and the call is fast.

### Agent creation, editing, and deletion write directly to agent-office.yaml

The CLI (`agent create`) and GUI REST handlers (`POST /api/agents`, `PUT /api/agents`, `DELETE /api/agents`) read `agent-office.yaml`, modify the agents slice, and write the file back using `gopkg.in/yaml.v3` marshal.
- `POST /api/agents` appends a new agent.
- `PUT /api/agents` updates the agent matching the field `originalName`. If the name is changed, the agent is renamed in the list.
- `DELETE /api/agents?name=<name>` filters out the agent with the specified name from the list.
The REST handlers return JSON representations of the created/modified agents or error JSONs on failures.

### Go Config Agent Struct JSON Serialization

The `AgentConfig` struct in `pkg/config/config.go` is updated to include explicit `json` tags (e.g. `json:"name"`). Without these, the default Go JSON marshaler uses the capitalized field names (`Name`, `Role`, `Backstory`), causing the GUI frontend to show `undefined` as it expects lowercase keys.

### Session log format

On `StateCompleted` or `StateCancelled`, the coordinator callback writes a JSON file:

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

File path: `.agent-office-sessions/2026-06-10-run-12345.json`. Directory is created on first write.

### Cost estimation rates (hardcoded constants)

Rates per 1 000 tokens (blended input+output approximation):

| Provider | Rate (USD/1K tokens) |
|---|---|
| openai | 0.005 |
| gemini | 0.0005 |
| anthropic | 0.008 |
| openrouter | 0.002 |

These are rough estimates shown for awareness, not billing. Stored in `pkg/workforce/provider.go` as a `ProviderCostPer1KTokens` map.

## Implementation Contract

**CLI ? `agent-office models`**

- Reads `.agent-office-token` ? `default_provider`
- If `openrouter`: calls `GET https://openrouter.ai/api/v1/models` with `Authorization: Bearer <token>`, prints each `data[].id` one per line prefixed with `- `
- If `openai` / `gemini` / `anthropic`: prints the static list for that provider only
- If no token file or provider empty: prints error `"No provider configured. Run 'agent-office login' first."` and exits 1
- Verification: run `agent-office models` after `login --provider openrouter <token>` ? output contains OpenRouter model IDs, not OpenAI/Gemini/Anthropic models

**CLI ? `agent-office agent create`**

- Prompts: `Agent name:`, `Role:`, `Backstory:`
- Reads `agent-office.yaml`, appends new `AgentConfig`, writes file back
- Prints: `Agent '<name>' created successfully.`
- Error if `agent-office.yaml` missing: `"No configuration found. Run 'agent-office init' first."` exit 1
- Verification: run `agent-office agent create`, fill prompts, then run `agent-office agent list` ? new agent appears

**REST API - `GET /api/agents`**

- Returns: `{"agents": [{"name":"...","role":"...","backstory":"..."}]}`
- Error if yaml missing: HTTP 500 with `{"error": "configuration not found"}`

**REST API - `POST /api/agents`**

- Body: `{"name":"...","role":"...","backstory":"..."}`
- Validates: name and role required (non-empty); returns HTTP 400 with `{"error":"name and role are required"}` if missing
- Appends to yaml, returns HTTP 201 with the created agent as JSON
- Verification: POST to `/api/agents` - `agent-office agent list` shows new agent

**REST API - `PUT /api/agents`**

- Body: `{"originalName":"...","name":"...","role":"...","backstory":"..."}`
- Validates: `originalName`, `name` and `role` required (non-empty); returns HTTP 400 with `{"error":"originalName, name and role are required"}` if missing
- Updates the matching agent, returns HTTP 200 with the updated agent as JSON
- Verification: PUT to `/api/agents` - config file contains updated agent fields

**REST API - `DELETE /api/agents`**

- Method: `DELETE`, Query Param: `?name=<name>`
- Validates: `name` query parameter is required (non-empty); returns HTTP 400 if missing
- Deletes the matching agent, returns HTTP 200
- Verification: DELETE to `/api/agents?name=qa` - agent-office.yaml no longer lists that agent

**GUI - Agents tab**

- Tab bar added: "Run Thread" | "Agents"
- Agents tab: table listing name/role/backstory with Edit/Delete icons in an Actions column, plus "Add Agent" button
- Add Agent button opens the inline form in "Create" mode (title: "New Agent", button: "Create Agent")
- Clicking the Edit button on a row:
  - Fills the form inputs with that agent's current values
  - Stores `editingAgentName = name`
  - Switches form title to "Edit Agent" and submit button to "Save Changes"
  - Displays the form (slides in)
- Submitting the form:
  - If in Create mode, sends `POST /api/agents`
  - If in Edit mode, sends `PUT /api/agents` with `originalName = editingAgentName`
  - On success (200/210), closes the form, clears `editingAgentName`, and refreshes the table
- Clicking the Delete button on a row:
  - Prompts a confirmation dialog (e.g. `confirm("Are you sure you want to delete agent <name>?")`)
  - On confirmation, sends `DELETE /api/agents?name=<name>` and refreshes the table
- Verification: open GUI, manage agents (create, edit, delete) and verify changes persist and table refreshes instantly without page reload

**GUI ? Session Usage panel**

- Extends existing Telemetry card with two new metric boxes: `Est. Cost (USD)` and `Session Log`
- Cost updates on each `agent.speak` event: `cost = (totalTokens / 1000) * rateForProvider`
- Provider rate is injected into the page via a `<meta>` tag or a `/api/config` endpoint returning `{"provider":"openrouter","cost_per_1k":0.002}`
- On `state.change` ? `COMPLETED` or `CANCELLED`: fetch `/api/session/latest` to get the log file path and display it as a non-clickable label `Saved: .agent-office-sessions/...`
- Verification: run `agent-office gui`, watch run complete, check `.agent-office-sessions/` for JSON file, GUI shows estimated cost

**HTTP server restructure**

- `handleGUI()` creates `mux := http.NewServeMux()` and registers: `mux.Handle("/ws", wsUpgradeHandler)`, `mux.HandleFunc("/api/agents", agentsHandler)`, `mux.HandleFunc("/api/config", configHandler)`, `mux.Handle("/", http.FileServer(...))`
- `http.Serve(listener, mux)` replaces `http.Serve(listener, nil)`
- Scope boundary: WebSocket upgrade logic in `pkg/workforce/wsserver.go` is unchanged; only the mux registration in `main.go` changes

## Risks / Trade-offs

- [Risk] YAML round-trip via `gopkg.in/yaml.v3` drops comments in `agent-office.yaml` ? Mitigation: document this behavior; developer tool, acceptable trade-off.
- [Risk] OpenRouter API rate limit on `models` endpoint ? Mitigation: the call is user-initiated (not automatic), and OpenRouter's free tier has generous limits. No caching added in this change.
- [Risk] Cost estimates are inaccurate (mock token counts, blended rates) ? Mitigation: label the UI field clearly as "Est. Cost" not "Billed Cost".
