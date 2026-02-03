package manager

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/cloud-gt/ai-sensors/command"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestStore() *command.Store {
	return command.NewStore(command.NewMemoryRepository())
}

func TestManager_StartCommandAndReadOutput(t *testing.T) {
	store := newTestStore()
	cmd, err := store.Create(command.Command{
		Name:    "echo-test",
		Command: "echo hello",
	})
	require.NoError(t, err)

	m := New(store)

	started, err := m.Start(context.Background(), cmd.ID)
	require.NoError(t, err)
	assert.True(t, started)

	time.Sleep(100 * time.Millisecond)

	output, err := m.Output(cmd.ID)
	require.NoError(t, err)
	assert.Contains(t, output, "hello")
}

func TestManager_FullLifecycle(t *testing.T) {
	store := newTestStore()
	cmd, err := store.Create(command.Command{
		Name:    "sleep-test",
		Command: "sleep 60",
	})
	require.NoError(t, err)

	m := New(store)

	started, err := m.Start(context.Background(), cmd.ID)
	require.NoError(t, err)
	assert.True(t, started)

	time.Sleep(100 * time.Millisecond)

	_, err = m.Output(cmd.ID)
	assert.NoError(t, err)

	status, err := m.Status(cmd.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusRunning, status)

	err = m.Stop(cmd.ID)
	require.NoError(t, err)

	status, err = m.Status(cmd.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusStopped, status)
}

func TestManager_RunMultipleCommandsConcurrently(t *testing.T) {
	store := newTestStore()
	cmd1, err := store.Create(command.Command{
		Name:    "cmd1",
		Command: "echo first",
	})
	require.NoError(t, err)

	cmd2, err := store.Create(command.Command{
		Name:    "cmd2",
		Command: "echo second",
	})
	require.NoError(t, err)

	m := New(store)

	started1, err := m.Start(context.Background(), cmd1.ID)
	require.NoError(t, err)
	assert.True(t, started1)

	started2, err := m.Start(context.Background(), cmd2.ID)
	require.NoError(t, err)
	assert.True(t, started2)

	time.Sleep(100 * time.Millisecond)

	output1, err := m.Output(cmd1.ID)
	require.NoError(t, err)
	assert.Contains(t, output1, "first")

	output2, err := m.Output(cmd2.ID)
	require.NoError(t, err)
	assert.Contains(t, output2, "second")
}

func TestManager_AutoCleanupOnNaturalTermination(t *testing.T) {
	store := newTestStore()
	cmd, err := store.Create(command.Command{
		Name:    "short-lived",
		Command: "echo hello",
	})
	require.NoError(t, err)

	m := New(store)

	started, err := m.Start(context.Background(), cmd.ID)
	require.NoError(t, err)
	assert.True(t, started)

	time.Sleep(200 * time.Millisecond)

	status, err := m.Status(cmd.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusStopped, status)

	output, err := m.Output(cmd.ID)
	require.NoError(t, err)
	assert.Contains(t, output, "hello")
}

func TestManager_AccurateStatusTracking(t *testing.T) {
	store := newTestStore()
	cmd, err := store.Create(command.Command{
		Name:    "sleep-cmd",
		Command: "sleep 60",
	})
	require.NoError(t, err)

	m := New(store)

	_, err = m.Status(cmd.ID)
	assert.ErrorIs(t, err, ErrNotRunning)

	started, err := m.Start(context.Background(), cmd.ID)
	require.NoError(t, err)
	assert.True(t, started)

	time.Sleep(100 * time.Millisecond)

	status, err := m.Status(cmd.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusRunning, status)

	err = m.Stop(cmd.ID)
	require.NoError(t, err)

	status, err = m.Status(cmd.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusStopped, status)
}

func TestManager_StartUnknownCommandID(t *testing.T) {
	store := newTestStore()
	m := New(store)

	unknownID := uuid.New()
	_, err := m.Start(context.Background(), unknownID)

	assert.ErrorIs(t, err, ErrCommandNotFound)
}

func TestManager_StopUnknownCommandID(t *testing.T) {
	store := newTestStore()
	m := New(store)

	unknownID := uuid.New()
	err := m.Stop(unknownID)

	assert.ErrorIs(t, err, ErrNotRunning)
}

func TestManager_OutputForUnknownCommandID(t *testing.T) {
	store := newTestStore()
	m := New(store)

	unknownID := uuid.New()
	_, err := m.Output(unknownID)

	assert.ErrorIs(t, err, ErrNotRunning)
}

func TestManager_DoubleStartIdempotent(t *testing.T) {
	store := newTestStore()
	cmd, err := store.Create(command.Command{
		Name:    "sleep-test",
		Command: "sleep 60",
	})
	require.NoError(t, err)

	m := New(store)

	started1, err := m.Start(context.Background(), cmd.ID)
	require.NoError(t, err)
	assert.True(t, started1)

	time.Sleep(100 * time.Millisecond)

	started2, err := m.Start(context.Background(), cmd.ID)
	require.NoError(t, err)
	assert.False(t, started2)

	_ = m.Stop(cmd.ID)
}

func TestManager_DoubleStopIdempotent(t *testing.T) {
	store := newTestStore()
	cmd, err := store.Create(command.Command{
		Name:    "sleep-test",
		Command: "sleep 60",
	})
	require.NoError(t, err)

	m := New(store)

	started, err := m.Start(context.Background(), cmd.ID)
	require.NoError(t, err)
	assert.True(t, started)

	time.Sleep(100 * time.Millisecond)

	err = m.Stop(cmd.ID)
	require.NoError(t, err)

	err = m.Stop(cmd.ID)
	assert.NoError(t, err)

	err = m.Stop(cmd.ID)
	assert.NoError(t, err)
}

func TestManager_ConcurrentAccessFromMultipleGoroutines(t *testing.T) {
	store := newTestStore()
	cmd, err := store.Create(command.Command{
		Name:    "sleep-cmd",
		Command: "sleep 60",
	})
	require.NoError(t, err)

	m := New(store)

	started, err := m.Start(context.Background(), cmd.ID)
	require.NoError(t, err)
	assert.True(t, started)

	time.Sleep(100 * time.Millisecond)

	var wg sync.WaitGroup
	for range 10 {
		wg.Add(4)
		go func() {
			defer wg.Done()
			_, _ = m.Status(cmd.ID)
		}()
		go func() {
			defer wg.Done()
			_, _ = m.Output(cmd.ID)
		}()
		go func() {
			defer wg.Done()
			_, _ = m.Start(context.Background(), cmd.ID)
		}()
		go func() {
			defer wg.Done()
			_ = m.Stop(cmd.ID)
		}()
	}
	wg.Wait()
}

func TestManager_ResourceCleanupOnShutdown(t *testing.T) {
	store := newTestStore()
	cmd1, err := store.Create(command.Command{
		Name:    "sleep1",
		Command: "sleep 60",
	})
	require.NoError(t, err)

	cmd2, err := store.Create(command.Command{
		Name:    "sleep2",
		Command: "sleep 60",
	})
	require.NoError(t, err)

	m := New(store)

	_, err = m.Start(context.Background(), cmd1.ID)
	require.NoError(t, err)
	_, err = m.Start(context.Background(), cmd2.ID)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	status1, _ := m.Status(cmd1.ID)
	status2, _ := m.Status(cmd2.ID)
	require.Equal(t, StatusRunning, status1)
	require.Equal(t, StatusRunning, status2)

	err = m.Shutdown(context.Background())
	require.NoError(t, err)

	status1, _ = m.Status(cmd1.ID)
	status2, _ = m.Status(cmd2.ID)
	assert.Equal(t, StatusStopped, status1)
	assert.Equal(t, StatusStopped, status2)
}

func TestManager_WithBufferCapacity(t *testing.T) {
	store := newTestStore()
	cmd, err := store.Create(command.Command{
		Name:    "multi-line",
		Command: "sh -c 'for i in 1 2 3 4 5; do echo line$i; done'",
	})
	require.NoError(t, err)

	m := New(store, WithBufferCapacity(3))

	started, err := m.Start(context.Background(), cmd.ID)
	require.NoError(t, err)
	assert.True(t, started)

	time.Sleep(200 * time.Millisecond)

	output, err := m.Output(cmd.ID)
	require.NoError(t, err)
	assert.Len(t, output, 3)
	assert.Equal(t, []string{"line3", "line4", "line5"}, output)
}

func TestManager_OutputLastN(t *testing.T) {
	store := newTestStore()
	cmd, err := store.Create(command.Command{
		Name:    "multi-line",
		Command: "sh -c 'for i in 1 2 3 4 5; do echo line$i; done'",
	})
	require.NoError(t, err)

	m := New(store)

	started, err := m.Start(context.Background(), cmd.ID)
	require.NoError(t, err)
	assert.True(t, started)

	time.Sleep(200 * time.Millisecond)

	output, err := m.OutputLastN(cmd.ID, 2)
	require.NoError(t, err)
	assert.Len(t, output, 2)
	assert.Equal(t, []string{"line4", "line5"}, output)
}

func TestManager_OutputLastNUnknownCommand(t *testing.T) {
	store := newTestStore()
	m := New(store)

	unknownID := uuid.New()
	_, err := m.OutputLastN(unknownID, 5)

	assert.ErrorIs(t, err, ErrNotRunning)
}

func TestManager_ContextCancellationStopsCommand(t *testing.T) {
	store := newTestStore()
	cmd, err := store.Create(command.Command{
		Name:    "sleep-cmd",
		Command: "sleep 60",
	})
	require.NoError(t, err)

	m := New(store)

	ctx, cancel := context.WithCancel(context.Background())

	started, err := m.Start(ctx, cmd.ID)
	require.NoError(t, err)
	assert.True(t, started)

	time.Sleep(100 * time.Millisecond)

	status, err := m.Status(cmd.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusRunning, status)

	cancel()

	time.Sleep(200 * time.Millisecond)

	status, err = m.Status(cmd.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusStopped, status)
}

func TestManager_ShutdownRespectsContextCancellation(t *testing.T) {
	store := newTestStore()
	cmd, err := store.Create(command.Command{
		Name:    "trap-sigterm",
		Command: "sh -c 'trap \"\" TERM; sleep 60'",
	})
	require.NoError(t, err)

	m := New(store)

	_, err = m.Start(context.Background(), cmd.ID)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err = m.Shutdown(ctx)

	assert.ErrorIs(t, err, context.Canceled)
}
