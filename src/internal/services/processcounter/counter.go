package processcounter

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// ProcessCounter provides methods to count process tree descendants.
type ProcessCounter interface {
	// CountDescendants counts all descendants (direct children and their descendants)
	// of a process identified by pid. Returns 0 if the process has no descendants.
	CountDescendants(pid int) (int, error)
}

// linuxCounter implements ProcessCounter using the Linux /proc filesystem.
type linuxCounter struct{}

// CountDescendants counts descendants by reading /proc/<pid>/task/<tid>/children recursively.
func (c *linuxCounter) CountDescendants(pid int) (int, error) {
	children, err := c.getDirectChildren(pid)
	if err != nil {
		return 0, err
	}

	count := len(children)
	for _, childPID := range children {
		descendants, err := c.CountDescendants(childPID)
		if err != nil {
			// Ignore errors for child processes (they may have exited)
			continue
		}
		count += descendants
	}

	return count, nil
}

// getDirectChildren reads /proc/<pid>/task/<tid>/children to get direct child PIDs.
func (c *linuxCounter) getDirectChildren(pid int) ([]int, error) {
	// First, try reading the main process's children file
	childrenPath := filepath.Join("/proc", strconv.Itoa(pid), "task", strconv.Itoa(pid), "children")
	children, err := c.readChildrenFile(childrenPath)
	if err == nil && len(children) > 0 {
		return children, nil
	}

	// If that fails or is empty, scan all tasks for this process
	taskDir := filepath.Join("/proc", strconv.Itoa(pid), "task")
	taskEntries, err := os.ReadDir(taskDir)
	if err != nil {
		// Process may have exited or we don't have permissions
		return nil, nil
	}

	allChildren := make(map[int]bool)
	for _, entry := range taskEntries {
		if !entry.IsDir() {
			continue
		}

		tid := entry.Name()
		childrenPath := filepath.Join(taskDir, tid, "children")
		children, err := c.readChildrenFile(childrenPath)
		if err != nil {
			continue
		}

		for _, child := range children {
			allChildren[child] = true
		}
	}

	result := make([]int, 0, len(allChildren))
	for child := range allChildren {
		result = append(result, child)
	}

	return result, nil
}

// readChildrenFile reads a /proc/<pid>/task/<tid>/children file and parses space-separated PIDs.
func (c *linuxCounter) readChildrenFile(path string) ([]int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := strings.TrimSpace(string(data))
	if content == "" {
		return nil, nil
	}

	parts := strings.Fields(content)
	pids := make([]int, 0, len(parts))
	for _, part := range parts {
		pid, err := strconv.Atoi(part)
		if err != nil {
			continue
		}
		pids = append(pids, pid)
	}

	return pids, nil
}

// darwinCounter implements ProcessCounter using pgrep for macOS.
type darwinCounter struct{}

// CountDescendants counts descendants by using pgrep -P recursively.
func (c *darwinCounter) CountDescendants(pid int) (int, error) {
	children, err := c.getDirectChildren(pid)
	if err != nil {
		return 0, err
	}

	count := len(children)
	for _, childPID := range children {
		descendants, err := c.CountDescendants(childPID)
		if err != nil {
			// Ignore errors for child processes (they may have exited)
			continue
		}
		count += descendants
	}

	return count, nil
}

// getDirectChildren uses pgrep -P <pid> to get direct child PIDs.
func (c *darwinCounter) getDirectChildren(pid int) ([]int, error) {
	cmd := exec.Command("pgrep", "-P", strconv.Itoa(pid))
	output, err := cmd.Output()
	if err != nil {
		// pgrep returns exit code 1 when no children found - this is not an error
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil, nil
		}
		return nil, fmt.Errorf("pgrep failed: %w", err)
	}

	content := strings.TrimSpace(string(output))
	if content == "" {
		return nil, nil
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	var pids []int
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		pid, err := strconv.Atoi(line)
		if err != nil {
			continue
		}
		pids = append(pids, pid)
	}

	return pids, nil
}

// NewProcessCounter creates a platform-specific ProcessCounter implementation.
func NewProcessCounter() ProcessCounter {
	switch runtime.GOOS {
	case "linux":
		return &linuxCounter{}
	case "darwin":
		return &darwinCounter{}
	default:
		// Fallback to Linux implementation for other Unix-like systems
		return &linuxCounter{}
	}
}
