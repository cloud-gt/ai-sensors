package runner

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

type State string

const (
	StateInitial State = "initial"
	StateRunning State = "running"
	StateStopped State = "stopped"
)

const defaultStopTimeout = 5 * time.Second

var (
	ErrEmptyCommand   = errors.New("command cannot be empty")
	ErrNilOutput      = errors.New("output writer cannot be nil")
	ErrNilContext     = errors.New("context cannot be nil")
	ErrAlreadyStarted = errors.New("runner has already been started")
)

type Config struct {
	Command     string
	Args        []string
	Output      io.Writer
	StopTimeout time.Duration
}

type Runner struct {
	config Config

	mu        sync.RWMutex
	state     State
	cmd       *exec.Cmd
	stopOnce  sync.Once
	waitDone  chan struct{}
	waitErr   error
	cancelCtx context.CancelFunc
	stopped   bool
}

func New(cfg Config) (*Runner, error) {
	if cfg.Command == "" {
		return nil, ErrEmptyCommand
	}
	if cfg.Output == nil {
		return nil, ErrNilOutput
	}
	if cfg.StopTimeout == 0 {
		cfg.StopTimeout = defaultStopTimeout
	}

	return &Runner{
		config:   cfg,
		state:    StateInitial,
		waitDone: make(chan struct{}),
	}, nil
}

func (r *Runner) Start(ctx context.Context) error {
	if ctx == nil {
		return ErrNilContext
	}

	r.mu.Lock()
	if r.state != StateInitial {
		r.mu.Unlock()
		return ErrAlreadyStarted
	}

	internalCtx, cancel := context.WithCancel(context.Background())
	r.cancelCtx = cancel

	r.cmd = exec.Command(r.config.Command, r.config.Args...)
	r.cmd.Stdout = r.config.Output
	r.cmd.Stderr = r.config.Output
	r.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := r.cmd.Start(); err != nil {
		r.mu.Unlock()
		if errors.Is(err, exec.ErrNotFound) {
			return exec.ErrNotFound
		}
		return err
	}

	r.state = StateRunning
	r.mu.Unlock()

	processDone := make(chan error, 1)
	go func() {
		processDone <- r.cmd.Wait()
	}()

	go func() {
		select {
		case <-ctx.Done():
			r.doStop()
		case <-internalCtx.Done():
		}
	}()

	err := <-processDone

	r.mu.Lock()
	r.state = StateStopped
	r.waitErr = err
	wasStopped := r.stopped
	close(r.waitDone)
	r.mu.Unlock()

	if wasStopped || ctx.Err() == context.Canceled {
		return context.Canceled
	}

	return err
}

func (r *Runner) Stop() error {
	r.mu.RLock()
	state := r.state
	r.mu.RUnlock()

	if state == StateInitial || state == StateStopped {
		return nil
	}

	r.doStop()
	<-r.waitDone
	return nil
}

func (r *Runner) doStop() {
	r.stopOnce.Do(func() {
		r.mu.Lock()
		r.stopped = true
		cmd := r.cmd
		r.mu.Unlock()

		if cmd == nil || cmd.Process == nil {
			return
		}

		pgid := cmd.Process.Pid
		if err := syscall.Kill(-pgid, syscall.SIGTERM); err != nil {
			slog.Warn("failed to send SIGTERM to process group", "pgid", pgid, "error", err)
		}

		select {
		case <-r.waitDone:
			return
		case <-time.After(r.config.StopTimeout):
			if err := syscall.Kill(-pgid, syscall.SIGKILL); err != nil {
				slog.Warn("failed to send SIGKILL to process group", "pgid", pgid, "error", err)
			}
		}
	})
}

func (r *Runner) State() State {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state
}

func (r *Runner) Wait() error {
	<-r.waitDone
	return r.waitErr
}
