---
name: spec
description: Create a feature specification with test scenarios and technical considerations. Use when planning a new feature, defining acceptance tests, or documenting technical requirements.
argument-hint: [feature-name]
allowed-tools: Read, Grep, Glob, AskUserQuestion, Write
---

<spec-skill>

# Spec Skill: Feature Specification Generator

You are creating a structured feature specification for the ai-sensors project. Follow this workflow strictly.

## Phase 1: Context & Understanding

1. **Determine the feature name:**
   - If provided as argument (`$ARGUMENTS`), use it
   - Otherwise, ask: "What feature would you like to specify?"

3. **Clarify the feature with these questions:**

Ask the user about the **Purpose**:
- Question: "What is the main goal of this feature?"
- Options:
  - "Data storage/retrieval" - Storing, caching, or retrieving data
  - "Process management" - Running, monitoring, or controlling processes
  - "API endpoint" - Exposing functionality via HTTP
  - "Utility/helper" - Supporting other features

Ask the user about the **Package location**:
- Question: "Which package should contain this feature?"
- Options:
  - "buffer/" - Ring buffer and output storage
  - "runner/" - Process execution and lifecycle
  - "command/" - Command definitions and persistence
  - "manager/" - Orchestration of components
  - "server/" - HTTP API and handlers
  - (Other - let user specify for new packages)

## Phase 2: Test Discovery

Ask the user about **Happy path behaviors**:
- Question: "What are the main behaviors to test (happy path)?"
- Let the user describe 2-4 key scenarios that should work

Ask about **Edge cases** (multiSelect: true):
- Question: "Which edge cases should be handled?"
- Options:
  - "Empty/nil inputs" - Handle missing or null data gracefully
  - "Boundary conditions" - Min/max values, capacity limits
  - "Concurrent access" - Thread safety with multiple goroutines
  - "Resource cleanup" - Proper cleanup on stop/error

Ask about **Complexity**:
- Question: "How complex is the internal logic of this feature?"
- Options:
  - "Simple" - Straightforward logic, only acceptance tests needed
  - "Moderate" - Some internal state, may benefit from unit tests
  - "Complex" - Significant internal logic, needs both unit and acceptance tests

## Phase 3: Technical Considerations

Ask about **Error handling** (multiSelect: true):
- Question: "How should errors be handled?"
- Options:
  - "Return error" - Return errors to caller for handling
  - "Log and continue" - Log the error but continue operation
  - "Panic on critical" - Panic on unrecoverable errors
  - "Retry with backoff" - Automatic retry for transient failures

## Phase 4: Generate Specification

Based on the gathered information, create a spec file at `specs/<feature-name>.md` using this template:

```markdown
# Spec: [Feature Name]

## Purpose
[1-2 sentences describing the main objective based on user's answer]

## Rationale
[Why this feature is needed in the context of ai-sensors]

## Package
- **Location:** `[package]/`
- **Type:** [New | Extension | Cross-cutting]

---

## Test Scenarios

### Acceptance Tests (Module Level)

#### Happy Path
[For each behavior the user described:]
1. **[Scenario name]**
   - Given: [preconditions]
   - When: [action]
   - Then: [expected result]

#### Edge Cases
[For each selected edge case:]
1. **[Edge case name]**
   - Given: [preconditions]
   - When: [action]
   - Then: [expected result]

### Unit Tests (If Complex)
[Only include if complexity is "Moderate" or "Complex"]
- [Internal behavior 1 to test]
- [Internal behavior 2 to test]

---

## Technical Considerations

### Inputs
| Input | Type | Source | Validation |
|-------|------|--------|------------|
| [input1] | [type] | [where it comes from] | [validation rules] |

### Outputs
| Output | Type | Description |
|--------|------|-------------|
| [output1] | [type] | [what it represents] |

### Processing Rules
1. [Rule derived from happy path behaviors]
2. [Additional rules]

### Alternative Paths
[Any conditional behaviors based on context]

### Error Paths
| Condition | Handling | Recovery |
|-----------|----------|----------|
| [error condition] | [handling approach from user selection] | [recovery action] |

---

## Dependencies
- **Depends on:** [packages this feature requires]
- **Used by:** [packages that will likely use this feature]
```

## Important Notes

- Keep test scenarios concrete and actionable
- If the feature extends an existing package, note what interfaces or types it should implement

</spec-skill>
