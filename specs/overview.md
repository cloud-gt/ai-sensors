# AI-Sensors - Project Overview

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

## Feature Roadmap

### ✅ Feature 1: Ring Buffer
**Goal:** Store the N most recent output lines without exploding memory

**Package:** `buffer/`

**Spec:** [ring-buffer.md](./features/ring-buffer.md)

---

### ✅ Feature 2: Process Runner
**Goal:** Spawn an external process and capture its output

**Package:** `runner/`

**Spec:** [process-runner.md](./features/process-runner.md)

---

### ✅ Feature 3: Command Definition + JSON Store
**Goal:** Define and persist commands

**Package:** `command/`

**Spec:** [command-definition.md](./features/command-definition.md)

---

### ✅ Feature 4: Manager (orchestration)
**Goal:** Link Store + Runner + Buffer

**Package:** `manager/`

**Spec:** [command-execution-manager.md](./features/command-execution-manager.md)

---

### ✅ Feature 5: REST API
**Goal:** Expose everything via HTTP

**Package:** `server/`

**Spec:** [commands-rest-api.md](./features/commands-rest-api.md)

---

### Feature 6: Command Working Directory
**Goal:** Add a mandatory `work_dir` field to commands so each process runs in the correct directory

**Package:** `command/`, `runner/`, `manager/`, `server/` (cross-cutting)

**Spec:** [command-working-directory.md](./features/command-working-directory.md)

---

### Feature 7: Source Abstraction
**Goal:** Abstract data producers from ring buffers, allowing different input types beyond process execution

**Package:** `source/`

**Spec:** [source-abstraction.md](./features/source-abstraction.md)

---

### Feature 8: Line Pipeline (Observable Ring Buffer)
**Goal:** Enable real-time line post-processing via an observable pattern

**Package:** `buffer/` (extension of F1)

**Scope:**
- Add an observation mechanism to RingBuffer (callbacks + channels)
- Multi-observer support (multiple consumers can listen)
- Observers receive lines in real-time
- Circular storage continues to work

---

## Implementation Order

```
F1: buffer/     ─┐
                 ├──> F4: manager/ ──> F5: server/
F2: runner/     ─┤
                 │
F3: command/    ─┘

F6: command working directory ──> cross-cutting, next up
F7: source/                   ──> can be done after F4
F8: buffer/ (observable)      ──> can be added after F1 or after F5
```

---

## Backlog (future features)

- **SSE Streaming:** Real-time output without polling
- **Multi-project:** Support for multiple projects
- **Templates:** Presets for Go, Node, Rust, etc.
- **Filtering:** Grep-like on the buffer
- **MCP Integration:** Expose as MCP tool for agents
