package audit

// NoopLogger discards all audit entries.
type NoopLogger struct{}

func (l *NoopLogger) Log(entry Entry) error { return nil }
func (l *NoopLogger) Close() error          { return nil }
