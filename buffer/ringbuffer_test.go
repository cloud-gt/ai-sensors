package buffer

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRingBuffer_BasicWriteAndRead(t *testing.T) {
	rb, err := New(10)
	require.NoError(t, err)

	n, err := rb.Write([]byte("line1\nline2\nline3\n"))

	require.NoError(t, err)
	assert.Equal(t, 18, n)
	assert.Equal(t, []string{"line1", "line2", "line3"}, rb.Lines())
}

func TestRingBuffer_CircularBehaviorOverflow(t *testing.T) {
	rb, err := New(3)
	require.NoError(t, err)

	_, _ = rb.Write([]byte("a\nb\nc\nd\ne\n"))

	assert.Equal(t, []string{"c", "d", "e"}, rb.Lines())
}

func TestRingBuffer_LastNReturnsSubset(t *testing.T) {
	rb, err := New(20)
	require.NoError(t, err)
	for i := range 10 {
		fmt.Fprintf(rb, "line%d\n", i+1)
	}

	result := rb.LastN(3)

	assert.Equal(t, []string{"line8", "line9", "line10"}, result)
}

func TestRingBuffer_MultipleLinesInSingleWrite(t *testing.T) {
	rb, err := New(10)
	require.NoError(t, err)

	_, _ = rb.Write([]byte("a\nb\nc\nd\n"))

	assert.Equal(t, []string{"a", "b", "c", "d"}, rb.Lines())
}

func TestRingBuffer_EmptyWrite(t *testing.T) {
	rb, err := New(10)
	require.NoError(t, err)
	_, _ = rb.Write([]byte("existing\n"))

	n, err := rb.Write([]byte(""))

	assert.Equal(t, 0, n)
	assert.NoError(t, err)
	assert.Equal(t, []string{"existing"}, rb.Lines())
}

func TestRingBuffer_MinimumCapacity(t *testing.T) {
	rb, err := New(1)
	require.NoError(t, err)

	_, _ = rb.Write([]byte("first\nsecond\nthird\n"))

	assert.Equal(t, []string{"third"}, rb.Lines())
}

func TestRingBuffer_LastNGreaterThanLines(t *testing.T) {
	rb, err := New(10)
	require.NoError(t, err)
	_, _ = rb.Write([]byte("a\nb\nc\n"))

	result := rb.LastN(100)

	assert.Equal(t, []string{"a", "b", "c"}, result)
}

func TestRingBuffer_LastNZero(t *testing.T) {
	rb, err := New(10)
	require.NoError(t, err)
	_, _ = rb.Write([]byte("a\nb\nc\n"))

	result := rb.LastN(0)

	assert.Equal(t, []string{}, result)
}

func TestRingBuffer_ConcurrentWriteAccess(t *testing.T) {
	rb, err := New(50)
	require.NoError(t, err)

	var wg sync.WaitGroup
	for i := range 10 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := range 100 {
				fmt.Fprintf(rb, "goroutine%d-line%d\n", id, j)
			}
		}(i)
	}
	wg.Wait()

	lines := rb.Lines()
	assert.Equal(t, 50, len(lines))
}

func TestRingBuffer_ConcurrentReadWriteAccess(t *testing.T) {
	rb, err := New(100)
	require.NoError(t, err)

	var writerWg sync.WaitGroup
	done := make(chan struct{})

	for i := range 5 {
		writerWg.Add(1)
		go func(id int) {
			defer writerWg.Done()
			for j := range 50 {
				fmt.Fprintf(rb, "writer%d-line%d\n", id, j)
			}
		}(i)
	}

	var readerWg sync.WaitGroup
	for range 5 {
		readerWg.Add(1)
		go func() {
			defer readerWg.Done()
			for {
				select {
				case <-done:
					return
				default:
					_ = rb.Lines()
				}
			}
		}()
	}

	writerWg.Wait()
	close(done)
	readerWg.Wait()

	lines := rb.Lines()
	assert.LessOrEqual(t, len(lines), 100)
}

func TestRingBuffer_IncompleteLineNoTrailingNewline(t *testing.T) {
	rb, err := New(10)
	require.NoError(t, err)

	_, _ = rb.Write([]byte("line1\nline2"))

	assert.Equal(t, []string{"line1", "line2"}, rb.Lines())
}

func TestRingBuffer_FragmentedLineWrite(t *testing.T) {
	rb, err := New(10)
	require.NoError(t, err)

	_, _ = rb.Write([]byte("hel"))
	_, _ = rb.Write([]byte("lo\n"))

	assert.Equal(t, []string{"hello"}, rb.Lines())
}

func TestRingBuffer_CRLFLineEnding(t *testing.T) {
	rb, err := New(10)
	require.NoError(t, err)

	_, _ = rb.Write([]byte("line1\r\nline2\r\nline3\r\n"))

	assert.Equal(t, []string{"line1", "line2", "line3"}, rb.Lines())
}

func TestRingBuffer_InvalidCapacity(t *testing.T) {
	_, err := New(0)
	assert.Error(t, err)

	_, err = New(-5)
	assert.Error(t, err)
}

func TestRingBuffer_LastNNegative(t *testing.T) {
	rb, err := New(10)
	require.NoError(t, err)
	_, _ = rb.Write([]byte("a\nb\nc\n"))

	result := rb.LastN(-1)

	assert.Equal(t, []string{}, result)
}
