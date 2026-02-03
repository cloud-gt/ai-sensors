package runner

import (
	"bytes"
	"context"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunner_StartProcessCaptureOutputWaitForCompletion(t *testing.T) {
	var buf bytes.Buffer
	r, err := New(Config{
		Command: "echo",
		Args:    []string{"hello world"},
		Output:  &buf,
	})
	require.NoError(t, err)

	err = r.Start(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, "hello world\n", buf.String())
	assert.Equal(t, StateStopped, r.State())
}

func TestRunner_StartProcessWithArguments(t *testing.T) {
	var buf bytes.Buffer
	r, err := New(Config{
		Command: "echo",
		Args:    []string{"-n", "test"},
		Output:  &buf,
	})
	require.NoError(t, err)

	err = r.Start(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, "test", buf.String())
	assert.Equal(t, StateStopped, r.State())
}

func TestRunner_StopRunningProcessGracefullySIGTERM(t *testing.T) {
	var buf bytes.Buffer
	r, err := New(Config{
		Command:     "sleep",
		Args:        []string{"60"},
		Output:      &buf,
		StopTimeout: 1 * time.Second,
	})
	require.NoError(t, err)

	go func() {
		_ = r.Start(context.Background())
	}()

	time.Sleep(100 * time.Millisecond)
	require.Equal(t, StateRunning, r.State())

	err = r.Stop()

	assert.NoError(t, err)
	assert.Equal(t, StateStopped, r.State())
}

func TestRunner_StopWithSIGKILLAfterTimeout(t *testing.T) {
	var buf bytes.Buffer
	r, err := New(Config{
		Command:     "sh",
		Args:        []string{"-c", "trap '' TERM; sleep 60"},
		Output:      &buf,
		StopTimeout: 100 * time.Millisecond,
	})
	require.NoError(t, err)

	go func() {
		_ = r.Start(context.Background())
	}()

	time.Sleep(100 * time.Millisecond)
	require.Equal(t, StateRunning, r.State())

	start := time.Now()
	err = r.Stop()
	elapsed := time.Since(start)

	assert.NoError(t, err)
	assert.Equal(t, StateStopped, r.State())
	assert.GreaterOrEqual(t, elapsed, 100*time.Millisecond)
	assert.Less(t, elapsed, 500*time.Millisecond)
}

func TestRunner_StateTransitions(t *testing.T) {
	var buf bytes.Buffer
	r, err := New(Config{
		Command: "sleep",
		Args:    []string{"60"},
		Output:  &buf,
	})
	require.NoError(t, err)

	assert.Equal(t, StateInitial, r.State())

	go func() {
		_ = r.Start(context.Background())
	}()

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, StateRunning, r.State())

	_ = r.Stop()
	assert.Equal(t, StateStopped, r.State())
}

func TestRunner_StderrCombinedWithStdout(t *testing.T) {
	var buf bytes.Buffer
	r, err := New(Config{
		Command: "sh",
		Args:    []string{"-c", "echo stdout; echo stderr >&2"},
		Output:  &buf,
	})
	require.NoError(t, err)

	err = r.Start(context.Background())

	assert.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "stdout")
	assert.Contains(t, output, "stderr")
}

func TestRunner_DefaultStopTimeoutAppliedWhenZero(t *testing.T) {
	var buf bytes.Buffer
	r, err := New(Config{
		Command:     "sleep",
		Args:        []string{"60"},
		Output:      &buf,
		StopTimeout: 0,
	})
	require.NoError(t, err)

	go func() {
		_ = r.Start(context.Background())
	}()

	time.Sleep(100 * time.Millisecond)
	require.Equal(t, StateRunning, r.State())

	err = r.Stop()
	assert.NoError(t, err)
	assert.Equal(t, StateStopped, r.State())
}

func TestRunner_EmptyCommand(t *testing.T) {
	var buf bytes.Buffer
	_, err := New(Config{
		Command: "",
		Output:  &buf,
	})

	assert.Error(t, err)
}

func TestRunner_NilWriter(t *testing.T) {
	_, err := New(Config{
		Command: "echo",
		Args:    []string{"test"},
		Output:  nil,
	})

	assert.Error(t, err)
}

func TestRunner_CommandNotFound(t *testing.T) {
	var buf bytes.Buffer
	r, err := New(Config{
		Command: "nonexistent_command_xyz",
		Output:  &buf,
	})
	require.NoError(t, err)

	err = r.Start(context.Background())

	assert.Error(t, err)
	assert.ErrorIs(t, err, exec.ErrNotFound)
}

func TestRunner_StopCalledBeforeStart(t *testing.T) {
	var buf bytes.Buffer
	r, err := New(Config{
		Command: "echo",
		Args:    []string{"test"},
		Output:  &buf,
	})
	require.NoError(t, err)

	err = r.Stop()

	assert.NoError(t, err)
	assert.Equal(t, StateInitial, r.State())
}

func TestRunner_StopCalledMultipleTimes(t *testing.T) {
	var buf bytes.Buffer
	r, err := New(Config{
		Command: "sleep",
		Args:    []string{"60"},
		Output:  &buf,
	})
	require.NoError(t, err)

	go func() {
		_ = r.Start(context.Background())
	}()

	time.Sleep(100 * time.Millisecond)
	require.Equal(t, StateRunning, r.State())

	err1 := r.Stop()
	err2 := r.Stop()
	err3 := r.Stop()

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)
	assert.Equal(t, StateStopped, r.State())
}

func TestRunner_StartCalledAfterStop(t *testing.T) {
	var buf bytes.Buffer
	r, err := New(Config{
		Command: "echo",
		Args:    []string{"test"},
		Output:  &buf,
	})
	require.NoError(t, err)

	err = r.Start(context.Background())
	require.NoError(t, err)
	require.Equal(t, StateStopped, r.State())

	err = r.Start(context.Background())

	assert.Error(t, err)
}

func TestRunner_ContextCancellation(t *testing.T) {
	var buf bytes.Buffer
	r, err := New(Config{
		Command: "sleep",
		Args:    []string{"60"},
		Output:  &buf,
	})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err = r.Start(ctx)

	assert.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, StateStopped, r.State())
}

func TestRunner_ConcurrentStopCalls(t *testing.T) {
	var buf bytes.Buffer
	r, err := New(Config{
		Command: "sleep",
		Args:    []string{"60"},
		Output:  &buf,
	})
	require.NoError(t, err)

	go func() {
		_ = r.Start(context.Background())
	}()

	time.Sleep(100 * time.Millisecond)
	require.Equal(t, StateRunning, r.State())

	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = r.Stop()
		}()
	}
	wg.Wait()

	assert.Equal(t, StateStopped, r.State())
}

func TestRunner_ConcurrentStateReadsDuringTransitions(t *testing.T) {
	var buf bytes.Buffer
	r, err := New(Config{
		Command: "sleep",
		Args:    []string{"60"},
		Output:  &buf,
	})
	require.NoError(t, err)

	var wg sync.WaitGroup
	done := make(chan struct{})

	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					state := r.State()
					assert.Contains(t, []State{StateInitial, StateRunning, StateStopped}, state)
				}
			}
		}()
	}

	go func() {
		_ = r.Start(context.Background())
	}()

	time.Sleep(100 * time.Millisecond)
	_ = r.Stop()

	close(done)
	wg.Wait()

	assert.Equal(t, StateStopped, r.State())
}

func TestRunner_ProcessExitsWithNonZeroCode(t *testing.T) {
	var buf bytes.Buffer
	r, err := New(Config{
		Command: "sh",
		Args:    []string{"-c", "exit 1"},
		Output:  &buf,
	})
	require.NoError(t, err)

	err = r.Start(context.Background())

	assert.Error(t, err)
	var exitErr *exec.ExitError
	assert.ErrorAs(t, err, &exitErr)
	assert.Equal(t, 1, exitErr.ExitCode())
	assert.Equal(t, StateStopped, r.State())
}

func TestRunner_LargeOutputVolume(t *testing.T) {
	var buf bytes.Buffer
	r, err := New(Config{
		Command: "sh",
		Args:    []string{"-c", "for i in $(seq 1 10000); do echo \"line $i\"; done"},
		Output:  &buf,
	})
	require.NoError(t, err)

	err = r.Start(context.Background())

	assert.NoError(t, err)
	assert.Greater(t, buf.Len(), 80000)
	assert.Equal(t, StateStopped, r.State())
}

func TestRunner_NilContext(t *testing.T) {
	var buf bytes.Buffer
	r, err := New(Config{
		Command: "echo",
		Args:    []string{"test"},
		Output:  &buf,
	})
	require.NoError(t, err)

	err = r.Start(nil) //nolint:staticcheck // testing nil context behavior

	assert.Error(t, err)
}

func TestRunner_Wait(t *testing.T) {
	var buf bytes.Buffer
	r, err := New(Config{
		Command: "sh",
		Args:    []string{"-c", "sleep 0.1; echo done"},
		Output:  &buf,
	})
	require.NoError(t, err)

	go func() {
		_ = r.Start(context.Background())
	}()

	time.Sleep(50 * time.Millisecond)
	require.Equal(t, StateRunning, r.State())

	err = r.Wait()

	assert.NoError(t, err)
	assert.Equal(t, StateStopped, r.State())
	assert.Equal(t, "done\n", buf.String())
}
