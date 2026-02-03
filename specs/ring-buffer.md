# Spec: Ring Buffer

## Purpose
Store the N most recent lines of command output in memory using a circular buffer, ensuring bounded memory consumption regardless of the volume of data produced.

## Rationale
Watch/continuous commands can produce unlimited output volume. The ring buffer guarantees that only the most recent lines are kept, preventing memory explosion while maintaining a usable history for code agents.

## Package
- **Location:** `buffer/`
- **Type:** New

---

## Test Scenarios

### Acceptance Tests (Module Level)

#### Happy Path

1. **Basic write and read**
   - Given: A RingBuffer with capacity of 10 lines
   - When: I write "line1\nline2\nline3" via Write()
   - Then: Lines() returns ["line1", "line2", "line3"]

2. **Circular behavior (overflow)**
   - Given: A RingBuffer with capacity of 3 lines
   - When: I write 5 successive lines ("a", "b", "c", "d", "e")
   - Then: Lines() returns ["c", "d", "e"] (the first 2 were overwritten)

3. **LastN(n) returns a subset**
   - Given: A RingBuffer containing 10 lines
   - When: I call LastN(3)
   - Then: I receive only the last 3 lines in order

4. **Multiple lines in a single write**
   - Given: An empty RingBuffer with capacity of 10
   - When: I write "a\nb\nc\nd" in a single Write() call
   - Then: Lines() returns ["a", "b", "c", "d"] correctly parsed

#### Edge Cases

1. **Empty write**
   - Given: A RingBuffer with existing content
   - When: I write an empty string via Write()
   - Then: The buffer remains unchanged, Write returns (0, nil)

2. **Minimum capacity (1 line)**
   - Given: A RingBuffer with capacity of 1
   - When: I write multiple lines
   - Then: Only the last line is kept

3. **LastN with n > number of lines**
   - Given: A RingBuffer containing 3 lines
   - When: I call LastN(100)
   - Then: I receive the 3 available lines (no error)

4. **LastN with n = 0**
   - Given: A RingBuffer containing lines
   - When: I call LastN(0)
   - Then: I receive an empty slice

5. **Concurrent write access**
   - Given: A shared RingBuffer
   - When: 10 goroutines simultaneously write 100 lines each
   - Then: No data race, the buffer contains exactly `capacity` lines at the end

6. **Concurrent read/write access**
   - Given: A shared RingBuffer
   - When: Goroutines write while others read via Lines()
   - Then: No data race, reads return consistent snapshots

7. **Incomplete line (no trailing newline)**
   - Given: An empty RingBuffer
   - When: I write "line1\nline2" (without trailing \n)
   - Then: Lines() returns ["line1", "line2"] - the incomplete line is included

8. **Fragmented line write**
   - Given: An empty RingBuffer
   - When: I write "hel" then "lo\n" in two Write() calls
   - Then: Lines() returns ["hello"] - fragments are assembled

### Unit Tests (Internal Logic)

- Verify that the circular index wraps correctly at capacity
- Test line parsing with different separators (\n, \r\n)
- Validate that the mutex correctly protects internal state
- Test the pending line buffer behavior (incomplete line)

---

## Technical Considerations

### Inputs
| Input | Type | Source | Validation |
|-------|------|--------|------------|
| capacity | int | Constructor | Must be > 0, otherwise error |
| data | []byte | io.Writer.Write() | Accepts anything, including empty |
| n | int | LastN(n) | If < 0, treat as 0 |

### Outputs
| Output | Type | Description |
|--------|------|-------------|
| lines | []string | Slice of stored lines, in chronological order |
| n, err | (int, error) | Standard io.Writer return |

### Processing Rules
1. Each `\n` delimits a complete line
2. Lines are stored in arrival order
3. When the buffer is full, the oldest line is overwritten
4. Data between the last `\n` and end of Write is kept as "pending" until the next `\n`
5. Lines() and LastN() include the current pending line (if non-empty)

### Technical Decisions

#### Newline Handling
**Decision:** Split on `\n` and strip trailing `\r` characters.

**Rationale:**
- Go's `bufio.Scanner` with `ScanLines` handles both `\n` and `\r\n` automatically, but it requires an `io.Reader` interface, not `io.Writer`
- Since our RingBuffer implements `io.Writer` (receives `[]byte` via `Write()`), we handle line splitting manually
- The pragmatic approach: split on `\n`, then `strings.TrimSuffix(line, "\r")` to handle Windows-style CRLF
- This covers Unix (`\n`), Windows (`\r\n`), and old Mac (`\r` followed by `\n` in next write) line endings

**Limitation:** Pure `\r` (classic Mac OS pre-X) without `\n` is not treated as a line separator. This is acceptable for modern shell output.

### Interface
```go
type RingBuffer struct {
    // internal fields
}

func New(capacity int) (*RingBuffer, error)
func (rb *RingBuffer) Write(p []byte) (n int, err error)  // implements io.Writer
func (rb *RingBuffer) Lines() []string
func (rb *RingBuffer) LastN(n int) []string
```

### Error Paths
| Condition | Handling | Recovery |
|-----------|----------|----------|
| capacity <= 0 | New() returns error | Caller must provide a valid capacity |
| nil receiver | Panic (standard Go behavior) | N/A |

---

## Dependencies
- **Depends on:** None (standalone module)
- **Used by:** `manager/` (to capture process output)

---

## Future Extension (F6)
This spec covers storage-only functionality. Feature 6 (Line Pipeline) will extend RingBuffer with an observable pattern (`OnLine()` callbacks or `Subscribe()` channels) for real-time line processing. The current design is compatible with this extension - `Write()` will simply notify observers after storing each complete line.
