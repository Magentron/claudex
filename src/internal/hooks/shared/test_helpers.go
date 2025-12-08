package shared

import "claudex/internal/notify"

// mockEnv implements Environment for testing
type mockEnv struct {
	vars map[string]string
}

func (m *mockEnv) Get(key string) string {
	if m.vars == nil {
		return ""
	}
	return m.vars[key]
}

func (m *mockEnv) Set(key, value string) {
	if m.vars == nil {
		m.vars = make(map[string]string)
	}
	m.vars[key] = value
}

// MockEnv is an alias for mockEnv for compatibility with existing tests
type MockEnv = mockEnv

// NewMockEnv creates a new MockEnv instance
func NewMockEnv() *MockEnv {
	return &MockEnv{
		vars: make(map[string]string),
	}
}

// mockNotifier implements notify.Notifier for testing
type mockNotifier struct {
	sendCalls  []sendCall
	speakCalls []string
	sendErr    error
	speakErr   error
	available  bool
}

type sendCall struct {
	title   string
	message string
	sound   string
}

func (m *mockNotifier) Send(title, message, sound string) error {
	m.sendCalls = append(m.sendCalls, sendCall{
		title:   title,
		message: message,
		sound:   sound,
	})
	return m.sendErr
}

func (m *mockNotifier) Speak(message string) error {
	m.speakCalls = append(m.speakCalls, message)
	return m.speakErr
}

func (m *mockNotifier) IsAvailable() bool {
	return m.available
}

// Ensure mockNotifier implements notify.Notifier
var _ notify.Notifier = (*mockNotifier)(nil)
