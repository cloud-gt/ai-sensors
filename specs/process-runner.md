# Spec: Process Runner

## Purpose
Spawn external processes, capture their output, and manage their lifecycle with proper signal handling and resource cleanup.

## Rationale
The system needs to execute arbitrary shell commands (sensors) and capture their output for storage and analysis. The Process Runner provides a clean abstraction over Go's `os/exec` package, handling the complexity of process lifecycle, signal propagation, and output streaming while exposing a simple state machine to callers.

## Package
- **Location:** `runner/`
- **Type:** New

---

## Test Scenarios

### Acceptance Tests (Module Level)

#### Happy Path

1. **Start process, capture output, wait for completion**
   - Given: A Runner configured with command `echo "hello world"`
   - When: I call Start() and wait for completion
   - Then: The provided io.Writer receives "hello world\n", state transitions to Stopped, and no error is returned

2. **Start process with arguments**
   - Given: A Runner configured with command `echo` and args `["-n", "test"]`
   - When: I call Start() and wait for completion
   - Then: The provided io.Writer receives "test" (no newline), state is Stopped

3. **Stop running process gracefully (SIGTERM)**
   - Given: A Runner configured with StopTimeout=1s and a long-running process (e.g., `sleep 60`)
   - When: I call Stop()
   - Then: The process receives SIGTERM, terminates gracefully, state becomes Stopped

4. **Stop with SIGKILL after timeout**
   - Given: A Runner configured with StopTimeout=100ms and a process that ignores SIGTERM
   - When: I call Stop()
   - Then: After 100ms, SIGKILL is sent, process terminates, state becomes Stopped

5. **State transitions**
   - Given: A new Runner
   - Then: Initial state is "initial"
   - When: I call Start()
   - Then: State becomes "running"
   - When: Process completes or Stop() is called
   - Then: State becomes "stopped"

6. **Stderr combined with stdout**
   - Given: A Runner with a command that writes to both stdout and stderr
   - When: I call Start() and wait for completion
   - Then: Both stdout and stderr content appear in the provided io.Writer

7. **Default StopTimeout applied when zero**
   - Given: A Runner configured with StopTimeout=0 (default) and a long-running process
   - When: I call Stop()
   - Then: The default timeout of 5 seconds is used before SIGKILL

#### Edge Cases

1. **Empty command**
   - Given: A Runner configured with an empty command string
   - When: I call Start()
   - Then: An error is returned immediately, state remains "initial"

2. **Nil writer**
   - Given: A Runner configured with a nil io.Writer
   - When: I call Start()
   - Then: An error is returned, process is not started

3. **Command not found**
   - Given: A Runner configured with command `nonexistent_command_xyz`
   - When: I call Start()
   - Then: An error is returned indicating the command was not found

4. **Permission denied**
   - Given: A Runner configured with a non-executable file path
   - When: I call Start()
   - Then: An error is returned indicating permission denied

5. **Stop called before Start**
   - Given: A new Runner (state is "initial")
   - When: I call Stop()
   - Then: No error, state remains "initial" (no-op)

6. **Stop called multiple times**
   - Given: A Runner with a stopped process
   - When: I call Stop() again
   - Then: No error, state remains "stopped" (idempotent)

7. **Start called after Stop**
   - Given: A Runner that has been started and stopped
   - When: I call Start() again
   - Then: An error is returned (runners are single-use)

8. **Context cancellation**
   - Given: A Runner started with a cancellable context
   - When: The context is cancelled
   - Then: The process is terminated gracefully, state becomes "stopped"

9. **Concurrent Stop calls**
   - Given: A Runner with a running process
   - When: Multiple goroutines call Stop() simultaneously
   - Then: No data race, process is stopped exactly once

10. **Concurrent state reads during transitions**
    - Given: A Runner transitioning states
    - When: Goroutines read State() during Start()/Stop()
    - Then: No data race, reads return consistent states

11. **Process exits with non-zero code**
    - Given: A Runner with command `exit 1`
    - When: I call Start() and wait
    - Then: An error wrapping the exit code is returned, state is "stopped"

12. **Large output volume**
    - Given: A Runner with a command producing large output
    - When: Process runs to completion
    - Then: All output is written to the io.Writer without blocking or deadlock

### Unit Tests (Internal Logic)

- Verify state machine transitions are atomic and correct
- Test signal sending logic (SIGTERM followed by SIGKILL)
- Validate mutex protection of internal state
- Test pipe setup and cleanup
- Verify context cancellation propagation

---

## Technical Considerations

### Inputs
| Input | Type | Source | Validation |
|-------|------|--------|------------|
| Config.Command | string | Constructor | Must be non-empty |
| Config.Args | []string | Constructor | Optional, can be nil or empty |
| Config.Output | io.Writer | Constructor | Must be non-nil |
| Config.StopTimeout | time.Duration | Constructor | Graceful shutdown timeout before SIGKILL (default: 5s) |
| ctx | context.Context | Start() | Must be non-nil (use context.Background() if none needed) |

### Outputs
| Output | Type | Description |
|--------|------|-------------|
| state | State | Current runner state (Initial, Running, Stopped) |
| error | error | Execution errors (start failure, exit code, etc.) |

### Processing Rules
1. Process stdout and stderr are combined into a single stream written to the provided io.Writer
2. State transitions: Initial → Running → Stopped (one-way, no restart)
3. Stop() sends SIGTERM first, waits for timeout, then SIGKILL if still running
4. Context cancellation triggers graceful stop (same as Stop())
5. All resources (pipes, goroutines) are cleaned up when process ends
6. Errors during execution are returned to the caller for handling

### Technical Decisions

#### Error Handling Strategy
**Decision:** Return all errors to caller, do not handle internally.

**Rationale:**
- The orchestrator (manager) has better context for deciding how to handle failures
- Different error types may require different responses (retry, alert, ignore)
- Keeps the Runner focused on execution mechanics only

#### Context Support
**Decision:** Accept `context.Context` in Start() for cancellation support.

**Rationale:**
- Standard Go pattern for cancellation and timeouts
- Enables integration with HTTP request contexts, shutdown signals, etc.
- Allows caller to set overall execution timeout

#### Stderr Handling
**Decision:** Combine stdout and stderr into single io.Writer stream.

**Rationale:**
- Simplifies output handling for sensors that may write to either stream
- Ring buffer receives all output in chronological order
- Matches typical shell behavior (`2>&1`)

#### Runner Reusability
**Decision:** Runners are single-use (cannot restart after stop).

**Rationale:**
- Simplifies state machine (no complex reset logic)
- Avoids subtle bugs from reused state
- Manager creates new Runner instances as needed

#### Configuration Struct
**Decision:** Use a `Config` struct instead of individual constructor parameters.

**Rationale:**
- Extensible: new options can be added without breaking the API
- Self-documenting: field names clarify purpose
- Sensible defaults: zero values trigger default behavior (e.g., StopTimeout=0 → 5s)
- Cleaner call sites: `New(Config{Command: "echo", Output: buf})` vs many positional args

### Interface
```go
type State string

const (
    StateInitial State = "initial"
    StateRunning State = "running"
    StateStopped State = "stopped"
)

// Config holds the configuration for a Runner.
type Config struct {
    // Command is the executable to run (required).
    Command string

    // Args are the command-line arguments (optional).
    Args []string

    // Output receives combined stdout and stderr (required).
    Output io.Writer

    // StopTimeout is the grace period before SIGKILL after SIGTERM.
    // If zero, defaults to 5 seconds.
    StopTimeout time.Duration
}

type Runner struct {
    // internal fields
}

// New creates a Runner with the given configuration.
// Returns error if Command is empty or Output is nil.
func New(cfg Config) (*Runner, error)

// Start begins process execution. Blocks until process completes or context is cancelled.
// Returns error if process fails to start or exits with non-zero code.
func (r *Runner) Start(ctx context.Context) error

// Stop terminates the process gracefully. Sends SIGTERM, waits for StopTimeout,
// then SIGKILL if still running. Safe to call multiple times (idempotent).
func (r *Runner) Stop() error

// State returns the current runner state. Safe for concurrent access.
func (r *Runner) State() State

// Wait blocks until the process completes. Returns the same error as Start().
// Can be called from a different goroutine than Start().
func (r *Runner) Wait() error
```

### Error Paths
| Condition | Handling | Recovery |
|-----------|----------|----------|
| Empty command | New() returns error | Caller must provide valid command |
| Nil output writer | New() returns error | Caller must provide non-nil writer |
| Command not found | Start() returns exec.ErrNotFound | Caller decides (log, retry with different command) |
| Permission denied | Start() returns os.ErrPermission | Caller decides (log, alert) |
| Non-zero exit code | Start() returns *exec.ExitError | Caller can inspect exit code |
| Context cancelled | Start() returns context.Canceled | Expected behavior, not an error condition |
| Already started | Start() returns error | Runners are single-use |
| nil context | Start() returns error | Caller must provide context |

---

## Dependencies
- **Depends on:** None (uses only standard library: `os/exec`, `context`, `sync`)
- **Used by:** `manager/` (to run sensor commands)

---

## Notes

### Platform Considerations
- Signal handling (SIGTERM, SIGKILL) is Unix-specific
- Windows support would require different termination approach (not in scope for initial implementation)

### Future Extensions
The `Config` struct makes it easy to add new options without breaking the API:
- `Env []string` - Environment variable configuration
- `Dir string` - Working directory configuration
- Resource limits (CPU, memory) via cgroups
- Output transformation/filtering hooks
