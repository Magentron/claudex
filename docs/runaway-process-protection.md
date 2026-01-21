# Runaway Process Protection

Claudex includes comprehensive multi-layered protection against runaway process spawning. This document explains how the protection works, how to configure it, and how to troubleshoot issues.

## Table of Contents

- [Overview](#overview)
- [Protection Layers](#protection-layers)
- [Configuration](#configuration)
- [Environment Variables](#environment-variables)
- [How It Works](#how-it-works)
- [Troubleshooting](#troubleshooting)
- [Enabling Cgroups for Non-Root Users](#enabling-cgroups-for-non-root-users)
- [Cgroups in Containers](#cgroups-in-containers)
- [Platform Differences](#platform-differences)
- [Security Considerations](#security-considerations)

## Overview

When Claude Code generates and executes commands, there's a risk of runaway processes - malicious or buggy code that spawns processes recursively until the system becomes unresponsive. Claudex protects against this with four complementary layers:

1. **Application-level tracking** - Monitors spawned processes
2. **Rate limiting** - Prevents rapid spawning
3. **Cgroups v2 PID limiting** - Per-process-tree hard limits (Linux)
4. **Process groups** - Clean process isolation and termination

Default settings provide protection out of the box while allowing normal development workflows.

## Protection Layers

### 1. Application-Level Process Tracking

Claudex maintains a global registry of all spawned child processes and their descendants. Before spawning a new process, it counts active descendants and blocks execution if the limit is reached.

**Key features:**
- Thread-safe PID tracking with mutex protection
- Recursive descendant counting via `/proc` (Linux) or `pgrep` (macOS)
- Automatic cleanup when processes exit
- Configurable limit (default: 2 × CPU cores)

**Error message when triggered:**
```
Error: Process limit reached (50/50 active processes)
Cannot spawn new process. Increase CLAUDEX_MAX_PROCESSES or wait for processes to complete.
```

### 2. Rate Limiting

Prevents rapid process spawning with exponential backoff. Tracks spawn timestamps in a sliding window and enforces frequency limits.

**Behavior:**
- Default limit: 5 spawns per second
- Exponential backoff: 100ms, 200ms, 400ms, 800ms, 1600ms (capped at 3s)
- Automatic cooldown when spawn rate decreases

**Example:**
```
Spawn attempt #6 within 1 second → 100ms delay
Spawn attempt #7 within 1 second → 200ms delay
Spawn attempt #8 within 1 second → 400ms delay
```

### 3. Cgroups v2 PID Limiting (Linux only)

On Linux, Claudex uses cgroups v2 with the `pids` controller to enforce **true per-process-tree limits**. Each spawned command is placed in its own cgroup with a PID limit, preventing it (and all its descendants) from exceeding the configured maximum.

**Key features:**
- Limits apply to the entire process tree (children, grandchildren, etc.)
- Kernel-enforced - child processes cannot bypass by forking
- Uses same limit as `max_processes` configuration
- Automatic cgroup cleanup when processes exit
- Linux only (gracefully ignored on macOS)

**How it works:**
1. When a command is spawned, claudex creates a cgroup at `/sys/fs/cgroup/claudex/cmd_<pid>/`
2. Sets `pids.max` to the configured `max_processes` limit
3. Moves the child process into the cgroup
4. All descendants inherit the cgroup and its limits
5. Cleans up the cgroup when the process exits

**Verification:**
```bash
# Check cgroup of a claudex child process
cat /proc/<pid>/cgroup
# 0::/claudex/cmd_12345

# Check the PID limit
cat /sys/fs/cgroup/claudex/cmd_12345/pids.max
# 16

# Check current PID count
cat /sys/fs/cgroup/claudex/cmd_12345/pids.current
# 3
```

**Requirements:**
- Cgroups v2 unified hierarchy (default on modern Linux)
- Write access to `/sys/fs/cgroup/claudex/` (may require systemd delegation or root)

### 4. Process Groups

All child processes are isolated in their own process group using `Setpgid`. This enables:
- Clean signal handling (terminate entire process tree)
- Prevention of zombie processes
- Isolation from parent process group

## Configuration

### .claudex/config.toml

Add a `[features.process_protection]` section to customize protection settings:

```toml
[features.process_protection]
# Application-level limit and cgroups PID limit (default: 2 × CPU cores)
# Set to 0 to disable process limiting
max_processes = 16  # Example: override the dynamic default

# Max spawn frequency per second (default: 5)
# Spawns exceeding this rate trigger exponential backoff
rate_limit_per_second = 5

# Per-process timeout in seconds (default: 300 = 5 minutes)
# Set to 0 to disable timeouts
timeout_seconds = 300
```

### Default Values

If not specified in config, these defaults apply:

| Setting | Default | Description |
|---------|---------|-------------|
| `max_processes` | 2 × CPU cores | Application-level and cgroups PID limit |
| `rate_limit_per_second` | 5 | Max spawns per second before backoff |
| `timeout_seconds` | 300 | Process timeout (5 minutes) |

## Environment Variables

Environment variables override config file settings. Useful for one-off adjustments without editing config.

### CLAUDEX_MAX_PROCESSES

Override the application-level process limit.

```bash
# Disable application-level limit
CLAUDEX_MAX_PROCESSES=0 claudex

# Increase limit for parallel workloads
CLAUDEX_MAX_PROCESSES=200 claudex

# Decrease limit for conservative protection
CLAUDEX_MAX_PROCESSES=10 claudex
```

### CLAUDEX_RATE_LIMIT

Override the rate limit (spawns per second).

```bash
# More aggressive rate limiting
CLAUDEX_RATE_LIMIT=2 claudex

# More relaxed rate limiting
CLAUDEX_RATE_LIMIT=20 claudex

# Disable rate limiting
CLAUDEX_RATE_LIMIT=0 claudex
```

### CLAUDEX_TIMEOUT

Override the per-process timeout (seconds).

```bash
# 10 minute timeout
CLAUDEX_TIMEOUT=600 claudex

# 1 hour timeout
CLAUDEX_TIMEOUT=3600 claudex

# Disable timeouts
CLAUDEX_TIMEOUT=0 claudex
```

### Combining Overrides

Multiple environment variables can be combined:

```bash
# High-volume parallel processing configuration
CLAUDEX_MAX_PROCESSES=500 \
CLAUDEX_RATE_LIMIT=50 \
CLAUDEX_TIMEOUT=1800 \
claudex
```

## How It Works

### Process Lifecycle

```
┌─────────────────────────────────────────────────────────────────┐
│  1. Command execution requested                                  │
│       ↓                                                          │
│  2. Pre-flight checks:                                           │
│       - Count active descendants (via /proc or pgrep)            │
│       - Check against max_processes limit                        │
│       - Check rate limiting (last N spawn timestamps)            │
│       ↓                                                          │
│  3. If checks pass, spawn process with:                          │
│       - Process group isolation (Setpgid)                        │
│       - Optional timeout via context.Context                     │
│       ↓                                                          │
│  4. Register PID in global ProcessRegistry                       │
│       ↓                                                          │
│  5. Create cgroup and apply PID limit (Linux)                    │
│       ↓                                                          │
│  6. Process executes                                             │
│       ↓                                                          │
│  7. Process exits (or times out)                                 │
│       ↓                                                          │
│  8. Unregister PID from ProcessRegistry                          │
│       ↓                                                          │
│  9. Cleanup cgroup (Linux)                                       │
│       ↓                                                          │
│  10. Cleanup complete                                            │
└─────────────────────────────────────────────────────────────────┘
```

### Descendant Counting

**Linux** (`/proc` filesystem):
1. Read `/proc/<pid>/task/<tid>/children` for direct child PIDs
2. Recursively count descendants of each child
3. Return total count

**macOS** (`pgrep` fallback):
1. Execute `pgrep -P <pid>` to find child PIDs
2. Parse output line-by-line
3. Recursively count descendants of each child
4. Return total count

### Rate Limiting Algorithm

Claudex uses a sliding window with exponential backoff:

1. Track last N spawn timestamps in a ring buffer
2. On spawn attempt, check if last `rate_limit_per_second` spawns occurred within 1 second
3. If yes, calculate backoff: `100ms * 2^(violations - 1)` (capped at 3s)
4. Sleep for backoff duration, then retry
5. If spawn rate decreases, reset backoff multiplier

## Troubleshooting

### "Process limit reached" Error

**Symptom:**
```
Error: Process limit reached (50/50 active processes)
Cannot spawn new process.
```

**Causes:**
- Legitimate parallel workload exceeding limit
- Orphaned processes not being cleaned up
- Long-running background processes

**Solutions:**

1. **Increase limit temporarily:**
   ```bash
   CLAUDEX_MAX_PROCESSES=100 claudex
   ```

2. **Increase limit permanently:**
   ```toml
   [features.process_protection]
   max_processes = 100
   ```

3. **Disable limit entirely:**
   ```bash
   CLAUDEX_MAX_PROCESSES=0 claudex
   ```

4. **Check for orphaned processes:**
   ```bash
   ps aux | grep claudex
   # Kill orphaned processes if found
   pkill -f claudex
   ```

### Rate Limiting Too Aggressive

**Symptom:**
Noticeable delays between command executions, messages about spawn rate limiting.

**Solutions:**

1. **Increase rate limit temporarily:**
   ```bash
   CLAUDEX_RATE_LIMIT=20 claudex
   ```

2. **Increase rate limit permanently:**
   ```toml
   [features.process_protection]
   rate_limit_per_second = 20
   ```

3. **Disable rate limiting:**
   ```bash
   CLAUDEX_RATE_LIMIT=0 claudex
   ```

### Processes Timing Out

**Symptom:**
Long-running commands being killed with "context deadline exceeded" errors.

**Solutions:**

1. **Increase timeout temporarily:**
   ```bash
   CLAUDEX_TIMEOUT=1800 claudex  # 30 minutes
   ```

2. **Increase timeout permanently:**
   ```toml
   [features.process_protection]
   timeout_seconds = 1800
   ```

3. **Disable timeouts:**
   ```bash
   CLAUDEX_TIMEOUT=0 claudex
   ```

### Cgroups PID Limit Issues (Linux)

**Symptom:**
"fork: resource temporarily unavailable" errors when child processes try to spawn.

**Causes:**
- Child process tree exceeded `max_processes` limit
- Cgroups PID controller working as intended

**Solutions:**

1. **Increase the limit:**
   ```bash
   CLAUDEX_MAX_PROCESSES=100 claudex
   ```

2. **Disable cgroups limiting:**
   ```bash
   CLAUDEX_MAX_PROCESSES=0 claudex
   ```

### Cgroups Not Available

**Symptom:**
Cgroups limiting is silently disabled (check logs or test with fork bomb).

**Causes:**
- Cgroups v2 not available (older kernel or cgroups v1 only)
- No write permission to `/sys/fs/cgroup/`
- `pids` controller not enabled

**Solutions:**

1. **Check cgroups v2 availability:**
   ```bash
   cat /sys/fs/cgroup/cgroup.controllers
   # Should include "pids"
   ```

2. **Enable systemd delegation for user cgroups** (see [Enabling Cgroups for Non-Root Users](#enabling-cgroups-for-non-root-users))

3. **Run claudex in a container** with cgroup access (see [Cgroups in Containers](#cgroups-in-containers))

## Enabling Cgroups for Non-Root Users

By default, cgroups require root privileges. However, systemd can delegate cgroup control to regular users. This section explains how to enable cgroups-based process limiting without root access.

### Check Current Status

First, check if cgroups v2 is available and what controllers you have access to:

```bash
# Check if cgroups v2 is mounted
mount | grep cgroup2
# Should show: cgroup2 on /sys/fs/cgroup type cgroup2 ...

# Check available controllers at the system level
cat /sys/fs/cgroup/cgroup.controllers
# Should include: cpuset cpu io memory hugetlb pids rdma misc

# Check if you can write to the cgroup filesystem
touch /sys/fs/cgroup/claudex-test 2>&1 && rm /sys/fs/cgroup/claudex-test
# If this fails with "Permission denied", you need delegation
```

### Method 1: Systemd User Service Delegation (Recommended)

This is the cleanest approach for desktop Linux systems with systemd.

**Step 1: Create systemd override for user sessions**

```bash
# Create the override directory
sudo mkdir -p /etc/systemd/system/user@.service.d

# Create the delegation config
sudo tee /etc/systemd/system/user@.service.d/delegate.conf << 'EOF'
[Service]
Delegate=pids cpu memory
EOF

# Reload systemd
sudo systemctl daemon-reload
```

**Step 2: Re-login or restart your user session**

```bash
# Option 1: Logout and login again
# Option 2: Restart the user manager
systemctl restart user@$(id -u).service
```

**Step 3: Verify delegation is working**

```bash
# Check your user cgroup path
cat /proc/self/cgroup
# Should show something like: 0::/user.slice/user-1000.slice/user@1000.service/...

# Check delegated controllers in your user slice
cat /sys/fs/cgroup/user.slice/user-$(id -u).slice/cgroup.controllers
# Should include: pids

# Verify you can create cgroups
mkdir /sys/fs/cgroup/user.slice/user-$(id -u).slice/user@$(id -u).service/claudex-test
rmdir /sys/fs/cgroup/user.slice/user-$(id -u).slice/user@$(id -u).service/claudex-test
```

### Method 2: Running Claudex via Systemd User Service

You can run claudex as a systemd user service to automatically get cgroup delegation:

```bash
# Create user service directory
mkdir -p ~/.config/systemd/user

# Create the service file
cat > ~/.config/systemd/user/claudex.service << 'EOF'
[Unit]
Description=Claudex Development Assistant

[Service]
Type=simple
ExecStart=/usr/local/bin/claudex
Delegate=pids
StandardInput=tty
StandardOutput=tty
StandardError=tty
TTYPath=/dev/pts/0
TTYReset=yes

[Install]
WantedBy=default.target
EOF

# Reload user services
systemctl --user daemon-reload
```

### Method 3: Using sudo with Preserved Environment

For occasional use, you can run claudex with sudo while preserving your environment:

```bash
# Run with preserved HOME and USER
sudo -E claudex

# Or preserve specific variables
sudo HOME=$HOME USER=$USER claudex
```

**Note:** This grants root privileges to all spawned processes, which may not be desirable.

## Cgroups in Containers

Containers typically have cgroup access if properly configured. Here's how to enable it for Docker and Podman.

### Docker

**Option 1: Bind-mount the cgroup filesystem (Recommended)**

```bash
docker run -it \
  --mount type=bind,source=/sys/fs/cgroup,target=/sys/fs/cgroup,readonly=false \
  --cgroupns=host \
  your-image claudex
```

**Option 2: Using Docker Compose**

```yaml
version: '3.8'
services:
  claudex:
    image: your-image
    stdin_open: true
    tty: true
    volumes:
      - /sys/fs/cgroup:/sys/fs/cgroup:rw
    cgroup_parent: claudex
    # For Docker 20.10+
    cgroupns_mode: host
```

**Option 3: Privileged mode (Not recommended for production)**

```bash
docker run -it --privileged your-image claudex
```

### Podman

Podman has better rootless container support with cgroups v2.

**Rootless Podman (Recommended)**

```bash
# Podman automatically delegates cgroups to rootless containers on cgroups v2
podman run -it \
  --cgroups=split \
  your-image claudex
```

**With explicit cgroup mount**

```bash
podman run -it \
  -v /sys/fs/cgroup:/sys/fs/cgroup:rw \
  --cgroup-parent=claudex \
  your-image claudex
```

**Using Podman Compose**

```yaml
version: '3.8'
services:
  claudex:
    image: your-image
    stdin_open: true
    tty: true
    volumes:
      - /sys/fs/cgroup:/sys/fs/cgroup:rw
    cgroup: split
```

### Kubernetes

For Kubernetes deployments, cgroup access requires security context configuration:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: claudex
spec:
  containers:
  - name: claudex
    image: your-image
    command: ["claudex"]
    securityContext:
      privileged: false
      capabilities:
        add:
          - SYS_ADMIN  # Required for cgroup management
    volumeMounts:
    - name: cgroup
      mountPath: /sys/fs/cgroup
  volumes:
  - name: cgroup
    hostPath:
      path: /sys/fs/cgroup
      type: Directory
```

**Note:** Many Kubernetes environments restrict cgroup access. Check with your cluster administrator.

### Verifying Cgroup Access in Containers

After starting your container, verify cgroup access:

```bash
# Check cgroups v2 is available
cat /sys/fs/cgroup/cgroup.controllers
# Should include: pids

# Check if claudex cgroup can be created
mkdir -p /sys/fs/cgroup/claudex && rmdir /sys/fs/cgroup/claudex
# Should succeed without errors

# Run claudex and check if cgroup limiting is active
claudex --version  # Start claudex
# Then in another terminal:
ls /sys/fs/cgroup/claudex/
# Should show cmd_<pid> directories when processes are running
```

### Troubleshooting Container Cgroup Issues

**"Permission denied" when creating cgroups**

```bash
# Check if cgroup filesystem is mounted read-only
mount | grep cgroup
# If "ro" is shown, remount or adjust container config

# For Docker, ensure cgroupns=host is set
docker run --cgroupns=host ...
```

**"No such file or directory" for cgroup.controllers**

```bash
# The host may be using cgroups v1
# Check host cgroup version:
stat -fc %T /sys/fs/cgroup/
# "cgroup2fs" = v2, "tmpfs" = v1

# If using cgroups v1, cgroup PID limiting is not available
# Claudex will fall back to application-level protection
```

**Container exits with "OCI runtime error"**

```bash
# SELinux may be blocking cgroup access
# Check SELinux status
getenforce

# Temporarily set to permissive for testing
sudo setenforce 0

# For permanent fix, add SELinux policy or use :z/:Z volume options
docker run -v /sys/fs/cgroup:/sys/fs/cgroup:rw,z ...
```

### Verifying Protection is Active

**Check application-level tracking:**
```bash
# Spawn multiple processes and check limit
for i in {1..60}; do sleep 100 & done
# Should block around spawn #50
```

**Check cgroups PID limit (Linux):**
```bash
# List active claudex cgroups
ls /sys/fs/cgroup/claudex/

# Check a specific cgroup's PID limit
cat /sys/fs/cgroup/claudex/cmd_<pid>/pids.max

# Check current PID count in cgroup
cat /sys/fs/cgroup/claudex/cmd_<pid>/pids.current
```

**Check rate limiting:**
```bash
# Spawn processes rapidly
time (for i in {1..20}; do echo "test" & done)
# Should see delays due to rate limiting
```

## Platform Differences

### Linux

**Full protection:**
- Application-level tracking via `/proc` filesystem
- Cgroups v2 per-process-tree PID limiting
- Rate limiting
- Process groups

**Best accuracy:**
- `/proc/<pid>/task/<tid>/children` provides exact descendant counts
- Cgroups PID controller provides true per-process-tree isolation
- Child processes cannot bypass limits by forking

### macOS

**Limited protection:**
- Application-level tracking via `pgrep` command
- No cgroups support (gracefully ignored)
- Rate limiting
- Process groups

**Known limitations:**
- `pgrep` has slight delays and may miss rapid spawns
- No kernel-level backstop for child process forking
- Relies entirely on application-level checks
- A child process can spawn unlimited grandchildren

### Recommendation

For production environments or sensitive workloads, **Linux is recommended** due to cgroups enforcement and more accurate process tracking.

## Security Considerations

### Cgroups Permissions

Cgroups v2 PID limiting requires write access to `/sys/fs/cgroup/`. There are several ways to achieve this:

1. **Systemd user delegation** (recommended): Configure systemd to delegate cgroup control to user sessions
2. **Container environments**: Run claudex in a container with cgroup access
3. **Root privileges**: Not recommended for general use

If cgroups are not available, claudex falls back to application-level protection only.

### Race Conditions

Process counting via `/proc` or `pgrep` may race with process exits. This is acceptable - counts may be slightly stale but protection is still effective:
- Worst case: limit is exceeded by 1-2 processes temporarily
- Application-level tracking remains accurate
- Rate limiting provides additional safety

### Bypass Attempts

**Application-level tracking can be bypassed if:**
- Child processes use `exec` to replace themselves (tracked PID changes)
- Child processes daemonize and detach from parent

**Defense in depth:**
- Cgroups PID limit (Linux) provides kernel-enforced per-process-tree limits
- Cgroup membership is inherited by all descendants - cannot be escaped
- Rate limiting prevents rapid exploitation
- Process groups enable clean termination

### Resource Exhaustion

**Not protected against:**
- Slow accumulation of long-running processes under the limit
- CPU or memory exhaustion (separate resource limits needed)
- Disk I/O saturation

**Recommended additional protections:**
- System-level resource limits (`ulimit`)
- Container resource constraints (Docker, etc)
- OS-level monitoring and alerts

## Summary

Claudex's runaway process protection provides robust defense against runaway process spawning while remaining configurable for legitimate high-volume workloads. The multi-layered approach ensures protection even if one layer fails.

**Quick reference:**

| Protection Layer | Platform | Bypass Difficulty | Configuration |
|------------------|----------|-------------------|---------------|
| Application tracking | All | Medium | `max_processes` |
| Rate limiting | All | Easy | `rate_limit_per_second` |
| Cgroups PID limit | Linux only | Very Hard | `max_processes` |
| Process groups | All | Hard | Always enabled |

For most users, **default settings are sufficient**. Adjust only if you encounter legitimate workload blocks.
