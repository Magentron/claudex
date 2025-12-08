package session

import (
	"testing"

	"claudex/internal/testutil"

	"github.com/stretchr/testify/require"
)

// Test_ReadMetadata_AllFilesExist tests reading all metadata files
func Test_ReadMetadata_AllFilesExist(t *testing.T) {
	h := testutil.NewTestHarness()

	sessionPath := "/sessions/test-session"
	h.CreateSessionWithFiles(sessionPath, map[string]string{
		".description": "My awesome feature",
		".created":     "2024-01-15T10:30:00Z",
		".last_used":   "2024-01-15T14:00:00Z",
	})

	// Exercise
	metadata, err := ReadMetadata(h.FS, sessionPath)

	// Verify
	require.NoError(t, err)
	require.NotNil(t, metadata)
	require.Equal(t, "My awesome feature", metadata.Description)
	require.Equal(t, "2024-01-15T10:30:00Z", metadata.Created)
	require.Equal(t, "2024-01-15T14:00:00Z", metadata.LastUsed)
}

// Test_ReadMetadata_SomeFilesMissing tests reading with some files missing
func Test_ReadMetadata_SomeFilesMissing(t *testing.T) {
	h := testutil.NewTestHarness()

	sessionPath := "/sessions/test-session"
	h.CreateSessionWithFiles(sessionPath, map[string]string{
		".description": "Feature description",
		// .created and .last_used are missing
	})

	// Exercise
	metadata, err := ReadMetadata(h.FS, sessionPath)

	// Verify - missing files should result in empty strings
	require.NoError(t, err)
	require.NotNil(t, metadata)
	require.Equal(t, "Feature description", metadata.Description)
	require.Equal(t, "", metadata.Created)
	require.Equal(t, "", metadata.LastUsed)
}

// Test_ReadMetadata_AllFilesMissing tests reading with no metadata files
func Test_ReadMetadata_AllFilesMissing(t *testing.T) {
	h := testutil.NewTestHarness()

	sessionPath := "/sessions/test-session"
	h.CreateDir(sessionPath)

	// Exercise
	metadata, err := ReadMetadata(h.FS, sessionPath)

	// Verify - should succeed with empty metadata
	require.NoError(t, err)
	require.NotNil(t, metadata)
	require.Equal(t, "", metadata.Description)
	require.Equal(t, "", metadata.Created)
	require.Equal(t, "", metadata.LastUsed)
}

// Test_ReadMetadata_WhitespaceHandling tests that content is trimmed
func Test_ReadMetadata_WhitespaceHandling(t *testing.T) {
	h := testutil.NewTestHarness()

	sessionPath := "/sessions/test-session"
	h.CreateSessionWithFiles(sessionPath, map[string]string{
		".description": "  Feature with whitespace  \n",
		".created":     "\n2024-01-15T10:30:00Z\n  ",
		".last_used":   "  \n  2024-01-15T14:00:00Z",
	})

	// Exercise
	metadata, err := ReadMetadata(h.FS, sessionPath)

	// Verify - whitespace should be trimmed
	require.NoError(t, err)
	require.Equal(t, "Feature with whitespace", metadata.Description)
	require.Equal(t, "2024-01-15T10:30:00Z", metadata.Created)
	require.Equal(t, "2024-01-15T14:00:00Z", metadata.LastUsed)
}

// Test_ReadMetadata_EmptyFiles tests reading empty files
func Test_ReadMetadata_EmptyFiles(t *testing.T) {
	h := testutil.NewTestHarness()

	sessionPath := "/sessions/test-session"
	h.CreateSessionWithFiles(sessionPath, map[string]string{
		".description": "",
		".created":     "",
		".last_used":   "",
	})

	// Exercise
	metadata, err := ReadMetadata(h.FS, sessionPath)

	// Verify - empty files should result in empty strings
	require.NoError(t, err)
	require.Equal(t, "", metadata.Description)
	require.Equal(t, "", metadata.Created)
	require.Equal(t, "", metadata.LastUsed)
}

// Test_ReadDescription tests reading only description
func Test_ReadDescription(t *testing.T) {
	h := testutil.NewTestHarness()

	sessionPath := "/sessions/test-session"
	h.CreateSessionWithFiles(sessionPath, map[string]string{
		".description": "Login feature implementation",
	})

	// Exercise
	description, err := ReadDescription(h.FS, sessionPath)

	// Verify
	require.NoError(t, err)
	require.Equal(t, "Login feature implementation", description)
}

// Test_ReadDescription_Missing tests reading missing description
func Test_ReadDescription_Missing(t *testing.T) {
	h := testutil.NewTestHarness()

	sessionPath := "/sessions/test-session"
	h.CreateDir(sessionPath)

	// Exercise
	description, err := ReadDescription(h.FS, sessionPath)

	// Verify - should return empty string
	require.NoError(t, err)
	require.Equal(t, "", description)
}

// Test_ReadCreatedTimestamp tests reading created timestamp
func Test_ReadCreatedTimestamp(t *testing.T) {
	h := testutil.NewTestHarness()

	sessionPath := "/sessions/test-session"
	h.CreateSessionWithFiles(sessionPath, map[string]string{
		".created": "2024-01-15T10:30:00Z",
	})

	// Exercise
	created, err := ReadCreatedTimestamp(h.FS, sessionPath)

	// Verify
	require.NoError(t, err)
	require.Equal(t, "2024-01-15T10:30:00Z", created)
}

// Test_ReadCreatedTimestamp_Missing tests reading missing created timestamp
func Test_ReadCreatedTimestamp_Missing(t *testing.T) {
	h := testutil.NewTestHarness()

	sessionPath := "/sessions/test-session"
	h.CreateDir(sessionPath)

	// Exercise
	created, err := ReadCreatedTimestamp(h.FS, sessionPath)

	// Verify - should return empty string
	require.NoError(t, err)
	require.Equal(t, "", created)
}

// Test_ReadLastUsedTimestamp tests reading last used timestamp
func Test_ReadLastUsedTimestamp(t *testing.T) {
	h := testutil.NewTestHarness()

	sessionPath := "/sessions/test-session"
	h.CreateSessionWithFiles(sessionPath, map[string]string{
		".last_used": "2024-01-15T14:00:00Z",
	})

	// Exercise
	lastUsed, err := ReadLastUsedTimestamp(h.FS, sessionPath)

	// Verify
	require.NoError(t, err)
	require.Equal(t, "2024-01-15T14:00:00Z", lastUsed)
}

// Test_ReadLastUsedTimestamp_Missing tests reading missing last used timestamp
func Test_ReadLastUsedTimestamp_Missing(t *testing.T) {
	h := testutil.NewTestHarness()

	sessionPath := "/sessions/test-session"
	h.CreateDir(sessionPath)

	// Exercise
	lastUsed, err := ReadLastUsedTimestamp(h.FS, sessionPath)

	// Verify - should return empty string
	require.NoError(t, err)
	require.Equal(t, "", lastUsed)
}

// Test_ReadMetadata_MultilineDescription tests description with multiple lines
func Test_ReadMetadata_MultilineDescription(t *testing.T) {
	h := testutil.NewTestHarness()

	sessionPath := "/sessions/test-session"
	h.CreateSessionWithFiles(sessionPath, map[string]string{
		".description": "Line 1\nLine 2\nLine 3",
	})

	// Exercise
	metadata, err := ReadMetadata(h.FS, sessionPath)

	// Verify - multiline content should be preserved (only outer whitespace trimmed)
	require.NoError(t, err)
	require.Equal(t, "Line 1\nLine 2\nLine 3", metadata.Description)
}

// Test_ReadMetadata_SpecialCharacters tests handling of special characters
func Test_ReadMetadata_SpecialCharacters(t *testing.T) {
	h := testutil.NewTestHarness()

	sessionPath := "/sessions/test-session"
	h.CreateSessionWithFiles(sessionPath, map[string]string{
		".description": "Feature: Add 'quotes' and \"double quotes\" & special chars!",
	})

	// Exercise
	metadata, err := ReadMetadata(h.FS, sessionPath)

	// Verify - special characters should be preserved
	require.NoError(t, err)
	require.Equal(t, "Feature: Add 'quotes' and \"double quotes\" & special chars!", metadata.Description)
}

// Test_ReadMetadata_UnicodeCharacters tests handling of unicode characters
func Test_ReadMetadata_UnicodeCharacters(t *testing.T) {
	h := testutil.NewTestHarness()

	sessionPath := "/sessions/test-session"
	h.CreateSessionWithFiles(sessionPath, map[string]string{
		".description": "Implement feature 🚀 with emoji support",
	})

	// Exercise
	metadata, err := ReadMetadata(h.FS, sessionPath)

	// Verify - unicode should be preserved
	require.NoError(t, err)
	require.Equal(t, "Implement feature 🚀 with emoji support", metadata.Description)
}
