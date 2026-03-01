package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// FileLogger writes audit entries as JSON lines to a file.
type FileLogger struct {
	file *os.File
}

// NewFileLogger creates a FileLogger, creating parent directories if needed.
func NewFileLogger(path string) (*FileLogger, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("creating audit log directory: %w", err)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("opening audit log: %w", err)
	}

	return &FileLogger{file: f}, nil
}

func (l *FileLogger) Log(entry Entry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshaling audit entry: %w", err)
	}
	data = append(data, '\n')
	_, err = l.file.Write(data)
	return err
}

func (l *FileLogger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}
