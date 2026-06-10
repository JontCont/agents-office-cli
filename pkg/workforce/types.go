package workforce

type Agent struct {
	Name      string   `yaml:"name" json:"name"`
	Role      string   `yaml:"role" json:"role"`
	Backstory string   `yaml:"backstory" json:"backstory"`
	Skills    []string `yaml:"skills" json:"skills"`
	Tools     []string `yaml:"tools" json:"tools"`
	Hooks     []string `yaml:"hooks" json:"hooks"`
	Color     string   `yaml:"color" json:"color"`
	Avatar    string   `yaml:"avatar" json:"avatar"`
	Provider  string   `yaml:"provider" json:"provider"`
	Model     string   `yaml:"model" json:"model"`
}

type RunState string

const (
	StateQueued       RunState = "QUEUED"
	StateRunning      RunState = "RUNNING"
	StateInterrupting RunState = "INTERRUPTING"
	StateInterrupted  RunState = "INTERRUPTED"
	StateResuming     RunState = "RESUMING"
	StateCompleted    RunState = "COMPLETED"
	StateFailed       RunState = "FAILED"
	StateCancelled    RunState = "CANCELLED"
)

type Event struct {
	ID        string                 `json:"id"`
	RunID     string                 `json:"run_id"`
	Type      string                 `json:"type"` // "agent.speak" | "tool.call" | "tool.return" | "state.change" | "system.log"
	Timestamp int64                  `json:"timestamp"`
	Sender    string                 `json:"sender"`
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata"`
}

type Command struct {
	Type    string `json:"type"` // "run.interrupt" | "run.resume" | "run.abort"
	RunID   string `json:"run_id"`
	Message string `json:"message,omitempty"` // Supervisor feedback for resume
}
