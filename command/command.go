package command

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrNotFound     = errors.New("command not found")
	ErrEmptyName    = errors.New("command name cannot be empty")
	ErrEmptyCommand = errors.New("command string cannot be empty")
	ErrEmptyWorkDir = errors.New("command work_dir cannot be empty")
)

type Command struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Command string    `json:"command"`
	WorkDir string    `json:"work_dir"`
}
