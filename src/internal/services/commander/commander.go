package commander

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"syscall"

	"claudex/internal/services/cgroup"
	"claudex/internal/services/config"
	"claudex/internal/services/processcounter"
	"claudex/internal/services/processregistry"
	"claudex/internal/services/ratelimit"
	"github.com/spf13/afero"
)

// Commander abstracts process execution for testability
type Commander interface {
	// Run executes command and returns combined output
	Run(name string, args ...string) ([]byte, error)
	// Start launches interactive command with stdio attached
	Start(name string, stdin io.Reader, stdout, stderr io.Writer, args ...string) error
	// RunWithContext executes command with context support for timeout and cancellation
	RunWithContext(ctx context.Context, name string, args ...string) ([]byte, error)
	// StartWithContext launches interactive command with context support
	StartWithContext(ctx context.Context, name string, stdin io.Reader, stdout, stderr io.Writer, args ...string) error
}

// OsCommander is the production implementation of Commander with runaway process protection
type OsCommander struct {
	fs          afero.Fs
	cfg         *config.Config
	rateLimiter *ratelimit.RateLimiter
	counter     processcounter.ProcessCounter
	registry    processregistry.ProcessRegistry
	cgroupPIDs  *cgroup.PIDLimiter
}

// New creates a new Commander instance with default dependencies
func New() Commander {
	return NewWithDeps(afero.NewOsFs(), nil)
}

// NewWithDeps creates a new Commander instance with injected dependencies (for testing)
func NewWithDeps(fs afero.Fs, cfg *config.Config) Commander {
	if cfg == nil {
		// Load config with defaults if not provided
		loadedCfg, err := config.Load(fs, ".claudex/config.toml")
		if err != nil {
			// Use defaults if config load fails
			loadedCfg = &config.Config{
				Features: config.Features{
					ProcessProtection: config.ProcessProtection{
						MaxProcesses:       runtime.NumCPU() * 2,
						RateLimitPerSecond: 5,
						TimeoutSeconds:     300,
					},
				},
			}
		}
		cfg = loadedCfg
	}

	protection := cfg.Features.ProcessProtection

	return &OsCommander{
		fs:          fs,
		cfg:         cfg,
		rateLimiter: ratelimit.NewRateLimiter(protection.RateLimitPerSecond),
		counter:     processcounter.NewProcessCounter(),
		registry:    processregistry.DefaultRegistry,
		cgroupPIDs:  cgroup.NewPIDLimiter(protection.MaxProcesses),
	}
}

// Run executes command and returns combined output (backward compatible)
func (c *OsCommander) Run(name string, args ...string) ([]byte, error) {
	return c.RunWithContext(context.Background(), name, args...)
}

// Start launches interactive command with stdio attached (backward compatible)
func (c *OsCommander) Start(name string, stdin io.Reader, stdout, stderr io.Writer, args ...string) error {
	return c.StartWithContext(context.Background(), name, stdin, stdout, stderr, args...)
}

// RunWithContext executes command with context support and runaway process protection
func (c *OsCommander) RunWithContext(ctx context.Context, name string, args ...string) ([]byte, error) {
	if err := c.preflight(); err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, name, args...)
	c.applySysProcAttr(cmd)

	// Capture combined output using bytes.Buffer
	var outputBuf bytes.Buffer
	cmd.Stdout = &outputBuf
	cmd.Stderr = &outputBuf

	// Start the command and register PID
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	pid := cmd.Process.Pid
	c.registry.Register(pid)
	defer c.registry.Unregister(pid)

	// Apply cgroup PID limit (Linux only, no-op on other platforms)
	cgroupPath, _ := c.cgroupPIDs.CreateForProcess(pid)
	if cgroupPath != "" {
		defer c.cgroupPIDs.Cleanup(cgroupPath)
	}

	// Wait for the command to finish
	err := cmd.Wait()
	return outputBuf.Bytes(), err
}

// StartWithContext launches interactive command with context support and runaway process protection.
// This method blocks until the command completes (like the original Start).
func (c *OsCommander) StartWithContext(ctx context.Context, name string, stdin io.Reader, stdout, stderr io.Writer, args ...string) error {
	if err := c.preflight(); err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	c.applySysProcAttr(cmd)

	if err := cmd.Start(); err != nil {
		return err
	}

	pid := cmd.Process.Pid
	c.registry.Register(pid)
	defer c.registry.Unregister(pid)

	// Apply cgroup PID limit (Linux only, no-op on other platforms)
	cgroupPath, _ := c.cgroupPIDs.CreateForProcess(pid)
	if cgroupPath != "" {
		defer c.cgroupPIDs.Cleanup(cgroupPath)
	}

	// Wait for the command to complete (synchronous, like the original cmd.Run())
	return cmd.Wait()
}

// preflight performs pre-execution checks for runaway process protection
func (c *OsCommander) preflight() error {
	protection := c.cfg.Features.ProcessProtection

	// Check rate limiting
	if protection.RateLimitPerSecond > 0 {
		c.rateLimiter.Allow() // Blocks with exponential backoff if needed
	}

	// Check process limit
	if protection.MaxProcesses > 0 {
		currentPID := os.Getpid()
		descendants := 0

		// Try to count descendants, but don't fail if we can't
		count, err := c.counter.CountDescendants(currentPID)
		if err == nil {
			descendants = count
		}

		registeredCount := c.registry.Count()
		total := descendants + registeredCount

		if total >= protection.MaxProcesses {
			return fmt.Errorf("runaway process protection: process limit reached (%d/%d total processes)", total, protection.MaxProcesses)
		}
	}

	return nil
}

// applySysProcAttr applies process isolation
func (c *OsCommander) applySysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // Create new process group for isolation
	}
}
