// Package workerpool implements a worker pool for concurrent job execution.
package workerpool

import (
	"context"
	"sync/atomic"

	"golang.org/x/sync/errgroup"
)

// Job represents a function that can be executed as a job in the worker pool.
type Job func(ctx context.Context) error

// Pool represents a worker pool for concurrent job execution.
type Pool struct {
	// Number of worker goroutines in the pool
	workers int

	// Flag to disable error propagation
	disableErrorPropagation bool

	// Counter for successfully completed jobs
	jobsDone atomic.Int64

	// Counter for jobs that completed with errors
	jobsWithErrors atomic.Int64

	// Counter for jobs that caused panics and were recovered
	jobsRecovered atomic.Int64
}

// New creates a new worker pool with the specified number of workers and options.
func New(workers int, opts ...Option) *Pool {
	p := &Pool{
		workers: workers,
	}

	// Ensure there is at least one worker in the pools
	if p.workers < 1 {
		p.workers = 1
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Run starts the worker pool with the provided context and job channel.
func (p *Pool) Run(ctx context.Context, jobs <-chan Job) error {
	g, groupCtx := errgroup.WithContext(ctx)

	for i := 0; i < p.workers; i++ {
		g.Go(func() error {
			return p.worker(groupCtx, jobs)
		})
	}

	return g.Wait()
}

// worker processes jobs from the job channel.
func (p *Pool) worker(groupCtx context.Context, jobs <-chan Job) error {
	for job := range jobs {
		// Execute the job and handle errors
		if err := p.runJob(groupCtx, job); err != nil {
			// Check if error propagation is disabled
			if p.disableErrorPropagation {
				// Skip to the next job
				continue
			}

			// Propagate the error
			return err
		}
	}

	return nil
}

// runJob executes a single job and tracks job completion and errors.
func (p *Pool) runJob(ctx context.Context, job Job) (err error) {
	defer func() {
		p.jobsDone.Store(1)

		// Recover from panic and increment jobsRecovered counter
		if r := recover(); r != nil {
			p.jobsRecovered.Add(1)
			err = r.(error)
		}

		if err != nil {
			p.jobsWithErrors.Add(1)
		}
	}()

	if job == nil {
		return nil
	}

	return job(ctx)
}
