# Issue: Introduce Source abstraction to decouple data producers from ring buffers

## Summary

Currently, the ring buffer mechanism is tightly coupled to process execution via the Runner. The only way to write data to a ring buffer is through a command that spawns a process. We should introduce a "Source" abstraction that allows different types of data producers to feed into the ring buffer system.

## Problem

The current architecture has these limitations:

1. **Tight coupling in Manager** (`manager/manager.go:85-98`): Every runner creates its own buffer, and only process stdout/stderr can write to it
2. **Single data source pattern**: Only `Runner` (process execution) can produce data for buffers
3. **No flexibility for other input types**: Cannot use the ring buffer mechanism with other data sources

## Use Cases

Examples of data sources that should be able to use the ring buffer mechanism:

- **WebSocket connections**: Receive streaming data from external services
- **Sensor inputs**: Direct sensor data feeds
- **Log file tailing**: Watch and buffer log file changes
- **Message queues**: Consume messages from RabbitMQ, Kafka, etc.
- **HTTP streaming**: Server-Sent Events, chunked responses
- **Aggregated sources**: Combine multiple sources into a single buffer

## Proposed Solution

Introduce a `Source` interface that abstracts data producers:

```go
// Source represents any data producer that can write to a buffer
type Source interface {
    // Start begins producing data, writing to the provided writer
    Start(ctx context.Context, output io.Writer) error
    // Stop gracefully stops the source
    Stop() error
    // Status returns the current state of the source
    Status() SourceStatus
}
```

The existing `Runner` would become one implementation of `Source` (e.g., `ProcessSource`), and new implementations could be added:

```
Source (interface)
├── ProcessSource (current runner behavior)
├── WebSocketSource
├── FileWatchSource
└── ... (extensible)
```

## Benefits

- **Separation of concerns**: Data production is decoupled from buffering
- **Extensibility**: New source types can be added without modifying core buffer logic
- **Composability**: Multiple sources could potentially write to the same buffer
- **Testability**: Sources can be mocked independently of buffers

## Related Files

- `buffer/ringbuffer.go` - Already implements `io.Writer`, no changes needed
- `runner/runner.go` - Would be wrapped/refactored into a `ProcessSource`
- `manager/manager.go` - Would manage `Source` instances instead of `Runner` directly

## Notes

This is a foundational change that would enable more flexible data ingestion patterns while preserving the existing ring buffer mechanism as the processing entry point.
