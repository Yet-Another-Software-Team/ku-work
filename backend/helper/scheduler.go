package helper

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// ScheduledTask represents a single scheduled task
type ScheduledTask struct {
	Name     string
	Interval time.Duration
	Fn       func() error
}

// Scheduler manages multiple periodic tasks
type Scheduler struct {
	ctx    context.Context
	tasks  []ScheduledTask
	wg     sync.WaitGroup
	mutex  sync.Mutex
	stopCh chan struct{}
}

// NewScheduler creates a new scheduler instance
func NewScheduler(ctx context.Context) *Scheduler {
	return &Scheduler{
		ctx:    ctx,
		tasks:  make([]ScheduledTask, 0),
		stopCh: make(chan struct{}),
	}
}

// AddTask adds a new scheduled task to the scheduler
// The task will be executed periodically at the specified interval
func (s *Scheduler) AddTask(name string, interval time.Duration, fn func() error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.tasks = append(s.tasks, ScheduledTask{
		Name:     name,
		Interval: interval,
		Fn:       fn,
	})
	slog.Info("Scheduled task added", "name", name, "interval", interval)
}

// Start begins executing all scheduled tasks
// Each task runs in its own goroutine and executes immediately, then at intervals
func (s *Scheduler) Start() {
	s.mutex.Lock()
	tasksCopy := make([]ScheduledTask, len(s.tasks))
	copy(tasksCopy, s.tasks)
	s.mutex.Unlock()

	slog.Info("Starting scheduler", "tasks", len(tasksCopy))

	for _, task := range tasksCopy {
		s.wg.Add(1)
		go s.runTask(task)
	}
}

// runTask executes a single task periodically
func (s *Scheduler) runTask(task ScheduledTask) {
	defer s.wg.Done()

	ticker := time.NewTicker(task.Interval)
	defer ticker.Stop()

	// Run immediately on start
	s.executeTask(task)

	for {
		select {
		case <-ticker.C:
			s.executeTask(task)
		case <-s.ctx.Done():
			slog.Info("Task stopped", "name", task.Name)
			return
		case <-s.stopCh:
			slog.Info("Task stopped", "name", task.Name)
			return
		}
	}
}

// executeTask runs the task function and handles errors
func (s *Scheduler) executeTask(task ScheduledTask) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Task panicked", "name", task.Name, "reason", r)
		}
	}()

	if err := task.Fn(); err != nil {
		slog.Error("Task failed", "name", task.Name, "error", err)
	}
}

// Stop gracefully stops all scheduled tasks
func (s *Scheduler) Stop() {
	slog.Info("Stopping scheduler...")
	close(s.stopCh)
	s.wg.Wait()
	slog.Info("Scheduler stopped")
}

// Wait blocks until all tasks have stopped
func (s *Scheduler) Wait() {
	s.wg.Wait()
}
