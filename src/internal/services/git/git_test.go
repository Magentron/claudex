package git

import (
	"errors"
	"io"
	"testing"
)

// mockCommander is a mock implementation of commander.Commander for testing
type mockCommander struct {
	runFunc func(name string, args ...string) ([]byte, error)
}

func (m *mockCommander) Run(name string, args ...string) ([]byte, error) {
	if m.runFunc != nil {
		return m.runFunc(name, args...)
	}
	return nil, errors.New("mock not configured")
}

func (m *mockCommander) Start(name string, stdin io.Reader, stdout, stderr io.Writer, args ...string) error {
	return errors.New("Start not implemented in mock")
}

func TestGetCurrentSHA_Success(t *testing.T) {
	expectedSHA := "abc123def456"
	mock := &mockCommander{
		runFunc: func(name string, args ...string) ([]byte, error) {
			if name != "git" {
				t.Errorf("expected command 'git', got '%s'", name)
			}
			if len(args) != 2 || args[0] != "rev-parse" || args[1] != "HEAD" {
				t.Errorf("expected args [rev-parse HEAD], got %v", args)
			}
			return []byte(expectedSHA + "\n"), nil
		},
	}

	svc := New(mock)
	sha, err := svc.GetCurrentSHA()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sha != expectedSHA {
		t.Errorf("expected SHA '%s', got '%s'", expectedSHA, sha)
	}
}

func TestGetCurrentSHA_Error(t *testing.T) {
	expectedErr := errors.New("git command failed")
	mock := &mockCommander{
		runFunc: func(name string, args ...string) ([]byte, error) {
			return nil, expectedErr
		},
	}

	svc := New(mock)
	_, err := svc.GetCurrentSHA()

	if err != expectedErr {
		t.Errorf("expected error '%v', got '%v'", expectedErr, err)
	}
}

func TestGetCurrentSHA_TrimWhitespace(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected string
	}{
		{"trailing newline", "abc123\n", "abc123"},
		{"leading and trailing spaces", "  abc123  \n", "abc123"},
		{"multiple newlines", "abc123\n\n", "abc123"},
		{"tabs", "\tabc123\t", "abc123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockCommander{
				runFunc: func(name string, args ...string) ([]byte, error) {
					return []byte(tt.output), nil
				},
			}

			svc := New(mock)
			sha, err := svc.GetCurrentSHA()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if sha != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, sha)
			}
		})
	}
}

func TestGetChangedFiles_Success(t *testing.T) {
	mock := &mockCommander{
		runFunc: func(name string, args ...string) ([]byte, error) {
			if name != "git" {
				t.Errorf("expected command 'git', got '%s'", name)
			}
			expectedArgs := []string{"diff", "--name-only", "abc123..def456"}
			if len(args) != len(expectedArgs) {
				t.Errorf("expected %d args, got %d", len(expectedArgs), len(args))
			}
			for i, arg := range expectedArgs {
				if i < len(args) && args[i] != arg {
					t.Errorf("arg %d: expected '%s', got '%s'", i, arg, args[i])
				}
			}
			return []byte("file1.go\nfile2.go\nfile3.md\n"), nil
		},
	}

	svc := New(mock)
	files, err := svc.GetChangedFiles("abc123", "def456")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"file1.go", "file2.go", "file3.md"}
	if len(files) != len(expected) {
		t.Fatalf("expected %d files, got %d", len(expected), len(files))
	}
	for i, file := range expected {
		if files[i] != file {
			t.Errorf("file %d: expected '%s', got '%s'", i, file, files[i])
		}
	}
}

func TestGetChangedFiles_EmptyResult(t *testing.T) {
	mock := &mockCommander{
		runFunc: func(name string, args ...string) ([]byte, error) {
			return []byte(""), nil
		},
	}

	svc := New(mock)
	files, err := svc.GetChangedFiles("abc123", "def456")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected empty slice, got %v", files)
	}
}

func TestGetChangedFiles_FilterEmptyLines(t *testing.T) {
	mock := &mockCommander{
		runFunc: func(name string, args ...string) ([]byte, error) {
			return []byte("file1.go\n\nfile2.go\n  \nfile3.md\n\n"), nil
		},
	}

	svc := New(mock)
	files, err := svc.GetChangedFiles("abc123", "def456")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"file1.go", "file2.go", "file3.md"}
	if len(files) != len(expected) {
		t.Fatalf("expected %d files, got %d: %v", len(expected), len(files), files)
	}
	for i, file := range expected {
		if files[i] != file {
			t.Errorf("file %d: expected '%s', got '%s'", i, file, files[i])
		}
	}
}

func TestGetChangedFiles_Error(t *testing.T) {
	expectedErr := errors.New("git diff failed")
	mock := &mockCommander{
		runFunc: func(name string, args ...string) ([]byte, error) {
			return nil, expectedErr
		},
	}

	svc := New(mock)
	_, err := svc.GetChangedFiles("abc123", "def456")

	if err != expectedErr {
		t.Errorf("expected error '%v', got '%v'", expectedErr, err)
	}
}

func TestValidateCommit_Valid(t *testing.T) {
	mock := &mockCommander{
		runFunc: func(name string, args ...string) ([]byte, error) {
			if name != "git" {
				t.Errorf("expected command 'git', got '%s'", name)
			}
			expectedArgs := []string{"cat-file", "-t", "abc123"}
			if len(args) != len(expectedArgs) {
				t.Errorf("expected %d args, got %d", len(expectedArgs), len(args))
			}
			for i, arg := range expectedArgs {
				if i < len(args) && args[i] != arg {
					t.Errorf("arg %d: expected '%s', got '%s'", i, arg, args[i])
				}
			}
			return []byte("commit\n"), nil
		},
	}

	svc := New(mock)
	valid, err := svc.ValidateCommit("abc123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !valid {
		t.Error("expected commit to be valid")
	}
}

func TestValidateCommit_NotCommit(t *testing.T) {
	tests := []struct {
		name   string
		output string
	}{
		{"tree object", "tree\n"},
		{"blob object", "blob\n"},
		{"tag object", "tag\n"},
		{"empty output", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockCommander{
				runFunc: func(name string, args ...string) ([]byte, error) {
					return []byte(tt.output), nil
				},
			}

			svc := New(mock)
			valid, err := svc.ValidateCommit("abc123")

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if valid {
				t.Error("expected commit to be invalid")
			}
		})
	}
}

func TestValidateCommit_CommandError(t *testing.T) {
	mock := &mockCommander{
		runFunc: func(name string, args ...string) ([]byte, error) {
			return nil, errors.New("commit not found")
		},
	}

	svc := New(mock)
	valid, err := svc.ValidateCommit("invalid")

	if err != nil {
		t.Fatalf("unexpected error (should return false, not error): %v", err)
	}
	if valid {
		t.Error("expected commit to be invalid")
	}
}

func TestGetMergeBase_Success(t *testing.T) {
	expectedSHA := "abc123def456"
	mock := &mockCommander{
		runFunc: func(name string, args ...string) ([]byte, error) {
			if name != "git" {
				t.Errorf("expected command 'git', got '%s'", name)
			}
			expectedArgs := []string{"merge-base", "HEAD", "main"}
			if len(args) != len(expectedArgs) {
				t.Errorf("expected %d args, got %d", len(expectedArgs), len(args))
			}
			for i, arg := range expectedArgs {
				if i < len(args) && args[i] != arg {
					t.Errorf("arg %d: expected '%s', got '%s'", i, arg, args[i])
				}
			}
			return []byte(expectedSHA + "\n"), nil
		},
	}

	svc := New(mock)
	sha, err := svc.GetMergeBase("main")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sha != expectedSHA {
		t.Errorf("expected SHA '%s', got '%s'", expectedSHA, sha)
	}
}

func TestGetMergeBase_Error(t *testing.T) {
	expectedErr := errors.New("branches have no common ancestor")
	mock := &mockCommander{
		runFunc: func(name string, args ...string) ([]byte, error) {
			return nil, expectedErr
		},
	}

	svc := New(mock)
	_, err := svc.GetMergeBase("main")

	if err != expectedErr {
		t.Errorf("expected error '%v', got '%v'", expectedErr, err)
	}
}

func TestGetMergeBase_TrimWhitespace(t *testing.T) {
	expectedSHA := "abc123def456"
	mock := &mockCommander{
		runFunc: func(name string, args ...string) ([]byte, error) {
			return []byte("  " + expectedSHA + "  \n\n"), nil
		},
	}

	svc := New(mock)
	sha, err := svc.GetMergeBase("main")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sha != expectedSHA {
		t.Errorf("expected SHA '%s', got '%s'", expectedSHA, sha)
	}
}

func TestSplitLines_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []string
	}{
		{
			name:     "empty input",
			input:    []byte(""),
			expected: []string{},
		},
		{
			name:     "only whitespace",
			input:    []byte("   \n  \n  "),
			expected: []string{},
		},
		{
			name:     "single line no newline",
			input:    []byte("file.go"),
			expected: []string{"file.go"},
		},
		{
			name:     "mixed whitespace",
			input:    []byte("\nfile1.go\n  \n\nfile2.go\n\n"),
			expected: []string{"file1.go", "file2.go"},
		},
		{
			name:     "lines with spaces",
			input:    []byte("  file1.go  \n  file2.go  "),
			expected: []string{"file1.go", "file2.go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitLines(tt.input)
			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d lines, got %d: %v", len(tt.expected), len(result), result)
			}
			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("line %d: expected '%s', got '%s'", i, expected, result[i])
				}
			}
		})
	}
}
