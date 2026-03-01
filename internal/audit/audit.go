package audit

import "time"

// Logger is the interface for audit logging.
type Logger interface {
	// Log writes an audit entry.
	Log(entry Entry) error
	// Close flushes and closes the logger.
	Close() error
}

// Entry represents a single audit log record.
type Entry struct {
	Timestamp  string            `json:"timestamp"`
	User       string            `json:"user"`
	Client     string            `json:"client"`
	Domain     string            `json:"domain"`
	Cwd        string            `json:"cwd"`
	ExecPath   string            `json:"exec_path"`
	Argv       []string          `json:"argv"`
	EnvChanged map[string]string `json:"env_changed"`
	ExitCode   int               `json:"exit_code"`
	DurationMs int64             `json:"duration_ms"`
}

// NewEntry creates a new Entry with the timestamp set to now.
func NewEntry() Entry {
	return Entry{
		Timestamp: time.Now().Format(time.RFC3339),
	}
}
