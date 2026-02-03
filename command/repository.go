package command

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Repository interface {
	Load() ([]Command, error)
	Save(commands []Command) error
}

type MemoryRepository struct {
	commands []Command
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		commands: []Command{},
	}
}

func (r *MemoryRepository) Load() ([]Command, error) {
	result := make([]Command, len(r.commands))
	copy(result, r.commands)
	return result, nil
}

func (r *MemoryRepository) Save(commands []Command) error {
	r.commands = make([]Command, len(commands))
	copy(r.commands, commands)
	return nil
}

type JSONFileRepository struct {
	filePath string
}

type jsonFileData struct {
	Commands []Command `json:"commands"`
}

func NewJSONFileRepository(path string) *JSONFileRepository {
	return &JSONFileRepository{
		filePath: path,
	}
}

func (r *JSONFileRepository) Load() ([]Command, error) {
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []Command{}, nil
		}
		return nil, err
	}

	if len(data) == 0 {
		return []Command{}, nil
	}

	var fileData jsonFileData
	if err := json.Unmarshal(data, &fileData); err != nil {
		return nil, err
	}

	return fileData.Commands, nil
}

func (r *JSONFileRepository) Save(commands []Command) error {
	dir := filepath.Dir(r.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	fileData := jsonFileData{Commands: commands}
	data, err := json.Marshal(fileData)
	if err != nil {
		return err
	}

	return os.WriteFile(r.filePath, data, 0644)
}
