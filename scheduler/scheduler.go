// Package scheduler provides a simple job scheduling abstraction on top of robfig/cron.
// It offers a clean API for scheduling periodic tasks with built-in panic recovery
// and optional overlap prevention.
package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// Job represents a scheduled job with its metadata.
type Job struct {
	Name     string
	Schedule string
	EntryID  cron.EntryID
}

// Option configures the Scheduler.
type Option func(*Scheduler)

// WithBaseContext sets the base context used for all scheduled jobs.
// A cancelable child context is created on Start() and canceled on Stop().
func WithBaseContext(ctx context.Context) Option {
	return func(s *Scheduler) {
		if ctx == nil {
			return
		}
		s.baseCtx = ctx
	}
}

// WithLogger sets a custom logger for the scheduler.
func WithLogger(logger *slog.Logger) Option {
	return func(s *Scheduler) {
		s.logger = logger
	}
}

// WithLocation sets the timezone for schedule parsing.
func WithLocation(loc *time.Location) Option {
	return func(s *Scheduler) {
		s.location = loc
	}
}

// WithSkipIfRunning prevents job overlap - skips execution if previous run is still active.
func WithSkipIfRunning() Option {
	return func(s *Scheduler) {
		s.skipIfRunning = true
	}
}

// Scheduler manages scheduled jobs using cron expressions or fixed intervals.
type Scheduler struct {
	cron          *cron.Cron
	logger        *slog.Logger
	location      *time.Location
	skipIfRunning bool
	jobs          map[string]Job
	mu            sync.RWMutex
	started       bool
	baseCtx       context.Context
	runCtx        context.Context
	runCancel     context.CancelFunc
}

// New creates a new Scheduler with the given options.
func New(opts ...Option) *Scheduler {
	s := &Scheduler{
		logger:   slog.Default(),
		location: time.UTC,
		baseCtx:  context.Background(),
		jobs:     make(map[string]Job),
	}

	for _, opt := range opts {
		opt(s)
	}

	// Build cron options
	cronOpts := []cron.Option{
		cron.WithLocation(s.location),
		cron.WithLogger(&cronLogAdapter{logger: s.logger}),
	}

	// Build chain with panic recovery and optional skip-if-running
	var chain []cron.JobWrapper
	chain = append(chain, cron.Recover(&cronLogAdapter{logger: s.logger}))
	if s.skipIfRunning {
		chain = append(chain, cron.SkipIfStillRunning(&cronLogAdapter{logger: s.logger}))
	}
	cronOpts = append(cronOpts, cron.WithChain(chain...))

	s.cron = cron.New(cronOpts...)
	return s
}

// Every schedules a job to run at fixed intervals.
// The interval string should be a duration like "5m", "1h", "30s".
func (s *Scheduler) Every(name string, interval time.Duration, fn func(ctx context.Context)) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Wrap the function to include context
	wrappedFn := func() {
		ctx := s.jobContext()
		fn(ctx)
	}

	entryID, err := s.cron.AddFunc("@every "+interval.String(), wrappedFn)
	if err != nil {
		return err
	}

	s.jobs[name] = Job{
		Name:     name,
		Schedule: "@every " + interval.String(),
		EntryID:  entryID,
	}

	s.logger.Debug("job scheduled", "name", name, "schedule", "@every "+interval.String())
	return nil
}

// Cron schedules a job using a cron expression.
// The expression uses standard 5-field format: minute hour day-of-month month day-of-week
// Examples: "0 * * * *" (every hour), "0 0 * * *" (daily at midnight)
func (s *Scheduler) Cron(name string, expr string, fn func(ctx context.Context)) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	wrappedFn := func() {
		ctx := s.jobContext()
		fn(ctx)
	}

	entryID, err := s.cron.AddFunc(expr, wrappedFn)
	if err != nil {
		return err
	}

	s.jobs[name] = Job{
		Name:     name,
		Schedule: expr,
		EntryID:  entryID,
	}

	s.logger.Debug("job scheduled", "name", name, "schedule", expr)
	return nil
}

// Remove removes a scheduled job by name.
func (s *Scheduler) Remove(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.jobs[name]
	if !ok {
		return false
	}

	s.cron.Remove(job.EntryID)
	delete(s.jobs, name)
	s.logger.Debug("job removed", "name", name)
	return true
}

// Jobs returns a list of all scheduled jobs.
func (s *Scheduler) Jobs() []Job {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		result = append(result, job)
	}
	return result
}

// Start begins executing scheduled jobs.
func (s *Scheduler) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return
	}

	s.runCtx, s.runCancel = context.WithCancel(s.baseContext())
	s.cron.Start()
	s.started = true
	s.logger.Info("scheduler started", "jobs", len(s.jobs))
}

// Stop stops the scheduler and waits for running jobs to complete.
func (s *Scheduler) Stop() context.Context {
	s.mu.Lock()
	if !s.started {
		s.mu.Unlock()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		return ctx
	}

	s.started = false
	cancel := s.runCancel
	s.runCancel = nil
	s.mu.Unlock()

	s.logger.Info("scheduler stopping")
	if cancel != nil {
		cancel()
	}
	return s.cron.Stop()
}

// Running returns true if the scheduler is running.
func (s *Scheduler) Running() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.started
}

func (s *Scheduler) baseContext() context.Context {
	if s.baseCtx != nil {
		return s.baseCtx
	}
	return context.Background()
}

func (s *Scheduler) jobContext() context.Context {
	s.mu.RLock()
	ctx := s.runCtx
	if ctx == nil {
		ctx = s.baseCtx
	}
	s.mu.RUnlock()
	if ctx != nil {
		return ctx
	}
	return context.Background()
}

// cronLogAdapter adapts slog.Logger to cron.Logger interface.
type cronLogAdapter struct {
	logger *slog.Logger
}

func (a *cronLogAdapter) Info(msg string, keysAndValues ...interface{}) {
	a.logger.Info(msg, keysAndValues...)
}

func (a *cronLogAdapter) Error(err error, msg string, keysAndValues ...interface{}) {
	args := append([]interface{}{"err", err}, keysAndValues...)
	a.logger.Error(msg, args...)
}
