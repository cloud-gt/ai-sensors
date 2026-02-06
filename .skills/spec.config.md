# spec Configuration — Go Project

## Language

Go

## Spec Directory

`specs/`

## Project Packages

- `buffer/` — Ring buffer and output storage
- `runner/` — Process execution and lifecycle
- `command/` — Command definitions and persistence
- `manager/` — Orchestration of components
- `server/` — HTTP API and handlers

## Purpose Categories

- "Data storage/retrieval" — Storing, caching, or retrieving data
- "Process management" — Running, monitoring, or controlling processes
- "API endpoint" — Exposing functionality via HTTP
- "Utility/helper" — Supporting other features

## Default Edge Cases

- "Empty/nil inputs" — Handle missing or null data gracefully
- "Boundary conditions" — Min/max values, capacity limits
- "Concurrent access" — Thread safety with multiple goroutines (sync.Mutex, channels)
- "Resource cleanup" — Proper cleanup on stop/error

## Error Handling Patterns

- "Return error" — Return errors to caller for handling
- "Log and continue" — Log the error with `slog.Warn()` but continue operation
- "Panic on critical" — Panic on unrecoverable errors
- "Retry with backoff" — Automatic retry for transient failures
