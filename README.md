# Agent Office CLI & Companion GUI

`agent-office` is a high-performance, local-first multi-agent orchestration runtime built in Go. It enables developers to coordinate agent workforces deterministically and monitor executions through a decoupled companion GUI dashboard communicating over local WebSockets.

---

[English](#agent-office-cli--companion-gui) | [繁體中文](#agent-office-cli--companion-gui-繁體中文)

---

## Features

- **Deterministic Turn Coordination**: Avoids unpredictable and costly LLM-driven turn selection by resolving the next active agent via explicit handoffs (`@agent` / `@name`), active project stage rules, or a fallback planner agent.
- **Safe Step-Boundary Interruption**: Implements a robust state-transition engine (`QUEUED`, `RUNNING`, `INTERRUPTING`, `INTERRUPTED`, `RESUMING`, `COMPLETED`, `FAILED`, `CANCELLED`). Execution halts safely before sending LLM prompts or initiating tool calls, guaranteeing state consistency and preventing corrupted sessions.
- **Embedded Web Dashboard**: Compiles all static web assets (HTML, CSS, JavaScript) directly into the Go executable via `go:embed`. Serves the companion app on port `8080` and upgrades client connections to WebSockets over a single network port.
- **Tauri Companion GUI Scaffolding**: Provides structured configurations inside the frontend directory, pre-configured to build a native cross-platform desktop application using Tauri.
- **CLI Workspace Commands**: Quick initialization commands (`init`) to scaffold default configurations and list (`agent list`) the defined agent capabilities.

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

### 2. Listing Configured Agents
Verify the active configuration and view the list of configured agents and roles:
```powershell
.\agent-office.exe agent list
```

### 3. Launching the GUI Dashboard Companion
Starts the WebSocket backend, serves the dashboard assets, and automatically launches your default web browser to view the real-time simulation:
```powershell
.\agent-office.exe gui
```
*Note: If you run `gui`, the console will start streaming mock agent discussion thread turns. You can use the buttons on the dashboard to **Interrupt**, **Abort**, or **Resume** (with feedback message) the active execution.*

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

- **確定性輪替協調 (Deterministic Turn Coordination)**：避免因 LLM 驅動輪替選擇所帶來的不確定性與高昂成本，透過明確的移交（`@agent` / `@name`）、當前項目階段規則或後備規劃器智能體（Fallback planner agent）來解析下一個活動智能體。
- **安全的步驟邊界中斷 (Safe Step-Boundary Interruption)**：實現了強健的狀態轉移引擎（`QUEUED`、`RUNNING`、`INTERRUPTING`、`INTERRUPTED`、`RESUMING`、`COMPLETED`、`FAILED`、`CANCELLED`）。在發送 LLM 提示詞或啟動工具調用之前安全暫停執行，保證狀態一致性並防止會話損壞。
- **內嵌網頁儀表板 (Embedded Web Dashboard)**：透過 `go:embed` 將所有靜態網頁資源（HTML、CSS、JavaScript）直接編譯至 Go 可執行檔中。在連接埠 `8080` 上提供隨附的應用程式，並透過單一網路連接埠將用戶端連線升級為 WebSocket。
- **Tauri 隨附 GUI 腳手架 (Tauri Companion GUI Scaffolding)**：在前端目錄中提供結構化配置，預先配置為使用 Tauri 建置原生跨平台桌面應用程式。
- **CLI 工作區指令 (CLI Workspace Commands)**：提供快速初始化指令（`init`）以建置預設配置，並能列出（`agent list`）已定義的智能體能力。

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

### 2. 列出已配置的智能體
驗證活動配置並檢視已配置的智能體和角色列表：
```powershell
.\agent-office.exe agent list
```

### 3. 啟動 GUI 儀表板隨附程式
啟動 WebSocket 後端，提供儀表板資源，並自動啟動您的預設網頁瀏覽器以檢視即時模擬：
```powershell
.\agent-office.exe gui
```
*說明：如果您執行 `gui`，主控台將開始串流模擬智能體討論執行緒輪替。您可以使用儀表板上的按鈕來對活動執行進行**中斷（Interrupt）**、**中止（Abort）**或**恢復（Resume）**（包含反饋訊息）。*

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

