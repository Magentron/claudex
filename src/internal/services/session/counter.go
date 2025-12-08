package session

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/afero"
)

const (
	// DocUpdateCounterFile is the filename for the auto-doc update counter
	DocUpdateCounterFile = ".doc-update-counter"

	// LastProcessedLineFile is the filename for the last processed line tracker
	LastProcessedLineFile = ".last-processed-line-overview"
)

// ReadCounter reads an integer counter from a file in the session folder.
// Returns 0 if the file does not exist or is empty.
// Returns an error only if the file exists but contains invalid data.
func ReadCounter(fs afero.Fs, sessionPath string) (int, error) {
	path := filepath.Join(sessionPath, DocUpdateCounterFile)
	return readIntFile(fs, path)
}

// WriteCounter writes an integer counter to a file in the session folder.
func WriteCounter(fs afero.Fs, sessionPath string, value int) error {
	path := filepath.Join(sessionPath, DocUpdateCounterFile)
	return writeIntFile(fs, path, value)
}

// IncrementCounter atomically reads, increments, and writes the counter.
// Returns the new counter value.
func IncrementCounter(fs afero.Fs, sessionPath string) (int, error) {
	current, err := ReadCounter(fs, sessionPath)
	if err != nil {
		return 0, fmt.Errorf("failed to read counter: %w", err)
	}

	newValue := current + 1
	if err := WriteCounter(fs, sessionPath, newValue); err != nil {
		return 0, fmt.Errorf("failed to write incremented counter: %w", err)
	}

	return newValue, nil
}

// ResetCounter sets the counter to 0.
func ResetCounter(fs afero.Fs, sessionPath string) error {
	return WriteCounter(fs, sessionPath, 0)
}

// ReadLastProcessedLine reads the last processed line number for transcript tracking.
// Returns 0 if the file does not exist (meaning no lines have been processed yet).
func ReadLastProcessedLine(fs afero.Fs, sessionPath string) (int, error) {
	path := filepath.Join(sessionPath, LastProcessedLineFile)
	return readIntFile(fs, path)
}

// WriteLastProcessedLine writes the last processed line number.
func WriteLastProcessedLine(fs afero.Fs, sessionPath string, line int) error {
	path := filepath.Join(sessionPath, LastProcessedLineFile)
	return writeIntFile(fs, path, line)
}

// readIntFile reads an integer from a file, returning 0 if the file doesn't exist.
func readIntFile(fs afero.Fs, path string) (int, error) {
	data, err := afero.ReadFile(fs, path)
	if err != nil {
		// File doesn't exist - return default value of 0
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	// Handle empty file
	content := strings.TrimSpace(string(data))
	if content == "" {
		return 0, nil
	}

	// Parse integer
	value, err := strconv.Atoi(content)
	if err != nil {
		return 0, fmt.Errorf("invalid integer in file %s: %w", path, err)
	}

	return value, nil
}

// writeIntFile writes an integer to a file.
func writeIntFile(fs afero.Fs, path string, value int) error {
	content := fmt.Sprintf("%d", value)
	if err := afero.WriteFile(fs, path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}
	return nil
}
