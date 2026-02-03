package buffer

import (
	"errors"
	"strings"
	"sync"
)

var ErrInvalidCapacity = errors.New("capacity must be greater than 0")

type RingBuffer struct {
	mu       sync.RWMutex
	lines    []string
	capacity int
	head     int
	count    int
	pending  string
}

func New(capacity int) (*RingBuffer, error) {
	if capacity <= 0 {
		return nil, ErrInvalidCapacity
	}
	return &RingBuffer{
		lines:    make([]string, capacity),
		capacity: capacity,
	}, nil
}

func (rb *RingBuffer) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	rb.mu.Lock()
	defer rb.mu.Unlock()

	data := rb.pending + string(p)
	rb.pending = ""

	parts := strings.Split(data, "\n")

	for i := 0; i < len(parts)-1; i++ {
		line := strings.TrimSuffix(parts[i], "\r")
		rb.addLine(line)
	}

	lastPart := parts[len(parts)-1]
	if lastPart != "" {
		rb.pending = lastPart
	}

	return len(p), nil
}

func (rb *RingBuffer) addLine(line string) {
	rb.lines[rb.head] = line
	rb.head = (rb.head + 1) % rb.capacity
	if rb.count < rb.capacity {
		rb.count++
	}
}

func (rb *RingBuffer) Lines() []string {
	return rb.getLines(rb.capacity + 1)
}

func (rb *RingBuffer) LastN(n int) []string {
	if n <= 0 {
		return []string{}
	}
	return rb.getLines(n)
}

func (rb *RingBuffer) getLines(n int) []string {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	hasPending := rb.pending != ""
	available := rb.count
	if hasPending {
		available++
	}

	if n > available {
		n = available
	}
	if n <= 0 {
		return []string{}
	}

	fromBuffer := n
	if hasPending {
		fromBuffer--
	}

	result := make([]string, 0, n)

	if fromBuffer > 0 {
		start := (rb.head - fromBuffer + rb.capacity) % rb.capacity
		for i := range fromBuffer {
			result = append(result, rb.lines[(start+i)%rb.capacity])
		}
	}

	if hasPending {
		result = append(result, rb.pending)
	}

	return result
}
