---
name: impl
description: Implement a feature from its specification. Reads a spec file from specs/, implements tests and production code, then verifies all exit conditions (format, build, tests).
argument-hint: <feature-name>
allowed-tools: Read, Write, Edit, Bash, Glob, Grep, TaskCreate, TaskUpdate, TaskList, TaskGet
---

<impl-skill>

# Impl Skill: Feature Implementation from Specification

You are implementing a feature for the ai-sensors project based on its specification. Follow this workflow strictly.

## Phase 1: Load Specification

1. **Determine the feature name:**
   - If provided as argument (`$ARGUMENTS`), use it
   - Otherwise, ask: "What feature would you like to implement?"

2. **Load the spec file:**
   - Read `specs/<feature-name>.md`
   - If the file doesn't exist, inform the user and stop

3. **Extract key information from the spec:**
   - Package location (e.g., `buffer/`)
   - Interface definition (struct and functions)
   - Test scenarios (Happy Path + Edge Cases)
   - Processing rules
   - Error handling requirements

4. **Explore existing patterns:**
   - Look at existing test files for testify/assert patterns
   - Check `.air.toml` for build command reference
   - Identify any existing code in the target package

## Phase 2: Create Task List

Use TaskCreate to track progress. Create one task per test scenario plus implementation tasks:

**Test tasks** (one per scenario from the spec):
- Subject: "Test: [scenario name]"
- Description: The Given/When/Then from the spec

**Implementation tasks:**
- Subject: "Implement [feature] production code"
- Description: Implement the interface defined in the spec

**Verification task:**
- Subject: "Verify exit conditions"
- Description: Run format, build, and tests checks

Set up dependencies:
- Production code task is blocked by all test tasks
- Verification task is blocked by production code task

## Phase 3: Implement Tests

For each test scenario from the spec, translate Given/When/Then to Go test code.

**Test file location:** `<package>/<feature>_test.go`

**Test structure pattern:**
```go
package <package>

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func Test<FeatureName>_<ScenarioName>(t *testing.T) {
    // Given: [setup from spec]

    // When: [action from spec]

    // Then: [assertion from spec]
    assert.Equal(t, expected, actual)
}
```

**Guidelines:**
- Use `require` for setup assertions that should fail fast
- Use `assert` for test assertions
- Group related scenarios in subtests with `t.Run()` when appropriate
- For concurrent tests, use `t.Parallel()` and sync primitives
- Mark each test task as `in_progress` before starting, `completed` when done

## Phase 4: Implement Production Code

**Production file location:** `<package>/<feature>.go`

1. Start by creating the file with the interface defined in the spec
2. Implement each function to make tests pass
3. Run tests incrementally after each significant change:
   ```bash
   go test ./<package>/... -v
   ```

**Guidelines:**
- Follow the processing rules exactly as specified
- Implement error handling as defined in the spec
- Keep the implementation minimal - only what's needed to pass tests
- Mark the implementation task as `in_progress` before starting

## Phase 5: Verification (Exit Conditions)

All three checks must pass before marking the feature complete:

### 1. Format Check
```bash
go fmt ./...
```
**Success criteria:** No files are modified (output is empty or shows no changes)

### 2. Build Check
```bash
go build -o ./tmp/main .
```
**Success criteria:** Exit code 0, no errors

### 3. Test Check
```bash
go test ./<package>/... -v
```
**Success criteria:** All tests pass, exit code 0

**If any check fails:**
1. Fix the issue
2. Re-run all checks from the beginning
3. Repeat until all pass

**When all checks pass:**
- Mark the verification task as `completed`
- Report success to the user with a summary of what was implemented

## Important Notes

- **Follow the spec exactly** - do not add features not in the spec
- **Test-first approach** - write tests before production code when possible
- **Incremental progress** - run tests frequently, fix issues as they arise
- **Use TaskUpdate** - keep task statuses current for visibility
- **Error messages matter** - when tests fail, read the error carefully before fixing

## Reference Files
- `.air.toml` - Contains the project's build command
- `server/server_test.go` - Example of test patterns used in this project
- Existing files in target package - Match their style and patterns

</impl-skill>
