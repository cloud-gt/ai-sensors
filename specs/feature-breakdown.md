# Plan: AI-Sensors - Feature Breakdown

## Project Vision

A server that shortens the feedback loop for code agents by:
- Running commands in watch/continuous mode
- Capturing their output in real-time
- Exposing these logs via a simple API

## Architectural Decisions

- **Scope:** Single-project (one server = one project)
- **Structure:** Separate Go modules for each logical component (reusability)
- **Persistence:** Simple JSON file (load at startup, save on each modification)
- **Output format:** Raw lines (no metadata)

## Testing Philosophy

- **Decoupled modules:** Each package (buffer, runner, command) must be testable in isolation
- **No unnecessary mocks:** Unit tests test the module directly, without mocking its internals
- **Dependencies via interfaces:** When a module depends on another, use interfaces to allow injection

---

## Feature Breakdown

### Feature 1: Ring Buffer
**Goal:** Store the N most recent output lines without exploding memory

**Package:** `buffer/`

**Full spec:** [ring-buffer.md](./ring-buffer.md)

---

### Feature 2: Process Runner
**Goal:** Spawn an external process and capture its output

**Package:** `runner/`

**Full spec:** [process-runner.md](./process-runner.md)

---

### Feature 3: Command Definition + JSON Store
**Goal:** Define and persist commands

**Package:** `command/`

**Full spec:** [command-definition.md](./command-definition.md)
- Full CRUD (Add, Get, Remove, List)
- Verify that Save/Load preserves data
- Test with non-existent file (automatic creation)

---

### Feature 4: Manager (orchestration)
**Goal:** Link Store + Runner + Buffer

**Package:** `manager/`

**Full spec:** [command-execution-manager.md](./command-execution-manager.md)

---

### Feature 5: REST API
**Goal:** Expose everything via HTTP

**Full spec:** [commands-rest-api.md](./commands-rest-api.md)

**Command Endpoints:**
- `GET /commands` - List defined commands
- `POST /commands` - Create a command
- `GET /commands/{id}` - Command details
- `DELETE /commands/{id}` - Delete

**Execution Endpoints:**
- `POST /commands/{id}/start` - Start
- `POST /commands/{id}/stop` - Stop
- `GET /commands/{id}/status` - State (running/stopped/not_started)

**Output Endpoints:**
- `GET /commands/{id}/output` - Full buffer
- `GET /commands/{id}/output?lines=N` - Last N lines

**Package:** `server/`
- `server.go` - Main server (mounts sub-routers)
- `commands.go` - Commands API handlers + router
- `commands_test.go` - HTTP tests
- `testclient_test.go` - Test client wrapper for cleaner tests

**Tests:** HTTP tests using a TestClient wrapper
- Use `httptest` to test handlers
- E2E scenarios: create command via API, start, read output, stop
- Verify HTTP return codes, errors, etc.

---

### Feature 6: Line Pipeline (Observable Ring Buffer)
**Goal:** Enable real-time line post-processing via an observable pattern

**Context:**
- The RingBuffer (F1) stores lines but is "passive" (read on demand)
- We want to react to each new line (push notification)
- Use cases: filtering, structured parsing, forwarding to other systems

**Scope:**
- Add an observation mechanism to RingBuffer:
  - Callback option: `OnLine(fn func(line string))` - called for each new line
  - Channel option: `Subscribe() <-chan string` - returns a channel that receives new lines
- Multi-observer support (multiple consumers can listen)
- Observers receive lines in real-time (in `Write()`)
- Circular storage continues to work (queryable history)

**Idiomatic Go pattern:**
```go
// Fan-out: RingBuffer becomes a "hub"
rb := buffer.New(1000)

// Observer 1: log all lines
rb.OnLine(func(line string) {
    log.Println(line)
})

// Observer 2: filter errors
rb.OnLine(func(line string) {
    if strings.Contains(line, "ERROR") {
        errorQueue <- line
    }
})

// Process writes normally
process.Start(rb)  // rb still implements io.Writer
```

**Package:** `buffer/` (extension of F1)

**Tests:**
- Verify callbacks are called for each new line
- Test with multiple simultaneous observers
- Verify circular storage still works
- Test concurrency (slow observers don't block `Write()`)

**Considerations:**
- Callbacks must be non-blocking (or executed in separate goroutines)
- Option: buffered channel to absorb spikes
- RingBuffer keeps its storage role, observation is opt-in

---

## Implementation Order

```
F1: buffer/     ─┐
                 ├──> F4: manager/ ──> F5: server/
F2: runner/     ─┤
                 │
F3: command/    ─┘

F6: buffer/ (observable) ──> can be added after F1 or after F5
```

**F1, F2, F3** are independent (can be done in parallel or in any order)
**F4** integrates all three
**F5** exposes the manager via HTTP
**F6** extends F1 with observability - can be done once F1 is stable, or after MVP (F5)

---

## Final Package Structure

```
ai-sensors/
├── main.go
├── commands.json            # Persistence (created at runtime)
├── buffer/
│   ├── ringbuffer.go
│   └── ringbuffer_test.go
├── runner/
│   ├── runner.go
│   └── runner_test.go
├── command/
│   ├── command.go           # Command struct + errors
│   ├── store.go             # In-memory store with CRUD
│   ├── store_test.go
│   ├── repository.go        # Repository interface
│   └── repository_test.go
├── manager/
│   ├── manager.go           # Orchestrates Store + Runner + Buffer
│   └── manager_test.go
└── server/
    ├── server.go            # Main server (mounts sub-routers)
    ├── commands.go          # Commands API handlers + router
    ├── commands_test.go     # HTTP integration tests
    └── testclient_test.go   # Test client wrapper
```

---

## Verification (End-to-End)

To test the complete system:

1. Start the server: `go run main.go`
2. Create a command:
   ```bash
   curl -X POST localhost:3000/commands \
     -H "Content-Type: application/json" \
     -d '{"name":"watch-tests","command":"go test ./... -v"}'
   # Response: {"id":"<uuid>","name":"watch-tests","command":"go test ./... -v"}
   ```
3. Start the command: `curl -X POST localhost:3000/commands/<uuid>/start`
4. Check status: `curl localhost:3000/commands/<uuid>/status`
5. Read output: `curl localhost:3000/commands/<uuid>/output`
6. Read last 10 lines: `curl localhost:3000/commands/<uuid>/output?lines=10`
7. Stop: `curl -X POST localhost:3000/commands/<uuid>/stop`
8. Delete: `curl -X DELETE localhost:3000/commands/<uuid>`

---

## Backlog (future features)

- **SSE Streaming:** Real-time output without polling
- **Multi-project:** Support for multiple projects
- **Templates:** Presets for Go, Node, Rust, etc.
- **Filtering:** Grep-like on the buffer
- **MCP Integration:** Expose as MCP tool for agents
