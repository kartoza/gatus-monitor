package scheduler

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	scheduler := New(60 * time.Second)
	require.NotNil(t, scheduler)
	assert.Equal(t, 60*time.Second, scheduler.interval)
	assert.NotNil(t, scheduler.tasks)
	assert.NotNil(t, scheduler.ctx)
}

func TestAddTask(t *testing.T) {
	scheduler := New(60 * time.Second)

	taskCalled := false
	taskFunc := func(ctx context.Context, id string) error {
		taskCalled = true
		return nil
	}

	scheduler.AddTask("task1", taskFunc)

	assert.Equal(t, 1, scheduler.GetTaskCount())
	task := scheduler.tasks["task1"]
	assert.NotNil(t, task)
	assert.Equal(t, "task1", task.ID)
	assert.Equal(t, time.Duration(0), task.Offset) // First task has 0 offset
}

func TestAddTask_MultipleWithStagger(t *testing.T) {
	scheduler := New(60 * time.Second)

	taskFunc := func(ctx context.Context, id string) error {
		return nil
	}

	scheduler.AddTask("task1", taskFunc)
	scheduler.AddTask("task2", taskFunc)
	scheduler.AddTask("task3", taskFunc)

	assert.Equal(t, 3, scheduler.GetTaskCount())

	// Verify tasks have different offsets
	task1 := scheduler.tasks["task1"]
	task2 := scheduler.tasks["task2"]
	task3 := scheduler.tasks["task3"]

	assert.Equal(t, time.Duration(0), task1.Offset)
	assert.Greater(t, task2.Offset, time.Duration(0))
	assert.Greater(t, task3.Offset, task2.Offset)
}

func TestRemoveTask(t *testing.T) {
	scheduler := New(60 * time.Second)

	taskFunc := func(ctx context.Context, id string) error {
		return nil
	}

	scheduler.AddTask("task1", taskFunc)
	scheduler.AddTask("task2", taskFunc)

	assert.Equal(t, 2, scheduler.GetTaskCount())

	scheduler.RemoveTask("task1")

	assert.Equal(t, 1, scheduler.GetTaskCount())
	assert.Nil(t, scheduler.tasks["task1"])
	assert.NotNil(t, scheduler.tasks["task2"])
}

func TestScheduler_StartStop(t *testing.T) {
	scheduler := New(100 * time.Millisecond)

	var mu sync.Mutex
	executions := 0

	taskFunc := func(ctx context.Context, id string) error {
		mu.Lock()
		executions++
		mu.Unlock()
		return nil
	}

	scheduler.AddTask("task1", taskFunc)
	scheduler.Start()

	// Wait for a few executions
	time.Sleep(350 * time.Millisecond)

	scheduler.Stop()

	mu.Lock()
	finalCount := executions
	mu.Unlock()

	// Should have executed at least 2-3 times (initial + 2-3 intervals)
	assert.GreaterOrEqual(t, finalCount, 2)

	// Wait a bit more and verify no more executions
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	afterStopCount := executions
	mu.Unlock()

	assert.Equal(t, finalCount, afterStopCount, "Task should not execute after stop")
}

func TestScheduler_Staggering(t *testing.T) {
	scheduler := New(1 * time.Second)

	var mu sync.Mutex
	executionTimes := make(map[string][]time.Time)

	taskFunc := func(ctx context.Context, id string) error {
		mu.Lock()
		executionTimes[id] = append(executionTimes[id], time.Now())
		mu.Unlock()
		return nil
	}

	scheduler.AddTask("task1", taskFunc)
	scheduler.AddTask("task2", taskFunc)

	scheduler.Start()

	// Wait for first execution of both tasks
	time.Sleep(600 * time.Millisecond)

	scheduler.Stop()

	mu.Lock()
	times1 := executionTimes["task1"]
	times2 := executionTimes["task2"]
	mu.Unlock()

	// Both tasks should have executed
	assert.NotEmpty(t, times1)
	assert.NotEmpty(t, times2)

	// Tasks should have executed at different times (staggered)
	if len(times1) > 0 && len(times2) > 0 {
		diff := times2[0].Sub(times1[0]).Abs()
		assert.Greater(t, diff, 100*time.Millisecond, "Tasks should be staggered")
	}
}

func TestUpdateInterval(t *testing.T) {
	scheduler := New(100 * time.Millisecond)

	var mu sync.Mutex
	executions := 0

	taskFunc := func(ctx context.Context, id string) error {
		mu.Lock()
		executions++
		mu.Unlock()
		return nil
	}

	scheduler.AddTask("task1", taskFunc)
	scheduler.Start()

	// Wait for initial executions
	time.Sleep(350 * time.Millisecond)

	// Update interval to slower rate
	scheduler.UpdateInterval(1 * time.Second)

	mu.Lock()
	countBefore := executions
	mu.Unlock()

	// Wait for new interval period
	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	countAfter := executions
	mu.Unlock()

	// With new 1-second interval, should have only 1-2 more executions max
	assert.LessOrEqual(t, countAfter-countBefore, 2)

	scheduler.Stop()
}

func TestTask_ContextCancellation(t *testing.T) {
	scheduler := New(100 * time.Millisecond)

	var mu sync.Mutex
	executions := 0

	taskFunc := func(ctx context.Context, id string) error {
		mu.Lock()
		executions++
		mu.Unlock()
		return nil
	}

	scheduler.AddTask("task1", taskFunc)
	scheduler.Start()

	time.Sleep(250 * time.Millisecond)

	// Cancel context
	scheduler.cancel()

	mu.Lock()
	countBefore := executions
	mu.Unlock()

	// Wait and verify no more executions
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	countAfter := executions
	mu.Unlock()

	assert.Equal(t, countBefore, countAfter, "Task should stop on context cancellation")
}

func TestCalculateOffset(t *testing.T) {
	scheduler := New(60 * time.Second)

	// No tasks yet
	offset := scheduler.calculateOffset(0)
	assert.Equal(t, time.Duration(0), offset)

	// Add a task
	scheduler.tasks["task1"] = &Task{}
	offset = scheduler.calculateOffset(1)
	assert.Greater(t, offset, time.Duration(0))
	assert.Less(t, offset, 60*time.Second)
}

func TestRecalculateOffsets(t *testing.T) {
	scheduler := New(60 * time.Second)

	taskFunc := func(ctx context.Context, id string) error {
		return nil
	}

	scheduler.AddTask("task1", taskFunc)
	scheduler.AddTask("task2", taskFunc)
	scheduler.AddTask("task3", taskFunc)

	originalOffset2 := scheduler.tasks["task2"].Offset
	originalOffset3 := scheduler.tasks["task3"].Offset

	// Remove task1, which should cause recalculation
	scheduler.RemoveTask("task1")

	// Offsets should be different after recalculation
	newOffset2 := scheduler.tasks["task2"].Offset
	newOffset3 := scheduler.tasks["task3"].Offset

	// After removal and recalculation, task2 might become task with index 0
	// So its offset could be different
	assert.NotEqual(t, originalOffset2, newOffset2)
	assert.NotEqual(t, originalOffset3, newOffset3)
}
