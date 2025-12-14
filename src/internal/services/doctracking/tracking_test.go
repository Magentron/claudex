package doctracking

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileTrackingService_Read_MissingFile(t *testing.T) {
	// Setup
	fs := afero.NewMemMapFs()
	sessionPath := "/test/session"
	service := New(fs, sessionPath)

	// Execute
	tracking, err := service.Read()

	// Verify
	require.NoError(t, err)
	assert.Equal(t, "", tracking.LastProcessedCommit)
	assert.Equal(t, "", tracking.UpdatedAt)
	assert.Equal(t, "", tracking.StrategyVersion)
}

func TestFileTrackingService_Read_ValidJSON(t *testing.T) {
	// Setup
	fs := afero.NewMemMapFs()
	sessionPath := "/test/session"
	require.NoError(t, fs.MkdirAll(sessionPath, 0755))

	expectedTracking := DocUpdateTracking{
		LastProcessedCommit: "abc123",
		UpdatedAt:           "2025-12-13T10:00:00Z",
		StrategyVersion:     "v1",
	}

	data, err := json.Marshal(expectedTracking)
	require.NoError(t, err)

	trackingPath := filepath.Join(sessionPath, trackingFileName)
	require.NoError(t, afero.WriteFile(fs, trackingPath, data, 0644))

	service := New(fs, sessionPath)

	// Execute
	tracking, err := service.Read()

	// Verify
	require.NoError(t, err)
	assert.Equal(t, expectedTracking.LastProcessedCommit, tracking.LastProcessedCommit)
	assert.Equal(t, expectedTracking.UpdatedAt, tracking.UpdatedAt)
	assert.Equal(t, expectedTracking.StrategyVersion, tracking.StrategyVersion)
}

func TestFileTrackingService_Read_InvalidJSON(t *testing.T) {
	// Setup
	fs := afero.NewMemMapFs()
	sessionPath := "/test/session"
	require.NoError(t, fs.MkdirAll(sessionPath, 0755))

	trackingPath := filepath.Join(sessionPath, trackingFileName)
	require.NoError(t, afero.WriteFile(fs, trackingPath, []byte("invalid json"), 0644))

	service := New(fs, sessionPath)

	// Execute
	_, err := service.Read()

	// Verify
	require.Error(t, err)
}

func TestFileTrackingService_Write_CreatesFile(t *testing.T) {
	// Setup
	fs := afero.NewMemMapFs()
	sessionPath := "/test/session"
	require.NoError(t, fs.MkdirAll(sessionPath, 0755))

	service := New(fs, sessionPath)

	tracking := DocUpdateTracking{
		LastProcessedCommit: "def456",
		UpdatedAt:           "2025-12-13T11:00:00Z",
		StrategyVersion:     "v1",
	}

	// Execute
	err := service.Write(tracking)

	// Verify
	require.NoError(t, err)

	// Verify file was created
	trackingPath := filepath.Join(sessionPath, trackingFileName)
	exists, err := afero.Exists(fs, trackingPath)
	require.NoError(t, err)
	assert.True(t, exists)

	// Verify content
	data, err := afero.ReadFile(fs, trackingPath)
	require.NoError(t, err)

	var readTracking DocUpdateTracking
	require.NoError(t, json.Unmarshal(data, &readTracking))
	assert.Equal(t, tracking.LastProcessedCommit, readTracking.LastProcessedCommit)
	assert.Equal(t, tracking.UpdatedAt, readTracking.UpdatedAt)
	assert.Equal(t, tracking.StrategyVersion, readTracking.StrategyVersion)
}

func TestFileTrackingService_Write_UpdatesFile(t *testing.T) {
	// Setup
	fs := afero.NewMemMapFs()
	sessionPath := "/test/session"
	require.NoError(t, fs.MkdirAll(sessionPath, 0755))

	service := New(fs, sessionPath)

	// Write initial tracking
	initialTracking := DocUpdateTracking{
		LastProcessedCommit: "initial123",
		UpdatedAt:           "2025-12-13T09:00:00Z",
		StrategyVersion:     "v1",
	}
	require.NoError(t, service.Write(initialTracking))

	// Update tracking
	updatedTracking := DocUpdateTracking{
		LastProcessedCommit: "updated456",
		UpdatedAt:           "2025-12-13T12:00:00Z",
		StrategyVersion:     "v1",
	}

	// Execute
	err := service.Write(updatedTracking)

	// Verify
	require.NoError(t, err)

	// Verify updated content
	readTracking, err := service.Read()
	require.NoError(t, err)
	assert.Equal(t, updatedTracking.LastProcessedCommit, readTracking.LastProcessedCommit)
	assert.Equal(t, updatedTracking.UpdatedAt, readTracking.UpdatedAt)
	assert.Equal(t, updatedTracking.StrategyVersion, readTracking.StrategyVersion)
}

func TestFileTrackingService_Write_NoTempFileRemains(t *testing.T) {
	// Setup
	fs := afero.NewMemMapFs()
	sessionPath := "/test/session"
	require.NoError(t, fs.MkdirAll(sessionPath, 0755))

	service := New(fs, sessionPath)

	tracking := DocUpdateTracking{
		LastProcessedCommit: "abc123",
		UpdatedAt:           "2025-12-13T10:00:00Z",
		StrategyVersion:     "v1",
	}

	// Execute
	err := service.Write(tracking)
	require.NoError(t, err)

	// Verify temp file doesn't exist
	tempPath := filepath.Join(sessionPath, trackingFileName+".tmp")
	exists, err := afero.Exists(fs, tempPath)
	require.NoError(t, err)
	assert.False(t, exists, "temporary file should not remain after successful write")
}

func TestFileTrackingService_JSONRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		tracking DocUpdateTracking
	}{
		{
			name: "complete tracking",
			tracking: DocUpdateTracking{
				LastProcessedCommit: "abc123def456",
				UpdatedAt:           "2025-12-13T10:00:00Z",
				StrategyVersion:     "v1",
			},
		},
		{
			name: "empty commit",
			tracking: DocUpdateTracking{
				LastProcessedCommit: "",
				UpdatedAt:           "2025-12-13T10:00:00Z",
				StrategyVersion:     "v1",
			},
		},
		{
			name: "different strategy version",
			tracking: DocUpdateTracking{
				LastProcessedCommit: "xyz789",
				UpdatedAt:           "2025-12-13T10:00:00Z",
				StrategyVersion:     "v2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			fs := afero.NewMemMapFs()
			sessionPath := "/test/session"
			require.NoError(t, fs.MkdirAll(sessionPath, 0755))

			service := New(fs, sessionPath)

			// Execute write
			err := service.Write(tt.tracking)
			require.NoError(t, err)

			// Execute read
			readTracking, err := service.Read()
			require.NoError(t, err)

			// Verify
			assert.Equal(t, tt.tracking.LastProcessedCommit, readTracking.LastProcessedCommit)
			assert.Equal(t, tt.tracking.UpdatedAt, readTracking.UpdatedAt)
			assert.Equal(t, tt.tracking.StrategyVersion, readTracking.StrategyVersion)
		})
	}
}

func TestFileTrackingService_Initialize(t *testing.T) {
	// Setup
	fs := afero.NewMemMapFs()
	sessionPath := "/test/session"
	require.NoError(t, fs.MkdirAll(sessionPath, 0755))

	service := New(fs, sessionPath)
	headSHA := "abc123def456"

	// Execute
	err := service.Initialize(headSHA)

	// Verify
	require.NoError(t, err)

	// Read and verify created tracking
	tracking, err := service.Read()
	require.NoError(t, err)

	assert.Equal(t, headSHA, tracking.LastProcessedCommit)
	assert.Equal(t, "v1", tracking.StrategyVersion)

	// Verify timestamp is valid RFC3339
	_, err = time.Parse(time.RFC3339, tracking.UpdatedAt)
	assert.NoError(t, err, "UpdatedAt should be valid RFC3339 timestamp")

	// Verify timestamp is recent (within last 5 seconds)
	updatedAt, err := time.Parse(time.RFC3339, tracking.UpdatedAt)
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now(), updatedAt, 5*time.Second)
}

func TestFileTrackingService_Initialize_OverwritesExisting(t *testing.T) {
	// Setup
	fs := afero.NewMemMapFs()
	sessionPath := "/test/session"
	require.NoError(t, fs.MkdirAll(sessionPath, 0755))

	service := New(fs, sessionPath)

	// Create initial tracking
	initialSHA := "initial123"
	require.NoError(t, service.Initialize(initialSHA))

	// Wait a bit to ensure different timestamp
	time.Sleep(10 * time.Millisecond)

	// Initialize with new SHA
	newSHA := "new456"
	err := service.Initialize(newSHA)
	require.NoError(t, err)

	// Verify
	tracking, err := service.Read()
	require.NoError(t, err)
	assert.Equal(t, newSHA, tracking.LastProcessedCommit)
	assert.Equal(t, "v1", tracking.StrategyVersion)
}
