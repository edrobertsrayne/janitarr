package services

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestScheduler_StartStop(t *testing.T) {
	var mu sync.Mutex
	callbackCount := 0
	cb := func(ctx context.Context, isManual bool) error {
		mu.Lock()
		callbackCount++
		mu.Unlock()
		return nil
	}

	scheduler := NewScheduler(1, cb)
	ctx := context.Background()

	err := scheduler.Start(ctx)
	if err != nil {
		t.Fatalf("scheduler.Start() error = %v", err)
	}

	if !scheduler.IsRunning() {
		t.Error("scheduler should be running after Start()")
	}

	err = scheduler.Start(ctx)
	if err == nil {
		t.Error("scheduler.Start() should return an error when already running")
	}

	scheduler.Stop()
	if scheduler.IsRunning() {
		t.Error("scheduler should not be running after Stop()")
	}

	// check that callback was not called
	mu.Lock()
	if callbackCount > 0 {
		t.Errorf("callback should not be called, but was called %d times", callbackCount)
	}
	mu.Unlock()
}

func TestScheduler_IntervalConfig(t *testing.T) {
	cb := func(ctx context.Context, isManual bool) error {
		return nil
	}

	scheduler := NewScheduler(2, cb)
	ctx := context.Background()
	_ = scheduler.Start(ctx)
	defer scheduler.Stop()

	status := scheduler.GetStatus()
	if status.IntervalHours != 2 {
		t.Errorf("expected interval of 2, got %d", status.IntervalHours)
	}

	if status.NextRun.IsZero() {
		t.Error("NextRun should be set")
	}

	durationUntilNext := scheduler.GetTimeUntilNextRun()
	if durationUntilNext > 2*time.Hour || durationUntilNext < 1*time.Hour+59*time.Minute {
		t.Errorf("invalid time until next run: %v", durationUntilNext)
	}
}

func TestScheduler_PreventsConcurrent(t *testing.T) {
	startChan := make(chan struct{})
	finishChan := make(chan struct{})

	cb := func(ctx context.Context, isManual bool) error {
		startChan <- struct{}{}
		<-finishChan
		return nil
	}

	scheduler := NewScheduler(1, cb)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = scheduler.TriggerManual(ctx)
	}()

	<-startChan
	err := scheduler.TriggerManual(ctx)
	if err == nil {
		t.Error("expected an error when triggering a cycle while one is already active")
	}
	close(finishChan)
}

func TestScheduler_ManualTrigger(t *testing.T) {
	var mu sync.Mutex
	manualTrigger := false
	cb := func(ctx context.Context, isManual bool) error {
		mu.Lock()
		manualTrigger = isManual
		mu.Unlock()
		return nil
	}

	scheduler := NewScheduler(1, cb)
	ctx := context.Background()
	_ = scheduler.Start(ctx)
	defer scheduler.Stop()

	err := scheduler.TriggerManual(ctx)
	if err != nil {
		t.Fatalf("TriggerManual failed: %v", err)
	}

	mu.Lock()
	if !manualTrigger {
		t.Error("expected manual trigger to be true")
	}
	mu.Unlock()
}

func TestScheduler_CallbackError(t *testing.T) {
	expectedErr := errors.New("callback error")
	cb := func(ctx context.Context, isManual bool) error {
		return expectedErr
	}

	scheduler := NewScheduler(1, cb)
	ctx := context.Background()
	_ = scheduler.Start(ctx)
	defer scheduler.Stop()

	err := scheduler.TriggerManual(ctx)
	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestScheduler_StatusUpdates(t *testing.T) {
	cb := func(ctx context.Context, isManual bool) error {
		return nil
	}

	scheduler := NewScheduler(1, cb)
	ctx := context.Background()

	status := scheduler.GetStatus()
	if status.IsRunning {
		t.Error("scheduler should not be running")
	}

	_ = scheduler.Start(ctx)
	defer scheduler.Stop()

	status = scheduler.GetStatus()
	if !status.IsRunning {
		t.Error("scheduler should be running")
	}
	if status.NextRun.IsZero() {
		t.Error("NextRun should be set")
	}
}

func TestScheduler_GracefulShutdown(t *testing.T) {
	cbExecuted := make(chan struct{})
	cb := func(ctx context.Context, isManual bool) error {
		time.Sleep(100 * time.Millisecond)
		close(cbExecuted)
		return nil
	}

	scheduler := NewScheduler(1, cb)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := scheduler.Start(ctx)
	if err != nil {
		t.Fatalf("scheduler.Start() error = %v", err)
	}

	// Manually trigger and immediately stop
	go func() {
		_ = scheduler.TriggerManual(ctx)
	}()

	// Give a moment for TriggerManual to start
	time.Sleep(10 * time.Millisecond)
	scheduler.Stop()

	select {
	case <-cbExecuted:
		// success
	case <-time.After(200 * time.Millisecond):
		t.Error("callback did not execute during shutdown")
	}
}
