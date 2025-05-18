package horizon

import (
	"context"

	"github.com/robfig/cron/v3"
)

// Scheduler defines the interface for scheduling and managing cron jobs
type Scheduler interface {
	// Start initializes the scheduler and in-memory job store
	Start(ctx context.Context) error

	// Stop gracefully shuts down the scheduler and clears all timers
	Stop(ctx context.Context) error

	// CreateJob registers a new cron-style job with the specified schedule
	CreateJob(ctx context.Context, jobID string, schedule cron.Schedule, task func() error) error

	// RunJobNow immediately executes a scheduled job by ID
	RunJobNow(ctx context.Context, jobID string) error

	// RemoveJob deletes a job by its ID from the scheduler
	RemoveJob(ctx context.Context, jobID string) error

	// ListJobs returns all registered job IDs
	ListJobs(ctx context.Context) ([]string, error)
}
