package preferences

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

const (
	preferencesFileName = "preferences.json"
	claudexDir          = ".claudex"
)

// FileService is the production implementation of Service
type FileService struct {
	fs         afero.Fs
	projectDir string
}

// New creates a new Service instance
func New(fs afero.Fs, projectDir string) Service {
	return &FileService{
		fs:         fs,
		projectDir: projectDir,
	}
}

// Load reads preferences from storage
// Returns zero-value Preferences if file doesn't exist
func (fs *FileService) Load() (Preferences, error) {
	prefsPath := filepath.Join(fs.projectDir, claudexDir, preferencesFileName)

	data, err := afero.ReadFile(fs.fs, prefsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return zero value if file doesn't exist
			return Preferences{}, nil
		}
		return Preferences{}, err
	}

	var prefs Preferences
	if err := json.Unmarshal(data, &prefs); err != nil {
		return Preferences{}, err
	}

	return prefs, nil
}

// Save persists preferences to storage atomically
func (fs *FileService) Save(prefs Preferences) error {
	claudexPath := filepath.Join(fs.projectDir, claudexDir)
	prefsPath := filepath.Join(claudexPath, preferencesFileName)
	tempPath := prefsPath + ".tmp"

	// Ensure .claudex directory exists
	if err := fs.fs.MkdirAll(claudexPath, 0755); err != nil {
		return err
	}

	// Marshal to JSON with indentation for readability
	data, err := json.MarshalIndent(prefs, "", "  ")
	if err != nil {
		return err
	}

	// Write to temp file first
	if err := afero.WriteFile(fs.fs, tempPath, data, 0644); err != nil {
		return err
	}

	// Atomic rename
	return fs.fs.Rename(tempPath, prefsPath)
}
