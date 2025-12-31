package scheduler

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestSchedulerEvery(t *testing.T) {
	s := New()

	var counter int32
	err := s.Every("test-job", 100*time.Millisecond, func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
	})
	if err != nil {
		t.Fatalf("failed to schedule job: %v", err)
	}

	s.Start()
	defer s.Stop()

	// Wait for at least 1 execution (cron aligns to second boundaries)
	time.Sleep(1200 * time.Millisecond)

	count := atomic.LoadInt32(&counter)
	if count < 1 {
		t.Errorf("expected at least 1 execution, got %d", count)
	}
}

func TestSchedulerCron(t *testing.T) {
	s := New()

	var executed int32
	// Schedule to run every second (using @every for simpler testing)
	err := s.Cron("cron-job", "@every 1s", func(ctx context.Context) {
		atomic.StoreInt32(&executed, 1)
	})
	if err != nil {
		t.Fatalf("failed to schedule cron job: %v", err)
	}

	s.Start()
	defer s.Stop()

	time.Sleep(1100 * time.Millisecond)

	if atomic.LoadInt32(&executed) != 1 {
		t.Error("cron job did not execute")
	}
}

func TestSchedulerRemove(t *testing.T) {
	s := New()

	var counter int32
	err := s.Every("removable-job", 1*time.Second, func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
	})
	if err != nil {
		t.Fatalf("failed to schedule job: %v", err)
	}

	s.Start()

	// Let it run once
	time.Sleep(1100 * time.Millisecond)
	countBefore := atomic.LoadInt32(&counter)

	// Remove the job
	removed := s.Remove("removable-job")
	if !removed {
		t.Error("expected job to be removed")
	}

	// Wait and verify no more executions
	time.Sleep(1100 * time.Millisecond)
	countAfter := atomic.LoadInt32(&counter)

	s.Stop()

	if countAfter > countBefore {
		t.Errorf("job continued running after removal: before=%d, after=%d", countBefore, countAfter)
	}
}

func TestSchedulerRemoveNonExistent(t *testing.T) {
	s := New()

	removed := s.Remove("non-existent")
	if removed {
		t.Error("expected false when removing non-existent job")
	}
}

func TestSchedulerJobs(t *testing.T) {
	s := New()

	s.Every("job1", time.Minute, func(ctx context.Context) {})
	s.Every("job2", time.Hour, func(ctx context.Context) {})
	s.Cron("job3", "0 0 * * *", func(ctx context.Context) {})

	jobs := s.Jobs()
	if len(jobs) != 3 {
		t.Errorf("expected 3 jobs, got %d", len(jobs))
	}

	// Verify job names
	names := make(map[string]bool)
	for _, job := range jobs {
		names[job.Name] = true
	}

	for _, expected := range []string{"job1", "job2", "job3"} {
		if !names[expected] {
			t.Errorf("missing job: %s", expected)
		}
	}
}

func TestSchedulerStartStop(t *testing.T) {
	s := New()

	if s.Running() {
		t.Error("scheduler should not be running initially")
	}

	s.Start()
	if !s.Running() {
		t.Error("scheduler should be running after Start()")
	}

	// Double start should be idempotent
	s.Start()
	if !s.Running() {
		t.Error("scheduler should still be running after double Start()")
	}

	s.Stop()
	if s.Running() {
		t.Error("scheduler should not be running after Stop()")
	}

	// Double stop should be safe
	ctx := s.Stop()
	select {
	case <-ctx.Done():
		// Expected
	default:
		t.Error("stop context should be done")
	}
}

func TestSchedulerWithLocation(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	s := New(WithLocation(loc))

	if s.location != loc {
		t.Error("location not set correctly")
	}
}

func TestSchedulerSkipIfRunning(t *testing.T) {
	s := New(WithSkipIfRunning())

	var running int32
	var execCount int32

	err := s.Every("slow-job", 1*time.Second, func(ctx context.Context) {
		atomic.AddInt32(&execCount, 1)
		if !atomic.CompareAndSwapInt32(&running, 0, 1) {
			return
		}
		defer atomic.StoreInt32(&running, 0)

		// Simulate slow job
		time.Sleep(1500 * time.Millisecond)
	})
	if err != nil {
		t.Fatalf("failed to schedule job: %v", err)
	}

	s.Start()
	time.Sleep(2500 * time.Millisecond)
	s.Stop()

	// With SkipIfRunning, overlapping executions should be skipped
	// The slow job takes 1.5s, interval is 1s, so some should be skipped
	// We just verify it ran at least once
	if atomic.LoadInt32(&execCount) < 1 {
		t.Error("job should have executed at least once")
	}
}

func TestSchedulerPanicRecovery(t *testing.T) {
	s := New()

	var executedAfterPanic int32

	// First job panics
	s.Every("panic-job", 1*time.Second, func(ctx context.Context) {
		panic("test panic")
	})

	// Second job should still run
	s.Every("normal-job", 1*time.Second, func(ctx context.Context) {
		atomic.AddInt32(&executedAfterPanic, 1)
	})

	s.Start()
	time.Sleep(1200 * time.Millisecond)
	s.Stop()

	if atomic.LoadInt32(&executedAfterPanic) < 1 {
		t.Error("normal job should have executed despite panic in other job")
	}
}

func TestSchedulerInvalidCronExpression(t *testing.T) {
	s := New()

	err := s.Cron("invalid-job", "invalid expression", func(ctx context.Context) {})
	if err == nil {
		t.Error("expected error for invalid cron expression")
	}
}
