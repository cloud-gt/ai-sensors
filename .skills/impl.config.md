# impl Configuration — Go Project

## Language

Go

## Spec Directory

`specs/`

## Source File Convention

- Production code: `<package>/<feature>.go`
- Test code: `<package>/<feature>_test.go`

## Test Framework

- Framework: `testing` (stdlib) + `testify`
- Assertions: `github.com/stretchr/testify/assert` for test checks, `github.com/stretchr/testify/require` for setup that should fail fast
- Subtests: Use `t.Run()` for grouping related scenarios
- Concurrency: Use `t.Parallel()` and sync primitives when needed

### Test Pattern Template

```go
package <package>

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func Test<Feature>_<Scenario>(t *testing.T) {
    // setup
    thing, err := New(10)
    require.NoError(t, err)

    // act
    thing.DoSomething()

    // assert
    assert.Equal(t, expected, thing.Result())
}
```

## Verification Commands

### 1. Format

```bash
go fmt ./...
```

Success: no files modified (empty output).

### 2. Build

```bash
go build -o ./tmp/main .
```

Success: exit code 0, no errors.

### 3. Test

```bash
go test ./<package>/... -v
```

Success: all tests pass, exit code 0.

### 4. Lint

```bash
go run github.com/golangci/golangci-lint/cmd/golangci-lint run ./<package>/...
```

Success: no warnings or errors, exit code 0.

## Code Style Rules

- No comments unless the logic is truly non-obvious
- Handle or log all errors — never silently ignore return values; use `slog.Warn()` for errors that cannot be returned
- Explicitly ignore unused return values with `_ =`
- Use `//nolint:<linter>` directives only for intentional edge case tests
