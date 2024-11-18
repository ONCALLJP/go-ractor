package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ONCALLJP/goractor/internal/executor"
	"github.com/ONCALLJP/goractor/internal/task"
)

type Scheduler struct {
	tasks    *task.Manager
	executor *executor.Executor
	runners  map[string]*TaskRunner
	mu       sync.RWMutex
}

type TaskRunner struct {
	task     *task.Task
	cancel   context.CancelFunc
	interval time.Duration
}

func NewScheduler(tasks *task.Manager, executor *executor.Executor) *Scheduler {
	return &Scheduler{
		tasks:    tasks,
		executor: executor,
		runners:  make(map[string]*TaskRunner),
	}
}

func (s *Scheduler) GetExecutor() *executor.Executor {
	return s.executor
}

func (s *Scheduler) Start() error {
	tasks := s.tasks.List()
	for _, t := range tasks {
		if err := s.StartTask(&t); err != nil {
			return fmt.Errorf("failed to start task %s: %w", t.Name, err)
		}
	}
	return nil
}

func (s *Scheduler) StartTask(t *task.Task) error {
	interval, err := parseSchedule(t.Schedule)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Stop existing runner if any
	if runner, exists := s.runners[t.Name]; exists {
		runner.cancel()
	}

	ctx, cancel := context.WithCancel(context.Background())
	runner := &TaskRunner{
		task:     t,
		cancel:   cancel,
		interval: interval,
	}
	s.runners[t.Name] = runner

	go s.runTask(ctx, runner)
	return nil
}

func (s *Scheduler) StopTask(taskName string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if runner, exists := s.runners[taskName]; exists {
		runner.cancel()
		delete(s.runners, taskName)
	}
}

func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, runner := range s.runners {
		runner.cancel()
	}
	s.runners = make(map[string]*TaskRunner)
}

func (s *Scheduler) runTask(ctx context.Context, runner *TaskRunner) {
	ticker := time.NewTicker(runner.interval)
	defer ticker.Stop()

	// Run immediately on start
	if err := s.executor.Execute(ctx, runner.task); err != nil {
		fmt.Printf("Error executing task %s: %v\n", runner.task.Name, err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.executor.Execute(ctx, runner.task); err != nil {
				fmt.Printf("Error executing task %s: %v\n", runner.task.Name, err)
			}
		}
	}
}

func parseSchedule(schedule string) (time.Duration, error) {
	// Handle "every X" format
	var duration time.Duration
	var err error

	switch schedule {
	case "every 1h":
		duration = time.Hour
	case "every 30m":
		duration = 30 * time.Minute
	case "every 15m":
		duration = 15 * time.Minute
	case "every 5m":
		duration = 5 * time.Minute
	default:
		duration, err = time.ParseDuration(schedule)
		if err != nil {
			return 0, fmt.Errorf("invalid schedule format: %s", schedule)
		}
	}

	return duration, nil
}
