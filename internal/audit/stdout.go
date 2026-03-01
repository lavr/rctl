package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// StderrLogger writes audit entries as JSON lines to stderr.
type StderrLogger struct {
	w io.Writer
}

// NewStderrLogger creates a logger that writes to stderr.
func NewStderrLogger() *StderrLogger {
	return &StderrLogger{w: os.Stderr}
}

func (l *StderrLogger) Log(entry Entry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshaling audit entry: %w", err)
	}
	data = append(data, '\n')
	_, err = l.w.Write(data)
	return err
}

func (l *StderrLogger) Close() error {
	return nil
}
