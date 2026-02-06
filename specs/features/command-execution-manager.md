# Spec: Command Execution Manager

## Purpose
Central orchestration layer that links the command Store, process Runner, and RingBuffer to manage the complete lifecycle of command execution via command IDs.

## Rationale
The ai-sensors server needs a unified interface to start, stop, and monitor commands. The Manager acts as the glue layer that:
- Looks up command definitions from the Store
- Creates RingBuffers to capture output
- Spawns Runners to execute processes
- Tracks state and provides status/output access

This separation keeps each component testable in isolation while providing a cohesive API for the HTTP layer (Feature 5).

## Package
- **Location:** `manager/`
- **Type:** New

---

## Test Scenarios

### Acceptance Tests (Integration Level)

#### Happy Path

1. **Start command and read output**
   - Given: A command with ID `X` exists in the store
   - When: `Start(X)` is called
   - Then: The process runs and produces output accessible via `Output(X)`

2. **Full lifecycle: start → output → status → stop**
   - Given: A command with ID `X` exists in the store
   - When: `Start(X)` is called, then `Output(X)`, then `Status(X)`, then `Stop(X)`
   - Then: Each operation succeeds, status transitions from running to stopped

3. **Run multiple commands concurrently**
   - Given: Commands with IDs `X` and `Y` exist in the store
   - When: `Start(X)` and `Start(Y)` are called
   - Then: Both processes run independently with separate buffers

4. **Auto-cleanup on natural termination**
   - Given: A short-lived command (e.g., `echo hello`) with ID `X`
   - When: `Start(X)` is called and the process finishes naturally
   - Then: Status becomes stopped, output is still accessible, resources are cleaned up

5. **Accurate status tracking**
   - Given: A command with ID `X` exists in the store
   - When: Status is queried at different points (before start, while running, after stop)
   - Then: Status returns appropriate state (not_started, running, stopped)

#### Edge Cases

1. **Start unknown command ID**
   - Given: No command with ID `Z` exists in the store
   - When: `Start(Z)` is called
   - Then: Returns `ErrCommandNotFound` (or similar)

2. **Stop unknown command ID**
   - Given: No command with ID `Z` is running
   - When: `Stop(Z)` is called
   - Then: Returns `ErrNotRunning` (or similar)

3. **Output for unknown command ID**
   - Given: No command with ID `Z` is running or has run
   - When: `Output(Z)` is called
   - Then: Returns `ErrNotRunning` (or similar)

4. **Double start (command already running) — idempotent**
   - Given: Command with ID `X` is already running
   - When: `Start(X)` is called again
   - Then: Returns `(false, nil)` — no error, bool indicates no new process was started

5. **Double stop (command already stopped) — idempotent**
   - Given: Command with ID `X` was running but has been stopped
   - When: `Stop(X)` is called again
   - Then: Returns `nil` — purely idempotent, stopping a stopped command is a no-op

6. **Concurrent access from multiple goroutines**
   - Given: Command with ID `X` exists
   - When: Multiple goroutines call Start, Stop, Status, Output concurrently
   - Then: No data races, operations are thread-safe

7. **Resource cleanup on shutdown**
   - Given: Multiple commands are running
   - When: `Shutdown()` or `Close()` is called on the Manager
   - Then: All running processes are stopped, all resources freed

### Unit Tests

- Map of running processes is correctly maintained (add on start, update on stop)
- Buffer capacity is configurable and applied correctly
- Context cancellation propagates to runners
- Status transitions are atomic and consistent

---

## Technical Considerations

### Inputs

| Input | Type | Source | Validation |
|-------|------|--------|------------|
| commandID | `uuid.UUID` | Caller (HTTP handler) | Must exist in Store for Start; must be running for Stop/Output |
| bufferCapacity | `int` | Configuration | Must be > 0, use sensible default (e.g., 1000) |

### Outputs

| Output | Type | Description |
|--------|------|-------------|
| started | `bool` | From `Start()`: true if a new process was started, false if already running |
| output lines | `[]string` | Buffer contents for a running/stopped command |
| status | `Status` (enum) | Current state: not_started, running, stopped |
| error | `error` | Operation-specific errors (nil for idempotent no-ops) |

### Data Structures

```go
type Manager struct {
    store         *command.Store  // or interface
    bufferCap     int
    mu            sync.RWMutex
    instances     map[uuid.UUID]*Instance
}

type Instance struct {
    command  command.Command
    runner   *runner.Runner
    buffer   *buffer.RingBuffer
    status   Status
    cancel   context.CancelFunc
}

type Status string

const (
    StatusNotStarted Status = "not_started"
    StatusRunning    Status = "running"
    StatusStopped    Status = "stopped"
)
```

### Processing Rules

1. `Start(id)`: Look up command in store → create buffer → create runner → start runner in goroutine → update instance map
2. `Stop(id)`: Look up instance → call runner.Stop() → update status
3. `Output(id)`: Look up instance → return buffer.Lines()
4. `Status(id)`: Look up instance → return status
5. `Shutdown()`: Iterate all running instances → stop each → clear map

### Alternative Paths

- If command not found in store: return `ErrCommandNotFound`
- If instance not found in map (for Output/Status): return `ErrNotRunning`
- If command already running (Start): return `(false, nil)` — idempotent no-op
- If command not running (Stop): return `nil` — idempotent no-op
- If runner fails to start (e.g., command not found on system): return error, don't add to map

### Error Paths

| Condition | Handling | Recovery |
|-----------|----------|----------|
| Command ID not in store | Return `ErrCommandNotFound` | Caller should verify command exists |
| Command not running (for Output/Status) | Return `ErrNotRunning` | Caller should start command first |
| Command already running | Return `(false, nil)` — idempotent | No action needed, command is running |
| Command already stopped (Stop called) | Return `nil` — idempotent | No action needed |
| Runner fails to start | Return underlying error | Caller can retry or fix command |
| Store unavailable | Return store error | Fail fast, let caller handle |

---

## Dependencies

- **Depends on:**
  - `command.Store` - for looking up command definitions
  - `buffer.RingBuffer` - for capturing output
  - `runner.Runner` - for process execution
  - `github.com/google/uuid` - for command ID type

- **Used by:**
  - `server/` (Feature 5) - HTTP handlers will call Manager methods

---

## API Surface

```go
// New creates a new Manager with the given store and options
func New(store *command.Store, opts ...Option) *Manager

// Option pattern for configuration
type Option func(*Manager)
func WithBufferCapacity(cap int) Option

// Core operations (idempotent)
func (m *Manager) Start(ctx context.Context, id uuid.UUID) (started bool, err error)
func (m *Manager) Stop(id uuid.UUID) error  // idempotent: stopping a stopped command returns nil
func (m *Manager) Output(id uuid.UUID) ([]string, error)
func (m *Manager) OutputLastN(id uuid.UUID, n int) ([]string, error)
func (m *Manager) Status(id uuid.UUID) (Status, error)

// Lifecycle
func (m *Manager) Shutdown(ctx context.Context) error
```

---

## Notes

- The Manager should use interfaces for Store dependency to facilitate testing
- Consider whether Output should be available after a command is stopped (probably yes, for debugging)
- Buffer capacity could be per-command in the future, but start with global default
- The `Start` method should accept a context for cancellation propagation
- **Idempotency pattern**: `Start` returns `(bool, error)` where the bool indicates if a new process was actually started. `Stop` is purely idempotent (returns `nil` even if already stopped). This allows callers to use these methods as "ensure running" / "ensure stopped" without error handling for already-in-state cases.
