package command

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrNotFound     = errors.New("command not found")
	ErrEmptyName    = errors.New("command name cannot be empty")
	ErrEmptyCommand = errors.New("command string cannot be empty")
)

type Command struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Command string    `json:"command"`
}
