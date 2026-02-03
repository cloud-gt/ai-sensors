# AI Sensors Project Guidelines

## Language Policy

**All code, comments, documentation, specs, commit messages, and generated content must be written in English.**

This applies regardless of the language the user communicates in. Even if the user speaks French (or any other language), all artifacts produced must be in English.

## Project Structure

- `specs/` - Feature specifications (written in English)
- `server/` - HTTP API and handlers
- `buffer/` - Ring buffer and output storage
- `runner/` - Process execution and lifecycle
- `command/` - Command definitions and persistence
- `manager/` - Orchestration of components

## Testing

- Use `testify/assert` for assertions
- Use `testify/require` for setup assertions that should fail fast
- Test files: `<package>/<feature>_test.go`

## Build & Verification

```bash
go fmt ./...              # Format code
go build -o ./tmp/main .  # Build
go test ./... -v          # Run tests
```
