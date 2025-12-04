package testutil

import (
	"io"
	"strings"
)

// CommandInvocation represents a captured command execution
type CommandInvocation struct {
	Name  string
	Args  []string
	Stdin string
}

// MockCommander captures command invocations for verification
type MockCommander struct {
	Invocations []CommandInvocation
	responses   map[string]mockResponse
	patterns    []patternMapping
}

type mockResponse struct {
	output []byte
	err    error
}

type patternMapping struct {
	name       string
	argPattern []string
	response   mockResponse
}

type patternBuilder struct {
	mock       *MockCommander
	name       string
	argPattern []string
}

// NewMockCommander creates a new mock commander
func NewMockCommander() *MockCommander {
	return &MockCommander{
		responses: make(map[string]mockResponse),
		patterns:  []patternMapping{},
	}
}

// OnPattern sets up a pattern-based response
func (m *MockCommander) OnPattern(name string, argPattern ...string) *patternBuilder {
	return &patternBuilder{
		mock:       m,
		name:       name,
		argPattern: argPattern,
	}
}

// Return sets the response for the pattern
func (pb *patternBuilder) Return(output []byte, err error) {
	pb.mock.patterns = append(pb.mock.patterns, patternMapping{
		name:       pb.name,
		argPattern: pb.argPattern,
		response: mockResponse{
			output: output,
			err:    err,
		},
	})
}

// Run executes a command and captures the invocation
func (m *MockCommander) Run(name string, args ...string) ([]byte, error) {
	m.Invocations = append(m.Invocations, CommandInvocation{
		Name: name,
		Args: args,
	})

	// Match against patterns
	for _, pattern := range m.patterns {
		if pattern.name == name && matchesPattern(args, pattern.argPattern) {
			return pattern.response.output, pattern.response.err
		}
	}

	return nil, nil
}

// Start executes an interactive command and captures the invocation
func (m *MockCommander) Start(name string, stdin io.Reader, stdout, stderr io.Writer, args ...string) error {
	stdinContent := ""
	if stdin != nil {
		data, _ := io.ReadAll(stdin)
		stdinContent = string(data)
	}

	m.Invocations = append(m.Invocations, CommandInvocation{
		Name:  name,
		Args:  args,
		Stdin: stdinContent,
	})

	// Match against patterns and write output
	for _, pattern := range m.patterns {
		if pattern.name == name && matchesPattern(args, pattern.argPattern) {
			if stdout != nil && pattern.response.output != nil {
				stdout.Write(pattern.response.output)
			}
			return pattern.response.err
		}
	}

	return nil
}

// LastInvocation returns the most recent command invocation
func (m *MockCommander) LastInvocation() CommandInvocation {
	if len(m.Invocations) == 0 {
		return CommandInvocation{}
	}
	return m.Invocations[len(m.Invocations)-1]
}

// matchesPattern checks if args match the pattern
func matchesPattern(args, pattern []string) bool {
	if len(pattern) == 0 {
		return true
	}

	for _, p := range pattern {
		found := false
		for _, arg := range args {
			if strings.Contains(arg, p) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// MockEnv provides mock environment variable access
type MockEnv struct {
	vars map[string]string
}

// NewMockEnv creates a new mock environment
func NewMockEnv() *MockEnv {
	return &MockEnv{
		vars: make(map[string]string),
	}
}

// Get retrieves an environment variable
func (e *MockEnv) Get(key string) string {
	return e.vars[key]
}

// Set sets an environment variable
func (e *MockEnv) Set(key, value string) {
	e.vars[key] = value
}
