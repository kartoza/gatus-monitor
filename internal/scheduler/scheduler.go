// Copyright (c) 2026 Kartoza
// SPDX-License-Identifier: MIT

// Package scheduler provides staggered task scheduling
package scheduler

import (
	"context"
	"sync"
	"time"
)

// TaskFunc is a function that performs a scheduled task
type TaskFunc func(ctx context.Context, id string) error

// Task represents a scheduled task
type Task struct {
	ID          string
	Offset      time.Duration
	TaskFunc    TaskFunc
	ticker      *time.Ticker
	stopChan    chan struct{}
	stoppedChan chan struct{}
	running     bool
	mu          sync.Mutex
}

// Scheduler manages multiple tasks with staggered execution
type Scheduler struct {
	interval time.Duration
	tasks    map[string]*Task
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

// New creates a new scheduler with the specified interval
func New(interval time.Duration) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())

	return &Scheduler{
		interval: interval,
		tasks:    make(map[string]*Task),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// AddTask adds a task to the scheduler with staggered offset
func (s *Scheduler) AddTask(id string, taskFunc TaskFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Calculate offset based on number of existing tasks
	offset := s.calculateOffset(len(s.tasks))

	task := &Task{
		ID:          id,
		Offset:      offset,
		TaskFunc:    taskFunc,
		stopChan:    make(chan struct{}),
		stoppedChan: make(chan struct{}),
	}

	s.tasks[id] = task
}

// RemoveTask removes a task from the scheduler
func (s *Scheduler) RemoveTask(id string) {
	s.mu.Lock()
	task, exists := s.tasks[id]
	if exists {
		delete(s.tasks, id)
	}
	s.mu.Unlock()

	if exists && task != nil {
		task.Stop()
	}

	// Recalculate offsets for remaining tasks
	s.recalculateOffsets()
}

// Start starts all scheduled tasks
func (s *Scheduler) Start() {
	s.mu.RLock()
	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	s.mu.RUnlock()

	for _, task := range tasks {
		go task.Start(s.ctx, s.interval)
	}
}

// Stop stops all scheduled tasks
func (s *Scheduler) Stop() {
	s.cancel()

	s.mu.RLock()
	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	s.mu.RUnlock()

	for _, task := range tasks {
		task.Stop()
	}
}

// UpdateInterval updates the interval and restarts all tasks
func (s *Scheduler) UpdateInterval(interval time.Duration) {
	s.mu.Lock()
	s.interval = interval
	s.mu.Unlock()

	// Restart all tasks with new interval
	s.Stop()

	// Create new context
	s.ctx, s.cancel = context.WithCancel(context.Background())

	// Recalculate offsets
	s.recalculateOffsets()

	// Restart
	s.Start()
}

// calculateOffset calculates the stagger offset for a task
func (s *Scheduler) calculateOffset(taskIndex int) time.Duration {
	if len(s.tasks) == 0 {
		return 0
	}

	// Distribute tasks evenly across the interval
	return (s.interval / time.Duration(len(s.tasks)+1)) * time.Duration(taskIndex)
}

// recalculateOffsets recalculates offsets for all tasks
func (s *Scheduler) recalculateOffsets() {
	s.mu.Lock()
	defer s.mu.Unlock()

	i := 0
	for _, task := range s.tasks {
		task.mu.Lock()
		task.Offset = s.calculateOffset(i)
		task.mu.Unlock()
		i++
	}
}

// GetTaskCount returns the number of tasks
func (s *Scheduler) GetTaskCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.tasks)
}

// Start starts a task
func (t *Task) Start(ctx context.Context, interval time.Duration) {
	t.mu.Lock()
	if t.running {
		t.mu.Unlock()
		return
	}
	t.running = true
	offset := t.Offset
	// Recreate channels for fresh start
	t.stopChan = make(chan struct{})
	t.stoppedChan = make(chan struct{})
	stoppedChan := t.stoppedChan
	t.mu.Unlock()

	defer func() {
		t.mu.Lock()
		t.running = false
		t.mu.Unlock()
		close(stoppedChan)
	}()

	// Wait for offset before first execution
	if offset > 0 {
		select {
		case <-time.After(offset):
		case <-t.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}

	// Execute immediately after offset
	_ = t.TaskFunc(ctx, t.ID)

	// Create ticker for subsequent executions
	t.ticker = time.NewTicker(interval)
	defer t.ticker.Stop()

	for {
		select {
		case <-t.ticker.C:
			_ = t.TaskFunc(ctx, t.ID)
		case <-t.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

// Stop stops a task
func (t *Task) Stop() {
	t.mu.Lock()
	if !t.running {
		t.mu.Unlock()
		return
	}
	stopChan := t.stopChan
	stoppedChan := t.stoppedChan
	t.mu.Unlock()

	// Close stopChan - may already be closed if context cancelled first
	select {
	case <-stopChan:
		// Already closed
	default:
		close(stopChan)
	}

	// Wait for task to stop
	<-stoppedChan
}
