package command

import (
	"sync"

	"github.com/google/uuid"
)

type Store struct {
	repo     Repository
	mu       sync.RWMutex
	commands []Command
}

func NewStore(repo Repository) *Store {
	return &Store{
		repo:     repo,
		commands: []Command{},
	}
}

func (s *Store) Create(cmd Command) (Command, error) {
	if cmd.Name == "" {
		return Command{}, ErrEmptyName
	}
	if cmd.Command == "" {
		return Command{}, ErrEmptyCommand
	}
	if cmd.WorkDir == "" {
		return Command{}, ErrEmptyWorkDir
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	cmd.ID = uuid.Must(uuid.NewV7())
	s.commands = append(s.commands, cmd)

	if err := s.repo.Save(s.commands); err != nil {
		s.commands = s.commands[:len(s.commands)-1]
		return Command{}, err
	}

	return cmd, nil
}

func (s *Store) Get(id uuid.UUID) (Command, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, cmd := range s.commands {
		if cmd.ID == id {
			return cmd, nil
		}
	}

	return Command{}, ErrNotFound
}

func (s *Store) Update(cmd Command) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.commands {
		if existing.ID == cmd.ID {
			s.commands[i] = cmd
			if err := s.repo.Save(s.commands); err != nil {
				s.commands[i] = existing
				return err
			}
			return nil
		}
	}

	return ErrNotFound
}

func (s *Store) Delete(id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, cmd := range s.commands {
		if cmd.ID == id {
			deleted := s.commands[i]
			s.commands = append(s.commands[:i], s.commands[i+1:]...)
			if err := s.repo.Save(s.commands); err != nil {
				s.commands = append(s.commands[:i], append([]Command{deleted}, s.commands[i:]...)...)
				return err
			}
			return nil
		}
	}

	return ErrNotFound
}

func (s *Store) List() ([]Command, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Command, len(s.commands))
	copy(result, s.commands)

	return result, nil
}

func (s *Store) GetByName(name string) (Command, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, cmd := range s.commands {
		if cmd.Name == name {
			return cmd, nil
		}
	}

	return Command{}, ErrNotFound
}
