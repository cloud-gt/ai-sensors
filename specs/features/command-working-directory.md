# Spec: Command Working Directory

## Purpose
Add a mandatory `work_dir` field to command definitions so that each command's process executes in a specified working directory.

## Rationale
Currently, all commands inherit the server's working directory when their process is spawned. In practice, commands need to run from a specific project directory (e.g., `go test ./...` must run from the project root). Without a `work_dir` field, the system is unusable for real-world scenarios where the server and the target project live in different locations.

## Package
- **Location:** `command/`, `runner/`, `manager/`, `server/`
- **Type:** Cross-cutting (field added to model, threaded through all layers)

---

## Test Scenarios

### Acceptance Tests (Module Level)

#### Happy Path

1. **Create a command with work_dir via API**
   - Given: a running server
   - When: `POST /commands` with `{"name": "test", "command": "echo hello", "work_dir": "/some/path"}`
   - Then: response status is 201 and the returned command includes `work_dir: "/some/path"`

2. **Process runs in the specified directory**
   - Given: a command created with `work_dir` set to a temporary directory
   - When: the command is started and runs `pwd`
   - Then: the output contains the temporary directory path

3. **JSON persistence round-trip**
   - Given: a command created with `work_dir: "/some/path"`
   - When: the command is saved to JSON and reloaded
   - Then: the loaded command has `work_dir: "/some/path"`

4. **Get command returns work_dir**
   - Given: a command created with `work_dir`
   - When: `GET /commands/{id}`
   - Then: the response includes the `work_dir` field

#### Edge Cases

1. **Empty work_dir is rejected**
   - Given: a running server
   - When: `POST /commands` with `{"name": "test", "command": "echo hello", "work_dir": ""}`
   - Then: response status is 400 with an error message

2. **Missing work_dir field is rejected**
   - Given: a running server
   - When: `POST /commands` with `{"name": "test", "command": "echo hello"}` (no work_dir)
   - Then: response status is 400 with an error message

3. **Non-existent directory fails at process start**
   - Given: a command with `work_dir` pointing to a path that does not exist
   - When: the command is started
   - Then: start returns an error (OS-level failure)

4. **Concurrent commands with different work_dirs**
   - Given: two commands with different `work_dir` values
   - When: both are started concurrently
   - Then: each process runs in its own directory, outputs do not cross

---

## Technical Considerations

### Inputs

| Input | Type | Source | Validation |
|-------|------|--------|------------|
| work_dir | string | JSON request body / JSON file | Must not be empty (validated in `command.Store.Create`) |

### Outputs

| Output | Type | Description |
|--------|------|-------------|
| Command.WorkDir | string | The working directory persisted with the command |

### Processing Rules

1. `work_dir` is mandatory — `Store.Create` returns `ErrEmptyWorkDir` if empty
2. The `work_dir` value is stored as-is (no path resolution or validation at creation time)
3. Path validity is deferred to the OS at process start time (consistent with existing `Command` field behavior)
4. Runner sets `exec.Cmd.Dir` to the provided directory
5. Manager passes `Command.WorkDir` to `runner.Config.Dir` when creating the runner

### Changes Per Layer

| Layer | File | Change |
|-------|------|--------|
| Model | `command/command.go` | Add `WorkDir string` field with `json:"work_dir"` tag |
| Validation | `command/store.go` | Add `ErrEmptyWorkDir` sentinel, validate in `Create` |
| Runner | `runner/runner.go` | Add `Dir string` to `Config`, set `r.cmd.Dir` in `Start()` |
| Manager | `manager/manager.go` | Pass `cmd.WorkDir` as `Dir` in `runner.Config` |
| API | `server/commands.go` | Add `WorkDir` to create request struct, pass to `Store.Create` |
| Test helper | `server/testclient_test.go` | Add `CreateCommandWithWorkDir` helper |

### Error Paths

| Condition | Handling | Recovery |
|-----------|----------|----------|
| Empty `work_dir` on create | Return `ErrEmptyWorkDir` (400 in API) | Caller must provide a valid path |
| Non-existent directory at start | OS error propagated from `cmd.Start()` | Caller receives error, command stays in not_started state |

---

## Dependencies
- **Depends on:** `command/` (model), `runner/` (execution), `manager/` (orchestration), `server/` (API)
- **Used by:** All commands going forward — `work_dir` is mandatory for every command
