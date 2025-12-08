package session

import (
	"path/filepath"
	"testing"

	"claudex/internal/testutil"

	"github.com/stretchr/testify/require"
)

// Test_ReadCounter_FileExists tests reading an existing counter file
func Test_ReadCounter_FileExists(t *testing.T) {
	h := testutil.NewTestHarness()

	sessionPath := "/sessions/test-session"
	h.CreateSessionWithFiles(sessionPath, map[string]string{
		".doc-update-counter": "5",
	})

	// Exercise
	result, err := ReadCounter(h.FS, sessionPath)

	// Verify
	require.NoError(t, err)
	require.Equal(t, 5, result)
}

// Test_ReadCounter_FileNotExists tests reading when counter file doesn't exist
func Test_ReadCounter_FileNotExists(t *testing.T) {
	h := testutil.NewTestHarness()

	sessionPath := "/sessions/test-session"
	h.CreateDir(sessionPath)

	// Exercise
	result, err := ReadCounter(h.FS, sessionPath)

	// Verify - should return 0 (default)
	require.NoError(t, err)
	require.Equal(t, 0, result)
}

// Test_ReadCounter_EmptyFile tests reading an empty counter file
func Test_ReadCounter_EmptyFile(t *testing.T) {
	h := testutil.NewTestHarness()

	sessionPath := "/sessions/test-session"
	h.CreateSessionWithFiles(sessionPath, map[string]string{
		".doc-update-counter": "",
	})

	// Exercise
	result, err := ReadCounter(h.FS, sessionPath)

	// Verify - should return 0 for empty file
	require.NoError(t, err)
	require.Equal(t, 0, result)
}

// Test_ReadCounter_WhitespaceFile tests reading a counter file with only whitespace
func Test_ReadCounter_WhitespaceFile(t *testing.T) {
	h := testutil.NewTestHarness()

	sessionPath := "/sessions/test-session"
	h.CreateSessionWithFiles(sessionPath, map[string]string{
		".doc-update-counter": "  \n  ",
	})

	// Exercise
	result, err := ReadCounter(h.FS, sessionPath)

	// Verify - should return 0 for whitespace-only file
	require.NoError(t, err)
	require.Equal(t, 0, result)
}

// Test_ReadCounter_InvalidContent tests error on non-integer content
func Test_ReadCounter_InvalidContent(t *testing.T) {
	h := testutil.NewTestHarness()

	sessionPath := "/sessions/test-session"
	h.CreateSessionWithFiles(sessionPath, map[string]string{
		".doc-update-counter": "not-a-number",
	})

	// Exercise
	_, err := ReadCounter(h.FS, sessionPath)

	// Verify - should error
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid integer")
}

// Test_WriteCounter tests writing a counter value
func Test_WriteCounter(t *testing.T) {
	h := testutil.NewTestHarness()

	sessionPath := "/sessions/test-session"
	h.CreateDir(sessionPath)

	// Exercise
	err := WriteCounter(h.FS, sessionPath, 42)

	// Verify
	require.NoError(t, err)
	testutil.AssertFileExists(t, h.FS, filepath.Join(sessionPath, ".doc-update-counter"))
	testutil.AssertFileContains(t, h.FS, filepath.Join(sessionPath, ".doc-update-counter"), "42")
}

// Test_WriteCounter_Overwrite tests overwriting an existing counter
func Test_WriteCounter_Overwrite(t *testing.T) {
	h := testutil.NewTestHarness()

	sessionPath := "/sessions/test-session"
	h.CreateSessionWithFiles(sessionPath, map[string]string{
		".doc-update-counter": "10",
	})

	// Exercise
	err := WriteCounter(h.FS, sessionPath, 20)

	// Verify
	require.NoError(t, err)
	testutil.AssertFileContains(t, h.FS, filepath.Join(sessionPath, ".doc-update-counter"), "20")

	// Verify old value is gone
	result, _ := ReadCounter(h.FS, sessionPath)
	require.Equal(t, 20, result)
}

// Test_IncrementCounter tests atomic increment operation
func Test_IncrementCounter(t *testing.T) {
	h := testutil.NewTestHarness()

	sessionPath := "/sessions/test-session"
	h.CreateSessionWithFiles(sessionPath, map[string]string{
		".doc-update-counter": "5",
	})

	// Exercise
	newValue, err := IncrementCounter(h.FS, sessionPath)

	// Verify
	require.NoError(t, err)
	require.Equal(t, 6, newValue)

	// Verify file was updated
	storedValue, _ := ReadCounter(h.FS, sessionPath)
	require.Equal(t, 6, storedValue)
}

// Test_IncrementCounter_FromZero tests incrementing from default value
func Test_IncrementCounter_FromZero(t *testing.T) {
	h := testutil.NewTestHarness()

	sessionPath := "/sessions/test-session"
	h.CreateDir(sessionPath)

	// Exercise - no counter file exists
	newValue, err := IncrementCounter(h.FS, sessionPath)

	// Verify - should increment from 0 to 1
	require.NoError(t, err)
	require.Equal(t, 1, newValue)

	// Verify file was created
	testutil.AssertFileExists(t, h.FS, filepath.Join(sessionPath, ".doc-update-counter"))
}

// Test_ResetCounter tests resetting counter to zero
func Test_ResetCounter(t *testing.T) {
	h := testutil.NewTestHarness()

	sessionPath := "/sessions/test-session"
	h.CreateSessionWithFiles(sessionPath, map[string]string{
		".doc-update-counter": "42",
	})

	// Exercise
	err := ResetCounter(h.FS, sessionPath)

	// Verify
	require.NoError(t, err)

	// Verify counter is now 0
	result, _ := ReadCounter(h.FS, sessionPath)
	require.Equal(t, 0, result)
}

// Test_ReadLastProcessedLine tests reading last processed line
func Test_ReadLastProcessedLine(t *testing.T) {
	h := testutil.NewTestHarness()

	sessionPath := "/sessions/test-session"
	h.CreateSessionWithFiles(sessionPath, map[string]string{
		".last-processed-line-overview": "150",
	})

	// Exercise
	result, err := ReadLastProcessedLine(h.FS, sessionPath)

	// Verify
	require.NoError(t, err)
	require.Equal(t, 150, result)
}

// Test_ReadLastProcessedLine_NotExists tests default value when file doesn't exist
func Test_ReadLastProcessedLine_NotExists(t *testing.T) {
	h := testutil.NewTestHarness()

	sessionPath := "/sessions/test-session"
	h.CreateDir(sessionPath)

	// Exercise
	result, err := ReadLastProcessedLine(h.FS, sessionPath)

	// Verify - should return 0 (start from beginning)
	require.NoError(t, err)
	require.Equal(t, 0, result)
}

// Test_WriteLastProcessedLine tests writing last processed line
func Test_WriteLastProcessedLine(t *testing.T) {
	h := testutil.NewTestHarness()

	sessionPath := "/sessions/test-session"
	h.CreateDir(sessionPath)

	// Exercise
	err := WriteLastProcessedLine(h.FS, sessionPath, 250)

	// Verify
	require.NoError(t, err)
	testutil.AssertFileExists(t, h.FS, filepath.Join(sessionPath, ".last-processed-line-overview"))
	testutil.AssertFileContains(t, h.FS, filepath.Join(sessionPath, ".last-processed-line-overview"), "250")
}

// Test_WriteLastProcessedLine_Update tests updating last processed line
func Test_WriteLastProcessedLine_Update(t *testing.T) {
	h := testutil.NewTestHarness()

	sessionPath := "/sessions/test-session"
	h.CreateSessionWithFiles(sessionPath, map[string]string{
		".last-processed-line-overview": "100",
	})

	// Exercise
	err := WriteLastProcessedLine(h.FS, sessionPath, 200)

	// Verify
	require.NoError(t, err)

	// Verify new value
	result, _ := ReadLastProcessedLine(h.FS, sessionPath)
	require.Equal(t, 200, result)
}

// Test_CounterOperations_Sequence tests a realistic sequence of operations
func Test_CounterOperations_Sequence(t *testing.T) {
	h := testutil.NewTestHarness()

	sessionPath := "/sessions/test-session"
	h.CreateDir(sessionPath)

	// Start with no counter file
	val, _ := ReadCounter(h.FS, sessionPath)
	require.Equal(t, 0, val)

	// Increment several times
	val1, err := IncrementCounter(h.FS, sessionPath)
	require.NoError(t, err)
	require.Equal(t, 1, val1)

	val2, err := IncrementCounter(h.FS, sessionPath)
	require.NoError(t, err)
	require.Equal(t, 2, val2)

	val3, err := IncrementCounter(h.FS, sessionPath)
	require.NoError(t, err)
	require.Equal(t, 3, val3)

	// Reset
	err = ResetCounter(h.FS, sessionPath)
	require.NoError(t, err)

	// Verify reset
	val, _ = ReadCounter(h.FS, sessionPath)
	require.Equal(t, 0, val)

	// Increment again after reset
	val4, err := IncrementCounter(h.FS, sessionPath)
	require.NoError(t, err)
	require.Equal(t, 1, val4)
}
