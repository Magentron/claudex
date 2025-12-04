package commander

import (
	"io"
	"os/exec"
)

// Commander abstracts process execution for testability
type Commander interface {
	// Run executes command and returns combined output
	Run(name string, args ...string) ([]byte, error)
	// Start launches interactive command with stdio attached
	Start(name string, stdin io.Reader, stdout, stderr io.Writer, args ...string) error
}

// OsCommander is the production implementation of Commander
type OsCommander struct{}

func (c *OsCommander) Run(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).CombinedOutput()
}

func (c *OsCommander) Start(name string, stdin io.Reader, stdout, stderr io.Writer, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

// New creates a new Commander instance
func New() Commander {
	return &OsCommander{}
}
