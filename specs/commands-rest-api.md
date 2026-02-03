# Spec: Commands REST API

## Purpose
Expose all command management and execution functionality via a RESTful HTTP API, enabling external clients (including AI agents) to create, manage, execute, and monitor commands.

## Rationale
The ai-sensors server needs an HTTP interface to allow external tools and agents to interact with the system. The REST API acts as the public-facing layer that:
- Provides CRUD operations for command definitions
- Controls command execution lifecycle (start/stop)
- Exposes command status and output
- Returns appropriate HTTP status codes and error details

This keeps the HTTP layer thin (delegating to Manager) while providing a clean, standardized interface for clients.

## Package
- **Location:** `server/`
- **Type:** Extension (enriching existing minimal server)

---

## Test Scenarios

### Acceptance Tests (HTTP Level)

#### Happy Path

1. **List commands (empty)**
   - Given: No commands exist in the store
   - When: `GET /commands` is called
   - Then: Returns 200 with `{"commands": []}`

2. **Create a command**
   - Given: No command with name "test-cmd" exists
   - When: `POST /commands` with body `{"name": "test-cmd", "command": "echo hello"}`
   - Then: Returns 201 with command details including generated UUID

3. **Get command by ID**
   - Given: A command exists with ID `X`
   - When: `GET /commands/X` is called
   - Then: Returns 200 with command details (id, name, command)

4. **List commands (non-empty)**
   - Given: Commands "cmd-a" and "cmd-b" exist
   - When: `GET /commands` is called
   - Then: Returns 200 with `{"commands": [cmd-a, cmd-b]}`

5. **Delete a command**
   - Given: A command exists with ID `X`
   - When: `DELETE /commands/X` is called
   - Then: Returns 204 No Content, command is removed

6. **Start a command**
   - Given: A command with ID `X` exists and is not running
   - When: `POST /commands/X/start` is called
   - Then: Returns 200 with `{"started": true}` (or 200 with `{"started": false}` if already running)

7. **Stop a command**
   - Given: A command with ID `X` is running
   - When: `POST /commands/X/stop` is called
   - Then: Returns 200, command is stopped

8. **Get command status**
   - Given: A command with ID `X` has been started
   - When: `GET /commands/X/status` is called
   - Then: Returns 200 with `{"status": "running"}` or `{"status": "stopped"}`

9. **Get full output**
   - Given: A command with ID `X` is running and has produced output
   - When: `GET /commands/X/output` is called
   - Then: Returns 200 with `{"lines": ["line1", "line2", ...]}`

10. **Get last N lines of output**
    - Given: A command with ID `X` has produced 100 lines of output
    - When: `GET /commands/X/output?lines=10` is called
    - Then: Returns 200 with only the last 10 lines

11. **Full E2E lifecycle**
    - Given: Empty system
    - When: Create command → Start → Wait for output → Get status → Get output → Stop → Delete
    - Then: Each step succeeds with appropriate response codes

#### Edge Cases

1. **Create command with invalid JSON**
   - Given: Any state
   - When: `POST /commands` with malformed JSON body
   - Then: Returns 400 Bad Request with error details

2. **Create command with missing required fields**
   - Given: Any state
   - When: `POST /commands` with `{"name": "test"}` (missing command field)
   - Then: Returns 400 Bad Request with validation error

3. **Create command with duplicate name**
   - Given: A command "test-cmd" already exists
   - When: `POST /commands` with `{"name": "test-cmd", "command": "echo world"}`
   - Then: Returns 409 Conflict

4. **Get unknown command**
   - Given: No command with ID `Z` exists
   - When: `GET /commands/Z` is called
   - Then: Returns 404 Not Found

5. **Delete unknown command**
   - Given: No command with ID `Z` exists
   - When: `DELETE /commands/Z` is called
   - Then: Returns 404 Not Found

6. **Start unknown command**
   - Given: No command with ID `Z` exists
   - When: `POST /commands/Z/start` is called
   - Then: Returns 404 Not Found

7. **Invalid UUID format**
   - Given: Any state
   - When: `GET /commands/not-a-uuid` is called
   - Then: Returns 400 Bad Request

8. **Stop command that was never started**
   - Given: A command with ID `X` exists but has never been started
   - When: `POST /commands/X/stop` is called
   - Then: Returns 404 Not Found (or 200 if idempotent, per manager behavior)

9. **Get status of never-started command**
   - Given: A command with ID `X` exists but has never been started
   - When: `GET /commands/X/status` is called
   - Then: Returns 200 with `{"status": "not_started"}` or 404 if not tracked

10. **Get output of never-started command**
    - Given: A command with ID `X` exists but has never been started
    - When: `GET /commands/X/output` is called
    - Then: Returns 404 Not Found (no instance exists)

11. **Invalid lines parameter**
    - Given: A command with ID `X` is running
    - When: `GET /commands/X/output?lines=abc` is called
    - Then: Returns 400 Bad Request

12. **Negative lines parameter**
    - Given: A command with ID `X` is running
    - When: `GET /commands/X/output?lines=-5` is called
    - Then: Returns 400 Bad Request

13. **Concurrent requests to same command**
    - Given: A command with ID `X` exists
    - When: Multiple goroutines call start/stop/status/output concurrently
    - Then: No data races, all requests complete without error

14. **Delete running command**
    - Given: A command with ID `X` is currently running
    - When: `DELETE /commands/X` is called
    - Then: Returns 409 Conflict (cannot delete running command) OR stops and deletes

15. **Server shutdown with running commands**
    - Given: Multiple commands are running
    - When: Server receives shutdown signal
    - Then: All commands are stopped gracefully, resources freed

### Unit Tests

- Request body parsing and validation
- Query parameter parsing (lines=N)
- Response JSON serialization
- Error response formatting
- UUID parsing and validation from URL path

---

## Technical Considerations

### Inputs

| Input | Type | Source | Validation |
|-------|------|--------|------------|
| id | `uuid.UUID` | URL path parameter | Must be valid UUID format |
| name | `string` | JSON body (POST /commands) | Non-empty, valid command name |
| command | `string` | JSON body (POST /commands) | Non-empty shell command |
| lines | `int` | Query param (GET /output) | Optional, must be positive integer if present |

### Outputs

| Output | Type | Description |
|--------|------|-------------|
| Command | `JSON object` | `{id, name, command}` |
| Command list | `JSON object` | `{commands: [{id, name, command}, ...]}` |
| Status | `JSON object` | `{status: "not_started"|"running"|"stopped"}` |
| Output | `JSON object` | `{lines: ["line1", "line2", ...]}` |
| Start result | `JSON object` | `{started: bool}` |
| Error | `JSON object` | `{error: "message"}` |

### API Endpoints

| Method | Path | Description | Success | Errors |
|--------|------|-------------|---------|--------|
| GET | /commands | List all commands | 200 | - |
| POST | /commands | Create a command | 201 | 400, 409 |
| GET | /commands/{id} | Get command details | 200 | 404 |
| DELETE | /commands/{id} | Delete a command | 204 | 404, 409 |
| POST | /commands/{id}/start | Start command | 200 | 404 |
| POST | /commands/{id}/stop | Stop command | 200 | 404 |
| GET | /commands/{id}/status | Get command status | 200 | 404 |
| GET | /commands/{id}/output | Get command output | 200 | 404, 400 |

### Request/Response Formats

**POST /commands** (Create)
```json
// Request
{
  "name": "watch-tests",
  "command": "go test ./... -v"
}

// Response 201
{
  "id": "uuid-here",
  "name": "watch-tests",
  "command": "go test ./... -v"
}
```

**GET /commands/{id}** (Get)
```json
// Response 200
{
  "id": "uuid-here",
  "name": "watch-tests",
  "command": "go test ./... -v"
}
```

**GET /commands** (List)
```json
// Response 200
{
  "commands": [
    {"id": "uuid-1", "name": "watch-tests", "command": "go test ./..."},
    {"id": "uuid-2", "name": "lint", "command": "golangci-lint run"}
  ]
}
```

**POST /commands/{id}/start** (Start)
```json
// Response 200
{
  "started": true
}
```

**GET /commands/{id}/status** (Status)
```json
// Response 200
{
  "status": "running"
}
```

**GET /commands/{id}/output** (Output)
```json
// Response 200
{
  "lines": [
    "=== RUN   TestSomething",
    "--- PASS: TestSomething (0.00s)",
    "PASS"
  ]
}
```

**Error Response**
```json
// Response 4xx/5xx
{
  "error": "command not found"
}
```

### Processing Rules

1. All endpoints use JSON for request/response bodies
2. Command lookup is by UUID in URL path parameters
3. UUID must be parsed and validated before any operation
4. Start is idempotent: returns `{started: false}` if already running (not an error)
5. Stop returns 200 even if already stopped (idempotent per manager behavior)
6. Output and status require the command to have been started at least once

### Alternative Paths

- If UUID format invalid: return 400 Bad Request
- If command ID not found: return 404 with error body
- If command already exists (POST with same name): return 409 Conflict
- If command already running (start): return 200 with `{started: false}`
- If command already stopped (stop): return 200 (idempotent)
- If lines param invalid: return 400 Bad Request

### Error Paths

| Condition | HTTP Status | Response Body | Recovery |
|-----------|-------------|---------------|----------|
| Malformed JSON | 400 | `{error: "invalid JSON"}` | Fix request body |
| Missing required field | 400 | `{error: "name is required"}` | Add missing field |
| Invalid UUID format | 400 | `{error: "invalid command ID"}` | Use valid UUID |
| Invalid query param | 400 | `{error: "lines must be positive integer"}` | Fix query param |
| Command not found | 404 | `{error: "command not found"}` | Create command first |
| Command not running | 404 | `{error: "command not running"}` | Start command first |
| Duplicate command name | 409 | `{error: "command already exists"}` | Use different name |
| Cannot delete running | 409 | `{error: "cannot delete running command"}` | Stop command first |
| Internal error | 500 | `{error: "internal server error"}` | Check server logs |

---

## Dependencies

- **Depends on:**
  - `manager.Manager` - for command execution lifecycle
  - `command.Store` - for command CRUD operations
  - `github.com/go-chi/chi/v5` - HTTP router (already in use)
  - `encoding/json` - JSON serialization

- **Used by:**
  - External clients (curl, AI agents, etc.)
  - Future: Web UI

---

## Server Structure

```go
type Server struct {
    router  chi.Router
    manager *manager.Manager
    store   *command.Store
}

func New(store *command.Store, mgr *manager.Manager) *Server

func (s *Server) ListenAndServe(addr string) error
func (s *Server) Shutdown(ctx context.Context) error

// Handlers (internal)
func (s *Server) handleListCommands(w http.ResponseWriter, r *http.Request)
func (s *Server) handleCreateCommand(w http.ResponseWriter, r *http.Request)
func (s *Server) handleGetCommand(w http.ResponseWriter, r *http.Request)
func (s *Server) handleDeleteCommand(w http.ResponseWriter, r *http.Request)
func (s *Server) handleStartCommand(w http.ResponseWriter, r *http.Request)
func (s *Server) handleStopCommand(w http.ResponseWriter, r *http.Request)
func (s *Server) handleGetStatus(w http.ResponseWriter, r *http.Request)
func (s *Server) handleGetOutput(w http.ResponseWriter, r *http.Request)
```

---

## Notes

- The server should support graceful shutdown
- Consider adding request logging middleware (chi already provides this)
- Command IDs in URLs are UUIDs (validated on parse)
- The `lines` query parameter is optional; if omitted, return full buffer
- All responses should have `Content-Type: application/json` header
- Consider adding CORS headers for future web UI support
