# Agent Office CLI & Companion GUI

`agent-office` is a high-performance, local-first multi-agent orchestration runtime built in Go. It enables developers to coordinate agent workforces deterministically and monitor executions through a decoupled companion GUI dashboard communicating over local WebSockets.

---

[English](#agent-office-cli--companion-gui) | [繁體中文](#agent-office-cli--companion-gui-繁體中文)

---

## Features

- **Deterministic Turn Coordination**: Avoids unpredictable and costly LLM-driven turn selection by resolving the next active agent via explicit handoffs (`@agent` / `@name`), `@everyone` broadcasts, active project stage rules, or a fallback planner agent.
- **Safe Step-Boundary Interruption**: Implements a robust state-transition engine (`QUEUED`, `RUNNING`, `INTERRUPTING`, `INTERRUPTED`, `RESUMING`, `COMPLETED`, `FAILED`, `CANCELLED`). Execution halts safely before sending LLM prompts or initiating tool calls, guaranteeing state consistency and preventing corrupted sessions.
- **Dynamic Agent Skills System**: Load static skills from agent configuration or dynamically inject skills on-the-fly via `@AgentName /SkillName` tags. Skills are stored as markdown files in `.agents/skills/` and automatically appended to agent system prompts.
- **Flexible System Prompts**: Configure global instructions in `.agents/system_prompt.md` for all agents, or add agent-specific prompts in `.agents/{agent_name}/system_prompt.md` for role-specific behavior customization.
- **Embedded Web Dashboard**: Compiles all static web assets (HTML, CSS, JavaScript) directly into the Go executable via `go:embed`. Serves the companion app on port `8080` and upgrades client connections to WebSockets over a single network port.
- **Tauri Companion GUI Scaffolding**: Provides structured configurations inside the frontend directory, pre-configured to build a native cross-platform desktop application using Tauri.
- **Rich Agent Configuration**: Per-agent customization including color themes, avatar emojis, provider/model overrides, and individual API tokens for fine-grained control.
- **CLI Workspace Commands**: Full agent lifecycle management with `init`, `agent create`, `agent edit`, `agent delete`, and `agent list` commands for workspace configuration.

---

## Architecture

The Go backend serves as the single source of truth for the workspace, coordinating agent routing and checking step boundaries. The GUI is a detached monitor client that connects to the Go backend via WebSocket.

```mermaid
graph TD
    subgraph Go Backend Runtime (agent-office)
        CLI[Go CLI Entrypoint] --> Server[WebSocket Server :8080]
        CLI --> Config[Config Loader]
        CLI --> Coord[Deterministic Turn Coordinator]
        CLI --> Flow[Step-Boundary Interruption Engine]
    end

    subgraph Client Companion GUI
        Browser[System Browser / Tauri Window] <-->|WebSocket IPC| Server
        Browser --> Thread[Run Thread Viewer]
        Browser --> Telemetry[Metrics Dashboard]
        Browser --> Controls[Interruption / Resume Panel]
    end
```

---

## Project Structure

```
├── cmd/
│   └── agent-office/
│       ├── main.go               # Go entry point & CLI subcommands
│       └── gui/                  # Web Dashboard Frontend
│           ├── index.html        # Dashboard panel layout
│           ├── src/
│           │   ├── main.js       # Reconnect loops, event parsing, rendering
│           │   └── style.css     # Premium dark-mode glassmorphic styles
│           └── src-tauri/        # Tauri workspace Rust templates & configurations
├── pkg/
│   ├── config/
│   │   └── config.go             # YAML parser & agent configuration loader
│   └── workforce/
│       ├── types.go              # Shared RunState, Event, and Command definitions
│       ├── coordinator.go        # Deterministic turn-routing logic
│       ├── interruption.go       # Step-boundary checking & transition manager
│       └── server.go             # Upgrader & broadcast WebSocket hub
├── .agents/                      # Agent behavior customization (created manually)
│   ├── system_prompt.md          # Global system prompt for all agents
│   ├── skills/                   # Reusable skill modules
│   │   └── {skill_name}/
│   │       └── SKILL.md          # Skill instructions loaded on demand
│   └── {agent_name}/             # Per-agent customization
│       └── system_prompt.md      # Agent-specific prompt additions
├── go.mod                        # Go module manifest
└── agent-office.yaml             # Workspace config (generated on init)
```

---

## Getting Started

### Prerequisites
- [Go 1.22+](https://go.dev/dl/)

### Installation & Compilation
Clone the repository, download dependencies, and compile the CLI executable:
```powershell
# Clean and install dependencies
go mod tidy

# Build the executable
go build -o agent-office.exe ./cmd/agent-office
```

---

## CLI Commands

### 1. Workspace Initialization
Initialize a new workspace configuration in the current working directory. This creates a default `agent-office.yaml` file:
```powershell
.\agent-office.exe init
```

### 2. Agent Management

#### List Configured Agents
View all configured agents with their roles and settings:
```powershell
.\agent-office.exe agent list
```

#### Create a New Agent
Interactively create a new agent (prompts for name, role, and backstory):
```powershell
.\agent-office.exe agent create
```

#### Edit an Existing Agent
Interactively modify an agent's configuration:
```powershell
.\agent-office.exe agent edit
```

#### Delete an Agent
Remove an agent from the workspace configuration:
```powershell
.\agent-office.exe agent delete
```

### 3. Running the Workforce

#### Start the Runtime Server
Launch the multi-agent runtime without the GUI (headless mode):
```powershell
.\agent-office.exe run
```

#### Launch the GUI Dashboard Companion
Start the WebSocket backend, serve the dashboard assets, and automatically launch your default web browser:
```powershell
.\agent-office.exe gui
```
*Note: The GUI provides real-time monitoring, interactive controls for **Interrupt**, **Abort**, and **Resume** operations, and displays execution logs and telemetry.*

---

## Agent Customization

### System Prompts

Agent behavior can be customized through layered system prompts:

**Global Prompt** (`.agents/system_prompt.md`):
- Applied to all agents in the workspace
- Define shared collaboration protocols, tagging conventions, and response guidelines
- Example: Instructions for using `@AgentName` handoffs and `@User` for supervisor questions

**Agent-Specific Prompts** (`.agents/{agent_name}/system_prompt.md`):
- Override or extend the global prompt for individual agents
- Add role-specific instructions, expertise areas, or output formatting rules
- Example: `.agents/technical_architect/system_prompt.md` for architecture-specific guidelines

### Skills System

Skills are reusable instruction modules that can be loaded statically or dynamically:

**Static Skills** (configured in `agent-office.yaml`):
```yaml
agents:
  - name: Technical Architect
    role: System Designer
    skills:
      - frontend-design
      - api-design
```

**Dynamic Skills** (injected via message tags):
```
@TechnicalArchitect /frontend-design please review this component
```

Skills are stored in `.agents/skills/{skill_name}/SKILL.md` and automatically appended to the agent's system prompt for that turn.

**Creating a Skill**:
1. Create directory: `.agents/skills/my-skill/`
2. Add instructions: `.agents/skills/my-skill/SKILL.md`
3. Reference in config or use `@AgentName /my-skill` in messages

---

## Configuration Reference

The `agent-office.yaml` file supports the following fields:

```yaml
version: "1.0"
username: YourName              # Supervisor name for @username mentions
provider: anthropic             # Default LLM provider (anthropic, openrouter)
model: claude-haiku-4-5         # Default model for all agents
agents:
  - name: Agent Name
    role: Agent Role
    backstory: ""               # Optional agent backstory
    skills: []                  # Static skills to load
    tools: []                   # Reserved for future tool integration
    hooks: []                   # Reserved for future webhook integration
    color: '#6366f1'            # GUI avatar color (hex)
    avatar: "📋"                # Emoji avatar for GUI
    provider: anthropic         # Override global provider
    model: claude-haiku-4-5     # Override global model
    token: sk-ant-...           # Agent-specific API token
```

---

## Running Tests

All core coordination logic, state transitions, and IPC packet serializations are guarded by unit tests:

```powershell
# Run the complete test suite
go test ./... -v
```
- `coordinator_test.go`: Asserts explicit keyword routing, stage routing, and planner fallbacks.
- `interruption_test.go`: Verifies thread-safe boundary blocking, resume injections, and aborts.
- `event_test.go`: Tests event serialization compliance.

---

# Agent Office CLI & Companion GUI (繁體中文)

`agent-office` 是一個基於 Go 語言開發的高性能、在地優先（Local-first）多智能體（Multi-agent）編排運行時。它使開發人員能夠以確定性的方式協調智能體工作流，並透過本地 WebSocket 通訊的獨立 GUI 儀表板監控執行狀態。

---

## 特性 (Features)

- **確定性輪替協調 (Deterministic Turn Coordination)**：避免因 LLM 驅動輪替選擇所帶來的不確定性與高昂成本，透過明確的移交（`@agent` / `@name`）、`@everyone` 廣播、當前項目階段規則或後備規劃器智能體（Fallback planner agent）來解析下一個活動智能體。
- **安全的步驟邊界中斷 (Safe Step-Boundary Interruption)**：實現了強健的狀態轉移引擎（`QUEUED`、`RUNNING`、`INTERRUPTING`、`INTERRUPTED`、`RESUMING`、`COMPLETED`、`FAILED`、`CANCELLED`）。在發送 LLM 提示詞或啟動工具調用之前安全暫停執行，保證狀態一致性並防止會話損壞。
- **動態智能體技能系統 (Dynamic Agent Skills System)**：從智能體配置載入靜態技能，或透過 `@AgentName /SkillName` 標籤動態注入技能。技能以 Markdown 檔案儲存於 `.agents/skills/` 中，並自動附加至智能體的系統提示詞。
- **彈性系統提示詞 (Flexible System Prompts)**：在 `.agents/system_prompt.md` 中配置所有智能體的全域指令，或在 `.agents/{agent_name}/system_prompt.md` 中為特定智能體新增角色專屬的行為客製化。
- **內嵌網頁儀表板 (Embedded Web Dashboard)**：透過 `go:embed` 將所有靜態網頁資源（HTML、CSS、JavaScript）直接編譯至 Go 可執行檔中。在連接埠 `8080` 上提供隨附的應用程式，並透過單一網路連接埠將用戶端連線升級為 WebSocket。
- **Tauri 隨附 GUI 腳手架 (Tauri Companion GUI Scaffolding)**：在前端目錄中提供結構化配置，預先配置為使用 Tauri 建置原生跨平台桌面應用程式。
- **豐富的智能體配置 (Rich Agent Configuration)**：每個智能體可客製化顏色主題、頭像表情符號、提供者/模型覆寫，以及個別 API token，實現精細控制。
- **CLI 工作區指令 (CLI Workspace Commands)**：完整的智能體生命週期管理，包含 `init`、`agent create`、`agent edit`、`agent delete` 和 `agent list` 指令，用於工作區配置。

---

## 架構 (Architecture)

Go 後端作為工作區的單一事實來源，協調智能體路由並檢查步驟邊界。GUI 是一個獨立的監控用戶端，透過 WebSocket 連接到 Go 後端。

```mermaid
graph TD
    subgraph Go 後端運行時 (agent-office)
        CLI[Go CLI 入口點] --> Server[WebSocket 伺服器 :8080]
        CLI --> Config[配置載入器]
        CLI --> Coord[確定性輪替協調器]
        CLI --> Flow[步驟邊界中斷引擎]
    end

    subgraph 用戶端隨附 GUI (Client Companion GUI)
        Browser[系統瀏覽器 / Tauri 視窗] <-->|WebSocket IPC| Server
        Browser --> Thread[執行執行緒檢視器]
        Browser --> Telemetry[指標儀表板]
        Browser --> Controls[中斷 / 恢復控制面板]
    end
```

---

## 專案結構 (Project Structure)

```
├── cmd/
│   └── agent-office/
│       ├── main.go               # Go 入口點與 CLI 子指令
│       └── gui/                  # 網頁儀表板前端
│           ├── index.html        # 儀表板面板版面配置
│           ├── src/
│           │   ├── main.js       # 重連循環、事件解析、渲染
│           │   └── style.css     # 高級暗色模式磨砂玻璃風格 (Glassmorphic)
│           └── src-tauri/        # Tauri 工作區 Rust 範本與配置
├── pkg/
│   ├── config/
│   │   └── config.go             # YAML 解析器與智能體配置載入器
│   └── workforce/
│       ├── types.go              # 共用的 RunState、Event 和 Command 定義
│       ├── coordinator.go        # 確定性輪替路由邏輯
│       ├── interruption.go       # 步驟邊界檢查與轉移管理器
│       └── server.go             # 升級器與廣播 WebSocket 集線器 (Hub)
├── .agents/                      # 智能體行為客製化 (手動建立)
│   ├── system_prompt.md          # 所有智能體的全域系統提示詞
│   ├── skills/                   # 可重複使用的技能模組
│   │   └── {skill_name}/
│   │       └── SKILL.md          # 按需載入的技能指令
│   └── {agent_name}/             # 每個智能體的客製化
│       └── system_prompt.md      # 智能體專屬的提示詞附加內容
├── go.mod                        # Go 模組清單
└── agent-office.yaml             # 工作區配置 (於 init 時生成)
```

---

## 入門指南 (Getting Started)

### 前提條件
- [Go 1.22+](https://go.dev/dl/)

### 安裝與編譯
複製儲存庫，下載相依套件，並編譯 CLI 可執行檔：
```powershell
# 清理並安裝相依套件
go mod tidy

# 建置可執行檔
go build -o agent-office.exe ./cmd/agent-office
```

---

## CLI 指令 (CLI Commands)

### 1. 工作區初始化
在當前工作目錄中初始化新的工作區配置。這將建立一個預設的 `agent-office.yaml` 檔案：
```powershell
.\agent-office.exe init
```

### 2. 智能體管理

#### 列出已配置的智能體
檢視所有已配置的智能體及其角色和設定：
```powershell
.\agent-office.exe agent list
```

#### 建立新智能體
互動式建立新智能體（提示輸入名稱、角色和背景故事）：
```powershell
.\agent-office.exe agent create
```

#### 編輯現有智能體
互動式修改智能體的配置：
```powershell
.\agent-office.exe agent edit
```

#### 刪除智能體
從工作區配置中移除智能體：
```powershell
.\agent-office.exe agent delete
```

### 3. 執行工作流

#### 啟動運行時伺服器
啟動多智能體運行時（無 GUI 的 headless 模式）：
```powershell
.\agent-office.exe run
```

#### 啟動 GUI 儀表板隨附程式
啟動 WebSocket 後端，提供儀表板資源，並自動啟動您的預設網頁瀏覽器：
```powershell
.\agent-office.exe gui
```
*說明：GUI 提供即時監控、**中斷（Interrupt）**、**中止（Abort）** 和 **恢復（Resume）** 操作的互動式控制，並顯示執行日誌和遙測數據。*

---

## 智能體客製化 (Agent Customization)

### 系統提示詞 (System Prompts)

智能體行為可以透過分層的系統提示詞進行客製化：

**全域提示詞** (`.agents/system_prompt.md`):
- 適用於工作區中的所有智能體
- 定義共享的協作協定、標籤約定和回應準則
- 範例：使用 `@AgentName` 移交和 `@User` 向監督者提問的指令

**智能體專屬提示詞** (`.agents/{agent_name}/system_prompt.md`):
- 覆寫或擴展個別智能體的全域提示詞
- 新增角色特定的指令、專業領域或輸出格式規則
- 範例：`.agents/technical_architect/system_prompt.md` 用於架構特定的指導方針

### 技能系統 (Skills System)

技能是可重複使用的指令模組，可以靜態或動態載入：

**靜態技能** (在 `agent-office.yaml` 中配置):
```yaml
agents:
  - name: Technical Architect
    role: System Designer
    skills:
      - frontend-design
      - api-design
```

**動態技能** (透過訊息標籤注入):
```
@TechnicalArchitect /frontend-design please review this component
```

技能儲存於 `.agents/skills/{skill_name}/SKILL.md`，並在該輪次自動附加至智能體的系統提示詞。

**建立技能**:
1. 建立目錄：`.agents/skills/my-skill/`
2. 新增指令：`.agents/skills/my-skill/SKILL.md`
3. 在配置中引用或在訊息中使用 `@AgentName /my-skill`

---

## 配置參考 (Configuration Reference)

`agent-office.yaml` 檔案支援以下欄位：

```yaml
version: "1.0"
username: YourName              # 監督者名稱，用於 @username 提及
provider: anthropic             # 預設 LLM 提供者 (anthropic, openrouter)
model: claude-haiku-4-5         # 所有智能體的預設模型
agents:
  - name: Agent Name
    role: Agent Role
    backstory: ""               # 可選的智能體背景故事
    skills: []                  # 要載入的靜態技能
    tools: []                   # 保留供未來工具整合使用
    hooks: []                   # 保留供未來 webhook 整合使用
    color: '#6366f1'            # GUI 頭像顏色 (十六進位)
    avatar: "📋"                # GUI 的表情符號頭像
    provider: anthropic         # 覆寫全域提供者
    model: claude-haiku-4-5     # 覆寫全域模型
    token: sk-ant-...           # 智能體專屬的 API token
```

---

## 執行測試 (Running Tests)

所有核心協調邏輯、狀態轉移和 IPC 封包序列化均由單元測試保護：

```powershell
# 執行完整測試套件
go test ./... -v
```
- `coordinator_test.go`：斷言明確關鍵字路由、階段路由和規劃器後備。
- `interruption_test.go`：驗證安全執行緒邊界阻塞、恢復注入和中止。
- `event_test.go`：測試事件序列化合規性。

