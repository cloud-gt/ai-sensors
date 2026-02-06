package command

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONFileRepository_SaveAndLoadRoundtrip(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "commands.json")
	repo := NewJSONFileRepository(path)

	id1 := uuid.MustParse("01234567-89ab-cdef-0123-456789abcdef")
	id2 := uuid.MustParse("fedcba98-7654-3210-fedc-ba9876543210")
	commands := []Command{
		{ID: id1, Name: "cmd1", Command: "echo 1", WorkDir: "/tmp"},
		{ID: id2, Name: "cmd2", Command: "echo 2", WorkDir: "/var"},
	}

	err := repo.Save(commands)
	require.NoError(t, err)

	loaded, err := repo.Load()
	require.NoError(t, err)

	assert.Equal(t, commands, loaded)
}

func TestJSONFileRepository_FileCreatedOnFirstSave(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "subdir", "commands.json")
	repo := NewJSONFileRepository(path)

	commands := []Command{
		{ID: uuid.MustParse("01234567-89ab-cdef-0123-456789abcdef"), Name: "cmd1", Command: "echo 1", WorkDir: "/tmp"},
	}

	err := repo.Save(commands)
	require.NoError(t, err)

	_, err = os.Stat(path)
	assert.NoError(t, err)

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(content), "01234567-89ab-cdef-0123-456789abcdef")
}

func TestJSONFileRepository_LoadFromExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "commands.json")
	content := `{"commands":[{"id":"01234567-89ab-cdef-0123-456789abcdef","name":"test","command":"echo test","work_dir":"/tmp"}]}`
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)

	repo := NewJSONFileRepository(path)
	loaded, err := repo.Load()

	require.NoError(t, err)
	require.Len(t, loaded, 1)
	assert.Equal(t, uuid.MustParse("01234567-89ab-cdef-0123-456789abcdef"), loaded[0].ID)
	assert.Equal(t, "test", loaded[0].Name)
	assert.Equal(t, "echo test", loaded[0].Command)
	assert.Equal(t, "/tmp", loaded[0].WorkDir)
}

func TestJSONFileRepository_LoadFromNonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "nonexistent.json")
	repo := NewJSONFileRepository(path)

	loaded, err := repo.Load()

	require.NoError(t, err)
	assert.Empty(t, loaded)
}

func TestJSONFileRepository_LoadFromEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "empty.json")
	err := os.WriteFile(path, []byte(""), 0644)
	require.NoError(t, err)

	repo := NewJSONFileRepository(path)
	loaded, err := repo.Load()

	require.NoError(t, err)
	assert.Empty(t, loaded)
}

func TestJSONFileRepository_SaveEmptyList(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "commands.json")
	repo := NewJSONFileRepository(path)

	err := repo.Save([]Command{})
	require.NoError(t, err)

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, `{"commands":[]}`, string(content))
}

func TestJSONFileRepository_LoadWithMalformedJSON(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "malformed.json")
	err := os.WriteFile(path, []byte("{invalid json}"), 0644)
	require.NoError(t, err)

	repo := NewJSONFileRepository(path)
	_, err = repo.Load()

	assert.Error(t, err)
}
