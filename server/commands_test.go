package server

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/cloud-gt/ai-sensors/command"
	"github.com/cloud-gt/ai-sensors/manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestServer() (*Server, *TestClient) {
	store := command.NewStore(command.NewMemoryRepository())
	mgr := manager.New(store)
	srv := New(store, mgr)
	return srv, newTestClient(srv)
}

func TestListCommands_Empty(t *testing.T) {
	_, tc := newTestServer()

	commands, resp := tc.ListCommands()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Empty(t, commands)
}

func TestCreateCommand(t *testing.T) {
	_, tc := newTestServer()

	cmd, resp := tc.CreateCommand("test-cmd", "echo hello")

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NotNil(t, cmd)
	assert.Equal(t, "test-cmd", cmd.Name)
	assert.Equal(t, "echo hello", cmd.Command)
	assert.NotEmpty(t, cmd.ID)
}

func TestGetCommandByID(t *testing.T) {
	_, tc := newTestServer()

	created, _ := tc.CreateCommand("test-cmd", "echo hello")
	require.NotNil(t, created)

	cmd, resp := tc.GetCommand(created.ID)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, cmd)
	assert.Equal(t, created.ID, cmd.ID)
	assert.Equal(t, "test-cmd", cmd.Name)
}

func TestListCommands_NonEmpty(t *testing.T) {
	_, tc := newTestServer()

	tc.CreateCommand("cmd-a", "echo a")
	tc.CreateCommand("cmd-b", "echo b")

	commands, resp := tc.ListCommands()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Len(t, commands, 2)
}

func TestDeleteCommand(t *testing.T) {
	_, tc := newTestServer()

	created, _ := tc.CreateCommand("test-cmd", "echo hello")
	require.NotNil(t, created)

	resp := tc.DeleteCommand(created.ID)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	_, resp = tc.GetCommand(created.ID)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestStartCommand(t *testing.T) {
	srv, tc := newTestServer()

	created, _ := tc.CreateCommand("test-cmd", "sleep 60")
	require.NotNil(t, created)

	started, resp := tc.StartCommand(created.ID)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, started)

	_ = srv.manager.Stop(created.ID)
}

func TestStartCommand_AlreadyRunning(t *testing.T) {
	srv, tc := newTestServer()

	created, _ := tc.CreateCommand("test-cmd", "sleep 60")
	require.NotNil(t, created)

	tc.StartCommand(created.ID)
	time.Sleep(100 * time.Millisecond)

	started, resp := tc.StartCommand(created.ID)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.False(t, started)

	_ = srv.manager.Stop(created.ID)
}

func TestStopCommand(t *testing.T) {
	_, tc := newTestServer()

	created, _ := tc.CreateCommand("test-cmd", "sleep 60")
	require.NotNil(t, created)

	tc.StartCommand(created.ID)
	time.Sleep(100 * time.Millisecond)

	resp := tc.StopCommand(created.ID)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetCommandStatus(t *testing.T) {
	srv, tc := newTestServer()

	created, _ := tc.CreateCommand("test-cmd", "sleep 60")
	require.NotNil(t, created)

	tc.StartCommand(created.ID)
	time.Sleep(100 * time.Millisecond)

	status, resp := tc.GetStatus(created.ID)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "running", status)

	_ = srv.manager.Stop(created.ID)
}

func TestGetFullOutput(t *testing.T) {
	_, tc := newTestServer()

	created, _ := tc.CreateCommand("test-cmd", "echo hello")
	require.NotNil(t, created)

	tc.StartCommand(created.ID)
	time.Sleep(200 * time.Millisecond)

	lines, resp := tc.GetOutput(created.ID)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, lines, "hello")
}

func TestGetLastNLinesOutput(t *testing.T) {
	_, tc := newTestServer()

	created, _ := tc.CreateCommand("test-cmd", "sh -c 'for i in 1 2 3 4 5; do echo line$i; done'")
	require.NotNil(t, created)

	tc.StartCommand(created.ID)
	time.Sleep(200 * time.Millisecond)

	lines, resp := tc.GetOutputLastN(created.ID, 2)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Len(t, lines, 2)
	assert.Equal(t, []string{"line4", "line5"}, lines)
}

func TestFullE2ELifecycle(t *testing.T) {
	_, tc := newTestServer()

	created, resp := tc.CreateCommand("e2e-cmd", "echo lifecycle")
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	_, resp = tc.StartCommand(created.ID)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	time.Sleep(200 * time.Millisecond)

	_, resp = tc.GetStatus(created.ID)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	lines, resp := tc.GetOutput(created.ID)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, lines, "lifecycle")

	resp = tc.StopCommand(created.ID)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	resp = tc.DeleteCommand(created.ID)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestCreateCommand_InvalidJSON(t *testing.T) {
	_, tc := newTestServer()

	resp := tc.Do(http.MethodPost, "/commands", "invalid")

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateCommand_MissingRequiredFields(t *testing.T) {
	_, tc := newTestServer()

	resp := tc.Do(http.MethodPost, "/commands", map[string]string{"name": "test-cmd"})

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateCommand_DuplicateName(t *testing.T) {
	_, tc := newTestServer()

	tc.CreateCommand("test-cmd", "echo hello")
	_, resp := tc.CreateCommand("test-cmd", "echo world")

	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestGetUnknownCommand(t *testing.T) {
	_, tc := newTestServer()

	resp := tc.Do(http.MethodGet, "/commands/01939f00-0000-7000-8000-000000000000", nil)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestDeleteUnknownCommand(t *testing.T) {
	_, tc := newTestServer()

	resp := tc.Do(http.MethodDelete, "/commands/01939f00-0000-7000-8000-000000000000", nil)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestStartUnknownCommand(t *testing.T) {
	_, tc := newTestServer()

	resp := tc.Do(http.MethodPost, "/commands/01939f00-0000-7000-8000-000000000000/start", nil)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestInvalidUUIDFormat(t *testing.T) {
	_, tc := newTestServer()

	resp := tc.Do(http.MethodGet, "/commands/not-a-uuid", nil)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestStopNeverStartedCommand(t *testing.T) {
	_, tc := newTestServer()

	created, _ := tc.CreateCommand("test-cmd", "echo hello")
	require.NotNil(t, created)

	resp := tc.StopCommand(created.ID)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetStatusNeverStartedCommand(t *testing.T) {
	_, tc := newTestServer()

	created, _ := tc.CreateCommand("test-cmd", "echo hello")
	require.NotNil(t, created)

	_, resp := tc.GetStatus(created.ID)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetOutputNeverStartedCommand(t *testing.T) {
	_, tc := newTestServer()

	created, _ := tc.CreateCommand("test-cmd", "echo hello")
	require.NotNil(t, created)

	_, resp := tc.GetOutput(created.ID)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestInvalidLinesParameter(t *testing.T) {
	srv, tc := newTestServer()

	created, _ := tc.CreateCommand("test-cmd", "sleep 60")
	require.NotNil(t, created)

	tc.StartCommand(created.ID)
	time.Sleep(100 * time.Millisecond)

	resp := tc.Do(http.MethodGet, "/commands/"+created.ID.String()+"/output?lines=abc", nil)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	_ = srv.manager.Stop(created.ID)
}

func TestNegativeLinesParameter(t *testing.T) {
	srv, tc := newTestServer()

	created, _ := tc.CreateCommand("test-cmd", "sleep 60")
	require.NotNil(t, created)

	tc.StartCommand(created.ID)
	time.Sleep(100 * time.Millisecond)

	resp := tc.Do(http.MethodGet, "/commands/"+created.ID.String()+"/output?lines=-5", nil)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	_ = srv.manager.Stop(created.ID)
}

func TestConcurrentRequestsToSameCommand(t *testing.T) {
	_, tc := newTestServer()

	created, _ := tc.CreateCommand("test-cmd", "sleep 60")
	require.NotNil(t, created)

	tc.StartCommand(created.ID)
	time.Sleep(100 * time.Millisecond)

	var wg sync.WaitGroup
	for range 10 {
		wg.Add(4)
		go func() {
			defer wg.Done()
			tc.GetStatus(created.ID)
		}()
		go func() {
			defer wg.Done()
			tc.GetOutput(created.ID)
		}()
		go func() {
			defer wg.Done()
			tc.StartCommand(created.ID)
		}()
		go func() {
			defer wg.Done()
			tc.StopCommand(created.ID)
		}()
	}
	wg.Wait()
}

func TestDeleteRunningCommand(t *testing.T) {
	srv, tc := newTestServer()

	created, _ := tc.CreateCommand("test-cmd", "sleep 60")
	require.NotNil(t, created)

	tc.StartCommand(created.ID)
	time.Sleep(100 * time.Millisecond)

	resp := tc.DeleteCommand(created.ID)

	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	_ = srv.manager.Stop(created.ID)
}

func TestServerShutdownWithRunningCommands(t *testing.T) {
	srv, tc := newTestServer()

	created, _ := tc.CreateCommand("test-cmd", "sleep 60")
	require.NotNil(t, created)

	tc.StartCommand(created.ID)
	time.Sleep(100 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := srv.Shutdown(ctx)
	assert.NoError(t, err)
}
