package workerpool

import (
	"context"
	"golang.org/x/sync/errgroup"
	"sync/atomic"
)

type Job func(ctx context.Context) error

type Pool struct {
	workers                 int
	disableErrorPropagation bool

	jobsDone       atomic.Int64
	jobsWithErrors atomic.Int64
	jobsRecovered  atomic.Int64
}

func New(workers int, opts ...Option) *Pool {
	p := &Pool{
		workers: workers,
	}

	if p.workers < 1 {
		p.workers = 1
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

func (p *Pool) Run(ctx context.Context, jobs <-chan Job) error {
	g, groupCtx := errgroup.WithContext(ctx)

	for i := 0; i < p.workers; i++ {
		g.Go(func() error {
			return p.worker(groupCtx, jobs)
		})
	}

	return g.Wait()
}

func (p *Pool) worker(groupCtx context.Context, jobs <-chan Job) error {
	for job := range jobs {
		if err := p.runJob(groupCtx, job); err != nil {
			p.jobsWithErrors.Add(1)

			if p.disableErrorPropagation {
				continue
			}

			return err
		}
	}

	return nil
}

func (p *Pool) runJob(ctx context.Context, job Job) (err error) {
	defer func() {
		p.jobsDone.Store(1)

		if r := recover(); r != nil {
			p.jobsRecovered.Add(1)
			err = r.(error)
		}
	}()

	if job == nil {
		return nil
	}

	return job(ctx)
}
