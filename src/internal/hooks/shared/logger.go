package shared

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/afero"
)

// Logger provides unified logging for hooks
type Logger struct {
	fs       afero.Fs
	env      Environment
	hookName string
}

// Environment abstracts environment variable access for testing
type Environment interface {
	Get(key string) string
	Set(key, value string)
}

// NewLogger creates a new Logger instance
func NewLogger(fs afero.Fs, env Environment, hookName string) *Logger {
	return &Logger{
		fs:       fs,
		env:      env,
		hookName: hookName,
	}
}

// Log writes a log message to the configured log file
func (l *Logger) Log(message string) error {
	logPath := l.env.Get("CLAUDEX_LOG_FILE")
	if logPath == "" {
		// If no log file configured, silently skip logging
		return nil
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("%s | [%s] %s\n", timestamp, l.hookName, message)

	// Append to log file
	file, err := l.fs.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	if _, err := io.WriteString(file, logEntry); err != nil {
		return fmt.Errorf("failed to write log entry: %w", err)
	}

	return nil
}

// Logf writes a formatted log message to the configured log file
func (l *Logger) Logf(format string, args ...interface{}) error {
	return l.Log(fmt.Sprintf(format, args...))
}

// LogError logs an error message
func (l *Logger) LogError(err error) error {
	return l.Logf("ERROR: %v", err)
}

// LogInfo logs an informational message
func (l *Logger) LogInfo(message string) error {
	return l.Logf("INFO: %s", message)
}

// LogDebug logs a debug message
func (l *Logger) LogDebug(message string) error {
	return l.Logf("DEBUG: %s", message)
}
