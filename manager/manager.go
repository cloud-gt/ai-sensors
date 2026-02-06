package manager

import (
	"context"
	"errors"
	"sync"

	"github.com/cloud-gt/ai-sensors/buffer"
	"github.com/cloud-gt/ai-sensors/command"
	"github.com/cloud-gt/ai-sensors/runner"
	"github.com/google/uuid"
)

var (
	ErrCommandNotFound = errors.New("command not found in store")
	ErrNotRunning      = errors.New("command is not running")
)

type Status string

const (
	StatusNotStarted Status = "not_started"
	StatusRunning    Status = "running"
	StatusStopped    Status = "stopped"
)

const defaultBufferCapacity = 1000

type Manager struct {
	store     *command.Store
	bufferCap int
	mu        sync.RWMutex
	instances map[uuid.UUID]*Instance
}

type Instance struct {
	command command.Command
	runner  *runner.Runner
	buffer  *buffer.RingBuffer
	status  Status
	cancel  context.CancelFunc
}

type Option func(*Manager)

func WithBufferCapacity(cap int) Option {
	return func(m *Manager) {
		if cap > 0 {
			m.bufferCap = cap
		}
	}
}

func New(store *command.Store, opts ...Option) *Manager {
	m := &Manager{
		store:     store,
		bufferCap: defaultBufferCapacity,
		instances: make(map[uuid.UUID]*Instance),
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func (m *Manager) Start(ctx context.Context, id uuid.UUID) (bool, error) {
	cmd, err := m.store.Get(id)
	if err != nil {
		if errors.Is(err, command.ErrNotFound) {
			return false, ErrCommandNotFound
		}
		return false, err
	}

	m.mu.Lock()
	if inst, exists := m.instances[id]; exists {
		if inst.status == StatusRunning {
			m.mu.Unlock()
			return false, nil
		}
	}

	buf, err := buffer.New(m.bufferCap)
	if err != nil {
		m.mu.Unlock()
		return false, err
	}

	r, err := runner.New(runner.Config{
		Command: "sh",
		Args:    []string{"-c", cmd.Command},
		Output:  buf,
		Dir:     cmd.WorkDir,
	})
	if err != nil {
		m.mu.Unlock()
		return false, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	inst := &Instance{
		command: cmd,
		runner:  r,
		buffer:  buf,
		status:  StatusRunning,
		cancel:  cancel,
	}
	m.instances[id] = inst
	m.mu.Unlock()

	go func() {
		_ = r.Start(ctx)

		m.mu.Lock()
		if m.instances[id] != nil {
			m.instances[id].status = StatusStopped
		}
		m.mu.Unlock()
	}()

	return true, nil
}

func (m *Manager) Stop(id uuid.UUID) error {
	m.mu.RLock()
	inst, exists := m.instances[id]
	m.mu.RUnlock()

	if !exists {
		return ErrNotRunning
	}

	if inst.status == StatusStopped {
		return nil
	}

	inst.cancel()
	_ = inst.runner.Stop()

	return nil
}

func (m *Manager) Output(id uuid.UUID) ([]string, error) {
	m.mu.RLock()
	inst, exists := m.instances[id]
	m.mu.RUnlock()

	if !exists {
		return nil, ErrNotRunning
	}

	return inst.buffer.Lines(), nil
}

func (m *Manager) OutputLastN(id uuid.UUID, n int) ([]string, error) {
	m.mu.RLock()
	inst, exists := m.instances[id]
	m.mu.RUnlock()

	if !exists {
		return nil, ErrNotRunning
	}

	return inst.buffer.LastN(n), nil
}

func (m *Manager) Status(id uuid.UUID) (Status, error) {
	m.mu.RLock()
	inst, exists := m.instances[id]
	m.mu.RUnlock()

	if !exists {
		return StatusNotStarted, ErrNotRunning
	}

	return inst.status, nil
}

func (m *Manager) Shutdown(ctx context.Context) error {
	m.mu.RLock()
	ids := make([]uuid.UUID, 0, len(m.instances))
	for id := range m.instances {
		ids = append(ids, id)
	}
	m.mu.RUnlock()

	done := make(chan struct{})
	go func() {
		for _, id := range ids {
			_ = m.Stop(id)
		}
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
