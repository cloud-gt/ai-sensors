package command

import (
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_CreateNewCommandDefinition(t *testing.T) {
	store := NewStore(NewMemoryRepository())

	cmd, err := store.Create(Command{
		Name:    "test-command",
		Command: "echo hello",
		WorkDir: "/tmp",
	})

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, cmd.ID)
	assert.Equal(t, "test-command", cmd.Name)
	assert.Equal(t, "echo hello", cmd.Command)
}

func TestStore_RetrieveExistingCommandDefinition(t *testing.T) {
	store := NewStore(NewMemoryRepository())
	created, err := store.Create(Command{
		Name:    "test-command",
		Command: "echo hello",
		WorkDir: "/tmp",
	})
	require.NoError(t, err)

	retrieved, err := store.Get(created.ID)

	require.NoError(t, err)
	assert.Equal(t, created, retrieved)
}

func TestStore_UpdateExistingCommandDefinition(t *testing.T) {
	store := NewStore(NewMemoryRepository())
	created, err := store.Create(Command{
		Name:    "original-name",
		Command: "original-command",
		WorkDir: "/tmp",
	})
	require.NoError(t, err)

	err = store.Update(Command{
		ID:      created.ID,
		Name:    "updated-name",
		Command: "updated-command",
		WorkDir: "/tmp",
	})
	require.NoError(t, err)

	retrieved, err := store.Get(created.ID)
	require.NoError(t, err)
	assert.Equal(t, Command{
		ID:      created.ID,
		Name:    "updated-name",
		Command: "updated-command",
		WorkDir: "/tmp",
	}, retrieved)
}

func TestStore_DeleteExistingCommandDefinition(t *testing.T) {
	store := NewStore(NewMemoryRepository())
	created, err := store.Create(Command{
		Name:    "test-command",
		Command: "echo hello",
		WorkDir: "/tmp",
	})
	require.NoError(t, err)

	err = store.Delete(created.ID)
	require.NoError(t, err)

	_, err = store.Get(created.ID)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestStore_ListAllCommandDefinitions(t *testing.T) {
	store := NewStore(NewMemoryRepository())
	_, err := store.Create(Command{Name: "cmd1", Command: "echo 1", WorkDir: "/tmp"})
	require.NoError(t, err)
	_, err = store.Create(Command{Name: "cmd2", Command: "echo 2", WorkDir: "/tmp"})
	require.NoError(t, err)
	_, err = store.Create(Command{Name: "cmd3", Command: "echo 3", WorkDir: "/tmp"})
	require.NoError(t, err)

	commands, err := store.List()

	require.NoError(t, err)
	assert.Len(t, commands, 3)
}

func TestStore_CreateWithEmptyName(t *testing.T) {
	store := NewStore(NewMemoryRepository())

	_, err := store.Create(Command{
		Name:    "",
		Command: "echo hello",
		WorkDir: "/tmp",
	})

	assert.ErrorIs(t, err, ErrEmptyName)
}

func TestStore_CreateWithEmptyCommand(t *testing.T) {
	store := NewStore(NewMemoryRepository())

	_, err := store.Create(Command{
		Name:    "test-command",
		Command: "",
		WorkDir: "/tmp",
	})

	assert.ErrorIs(t, err, ErrEmptyCommand)
}

func TestStore_CreateWithEmptyWorkDir(t *testing.T) {
	store := NewStore(NewMemoryRepository())

	_, err := store.Create(Command{
		Name:    "test-command",
		Command: "echo hello",
		WorkDir: "",
	})

	assert.ErrorIs(t, err, ErrEmptyWorkDir)
}

func TestStore_CreateWithWorkDir(t *testing.T) {
	store := NewStore(NewMemoryRepository())

	cmd, err := store.Create(Command{
		Name:    "test-command",
		Command: "echo hello",
		WorkDir: "/tmp",
	})

	require.NoError(t, err)
	assert.Equal(t, "/tmp", cmd.WorkDir)
}

func TestStore_GetNonExistentCommand(t *testing.T) {
	store := NewStore(NewMemoryRepository())

	_, err := store.Get(uuid.New())

	assert.ErrorIs(t, err, ErrNotFound)
}

func TestStore_UpdateNonExistentCommand(t *testing.T) {
	store := NewStore(NewMemoryRepository())

	err := store.Update(Command{
		ID:      uuid.New(),
		Name:    "test",
		Command: "echo",
	})

	assert.ErrorIs(t, err, ErrNotFound)
}

func TestStore_DeleteNonExistentCommand(t *testing.T) {
	store := NewStore(NewMemoryRepository())

	err := store.Delete(uuid.New())

	assert.ErrorIs(t, err, ErrNotFound)
}

func TestStore_ConcurrentAccessSafety(t *testing.T) {
	store := NewStore(NewMemoryRepository())
	var wg sync.WaitGroup

	for i := range 10 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			cmd, err := store.Create(Command{
				Name:    "cmd",
				Command: "echo",
				WorkDir: "/tmp",
			})
			if err != nil {
				return
			}

			_, _ = store.Get(cmd.ID)
			_ = store.Update(Command{
				ID:      cmd.ID,
				Name:    "updated",
				Command: "echo updated",
				WorkDir: "/tmp",
			})
			_, _ = store.List()
			_ = store.Delete(cmd.ID)
		}(i)
	}

	wg.Wait()
}
