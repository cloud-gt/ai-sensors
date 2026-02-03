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
go fmt ./...                                                    # Format code
go build -o ./tmp/main .                                        # Build
go test ./... -v                                                # Run tests
go run github.com/golangci/golangci-lint/cmd/golangci-lint run  # Lint
```

## Linting Rules

- **Always explicitly ignore unused return values** with `_ =` to avoid `errcheck` warnings
  ```go
  _ = r.Stop()  // Good: explicit ignore
  r.Stop()      // Bad: implicit ignore triggers warning
  ```
- **In tests**, use `//nolint:<linter>` directives only when testing edge cases that intentionally violate rules (e.g., passing `nil` context)
  ```go
  err = r.Start(nil) //nolint:staticcheck // testing nil context behavior
  ```
- **In production code**, handle or log all errors - never silently ignore them
