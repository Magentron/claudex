package session

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

const (
	// DescriptionFile is the filename for session description
	DescriptionFile = ".description"

	// CreatedFile is the filename for creation timestamp
	CreatedFile = ".created"

	// LastUsedFile is the filename for last used timestamp
	LastUsedFile = ".last_used"
)

// SessionMetadata represents metadata files stored in a session folder.
type SessionMetadata struct {
	Description string // Content of .description file
	Created     string // Content of .created file (RFC3339 timestamp)
	LastUsed    string // Content of .last_used file (RFC3339 timestamp)
}

// ReadMetadata reads all metadata files from a session folder.
// Missing files result in empty strings in the returned struct (not an error).
// Only returns an error if reading fails for reasons other than file not existing.
func ReadMetadata(fs afero.Fs, sessionPath string) (*SessionMetadata, error) {
	metadata := &SessionMetadata{}

	// Read description
	desc, err := readMetadataFile(fs, filepath.Join(sessionPath, DescriptionFile))
	if err != nil {
		return nil, fmt.Errorf("failed to read description: %w", err)
	}
	metadata.Description = desc

	// Read created timestamp
	created, err := readMetadataFile(fs, filepath.Join(sessionPath, CreatedFile))
	if err != nil {
		return nil, fmt.Errorf("failed to read created timestamp: %w", err)
	}
	metadata.Created = created

	// Read last used timestamp
	lastUsed, err := readMetadataFile(fs, filepath.Join(sessionPath, LastUsedFile))
	if err != nil {
		return nil, fmt.Errorf("failed to read last used timestamp: %w", err)
	}
	metadata.LastUsed = lastUsed

	return metadata, nil
}

// ReadDescription reads only the description file from a session folder.
// Returns empty string if the file doesn't exist.
func ReadDescription(fs afero.Fs, sessionPath string) (string, error) {
	path := filepath.Join(sessionPath, DescriptionFile)
	return readMetadataFile(fs, path)
}

// ReadCreatedTimestamp reads only the created timestamp file from a session folder.
// Returns empty string if the file doesn't exist.
func ReadCreatedTimestamp(fs afero.Fs, sessionPath string) (string, error) {
	path := filepath.Join(sessionPath, CreatedFile)
	return readMetadataFile(fs, path)
}

// ReadLastUsedTimestamp reads only the last used timestamp file from a session folder.
// Returns empty string if the file doesn't exist.
func ReadLastUsedTimestamp(fs afero.Fs, sessionPath string) (string, error) {
	path := filepath.Join(sessionPath, LastUsedFile)
	return readMetadataFile(fs, path)
}

// readMetadataFile reads a metadata file and returns its trimmed content.
// Returns empty string if file doesn't exist (not an error).
func readMetadataFile(fs afero.Fs, path string) (string, error) {
	data, err := afero.ReadFile(fs, path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}
