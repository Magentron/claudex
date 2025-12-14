package doctracking

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/afero"
)

const (
	trackingFileName = "doc_update_tracking.json"
	strategyVersion  = "v1"
)

// FileTrackingService is the production implementation of TrackingService
type FileTrackingService struct {
	fs          afero.Fs
	sessionPath string
}

// New creates a new TrackingService instance
func New(fs afero.Fs, sessionPath string) TrackingService {
	return &FileTrackingService{
		fs:          fs,
		sessionPath: sessionPath,
	}
}

// Read loads the current tracking state from storage
func (fts *FileTrackingService) Read() (DocUpdateTracking, error) {
	trackingPath := filepath.Join(fts.sessionPath, trackingFileName)

	data, err := afero.ReadFile(fts.fs, trackingPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return zero value if file doesn't exist
			return DocUpdateTracking{}, nil
		}
		return DocUpdateTracking{}, err
	}

	var tracking DocUpdateTracking
	if err := json.Unmarshal(data, &tracking); err != nil {
		return DocUpdateTracking{}, err
	}

	return tracking, nil
}

// Write persists the tracking state to storage atomically
func (fts *FileTrackingService) Write(tracking DocUpdateTracking) error {
	trackingPath := filepath.Join(fts.sessionPath, trackingFileName)
	tempPath := trackingPath + ".tmp"

	// Marshal to JSON with indentation for readability
	data, err := json.MarshalIndent(tracking, "", "  ")
	if err != nil {
		return err
	}

	// Write to temp file first
	if err := afero.WriteFile(fts.fs, tempPath, data, 0644); err != nil {
		return err
	}

	// Atomic rename
	return fts.fs.Rename(tempPath, trackingPath)
}

// Initialize creates initial tracking state with HEAD commit
func (fts *FileTrackingService) Initialize(headSHA string) error {
	tracking := DocUpdateTracking{
		LastProcessedCommit: headSHA,
		UpdatedAt:           time.Now().Format(time.RFC3339),
		StrategyVersion:     strategyVersion,
	}

	return fts.Write(tracking)
}
