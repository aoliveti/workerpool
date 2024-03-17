package workerpool

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("negative workers", func(t *testing.T) {
		workers := -1

		p := New(workers)
		assert.Equal(t, 1, p.workers)
	})
	t.Run("zero workers", func(t *testing.T) {
		workers := 0

		p := New(workers)
		assert.Equal(t, 1, p.workers)
	})
	t.Run("positive workers", func(t *testing.T) {
		workers := 2

		p := New(workers)
		assert.Equal(t, 2, p.workers)
	})
}

func TestPool_Run(t *testing.T) {
	const workers int = 2

	t.Parallel()

	t.Run("run empty", func(t *testing.T) {
		p := New(workers)

		jobs := make(chan Job)
		close(jobs)

		err := p.Run(context.TODO(), jobs)
		assert.NoError(t, err)

		assert.Zero(t, p.jobsDone.Load())
	})
	t.Run("run one job", func(t *testing.T) {
		p := New(workers)

		jobs := make(chan Job)
		go func() {
			defer close(jobs)
			jobs <- func(ctx context.Context) error {
				_ = ctx
				return nil
			}
		}()

		err := p.Run(context.TODO(), jobs)
		assert.NoError(t, err)
		assert.Equal(t, 1, int(p.jobsDone.Load()))
	})
	t.Run("run nil job", func(t *testing.T) {
		p := New(workers)

		jobs := make(chan Job)
		go func() {
			defer close(jobs)
			jobs <- nil
		}()

		err := p.Run(context.TODO(), jobs)
		assert.NoError(t, err)
		assert.Equal(t, 1, int(p.jobsDone.Load()))
	})
	t.Run("run one hundred jobs", func(t *testing.T) {
		p := New(workers)

		jobs := make(chan Job)
		go func() {
			defer close(jobs)

			for i := 0; i < 100; i++ {
				jobs <- func(ctx context.Context) error {
					_ = ctx
					return nil
				}
			}
		}()

		err := p.Run(context.TODO(), jobs)
		assert.NoError(t, err)
	})
	t.Run("propagate error", func(t *testing.T) {
		p := New(workers)
		expectedErr := errors.New("error")

		jobs := make(chan Job)
		go func() {
			defer close(jobs)

			for i := 0; i < 100; i++ {
				v := i
				jobs <- func(ctx context.Context) error {
					if v == 4 {
						return expectedErr
					}

					select {
					case <-time.After(250 * time.Millisecond):
					case <-ctx.Done():
						return ctx.Err()
					}
					return nil
				}
			}
		}()

		err := p.Run(context.TODO(), jobs)
		assert.Error(t, err, expectedErr)
		got := int(p.jobsWithErrors.Load())
		assert.NotZero(t, got)
	})
	t.Run("recover from panic", func(t *testing.T) {
		p := New(workers)
		expectedErr := errors.New("error")

		jobs := make(chan Job)
		go func() {
			defer close(jobs)

			for i := 0; i < 100; i++ {
				v := i
				jobs <- func(ctx context.Context) error {
					if v == 4 {
						panic(expectedErr)
					}

					select {
					case <-time.After(250 * time.Millisecond):
					case <-ctx.Done():
						return ctx.Err()
					}
					return nil
				}
			}
		}()

		err := p.Run(context.TODO(), jobs)
		assert.Error(t, err, expectedErr)
		je := int(p.jobsWithErrors.Load())
		assert.NotZero(t, je)
		jr := int(p.jobsRecovered.Load())
		assert.Equal(t, 1, jr)
	})
	t.Run("do not propagate error", func(t *testing.T) {
		p := New(workers, WithErrorPropagationDisabled())
		expectedErr := errors.New("error")

		jobs := make(chan Job)
		go func() {
			defer close(jobs)

			for i := 0; i < 100; i++ {
				v := i
				jobs <- func(ctx context.Context) error {
					if v == 4 {
						return expectedErr
					}

					return nil
				}
			}
		}()

		err := p.Run(context.TODO(), jobs)
		assert.NoError(t, err)
		got := int(p.jobsWithErrors.Load())
		assert.Equal(t, 1, got)
	})
	t.Run("cancel before end", func(t *testing.T) {
		p := New(workers)

		jobs := make(chan Job)
		go func() {
			defer close(jobs)

			for i := 0; i < 100; i++ {
				jobs <- func(ctx context.Context) error {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(500 * time.Millisecond):
						return nil
					}
				}
			}
		}()

		ctx, cancel := context.WithTimeout(context.TODO(), 50*time.Millisecond)
		defer cancel()

		err := p.Run(ctx, jobs)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})
}

func TestPool_worker(t *testing.T) {
	t.Parallel()

	t.Run("no jobs", func(t *testing.T) {
		p := Pool{}

		jobs := make(chan Job)
		close(jobs)

		err := p.worker(context.TODO(), jobs)
		assert.NoError(t, err)
	})
	t.Run("jobs without error", func(t *testing.T) {
		p := Pool{}

		jobs := make(chan Job)
		go func() {
			defer close(jobs)

			for i := 0; i < 10; i++ {
				jobs <- nil
			}
		}()

		err := p.worker(context.TODO(), jobs)
		assert.NoError(t, err)
		assert.Zero(t, p.jobsWithErrors.Load())
	})
	t.Run("jobs with error", func(t *testing.T) {
		expectedErr := errors.New("error")
		p := Pool{}

		jobs := make(chan Job)
		go func() {
			defer close(jobs)

			for i := 0; i < 10; i++ {
				if i == 2 {
					jobs <- func(ctx context.Context) error {
						_ = ctx
						return expectedErr
					}
				}
				jobs <- nil
			}
		}()

		err := p.worker(context.TODO(), jobs)
		assert.ErrorIs(t, err, expectedErr)
		assert.NotZero(t, p.jobsWithErrors.Load())
	})
}

func TestPool_runJob(t *testing.T) {
	t.Parallel()

	t.Run("run nil job", func(t *testing.T) {
		var job Job
		pool := Pool{}
		err := pool.runJob(context.TODO(), job)
		assert.NoError(t, err)
	})
	t.Run("run job", func(t *testing.T) {
		job := Job(func(ctx context.Context) error {
			_ = ctx
			return nil
		})
		pool := Pool{}
		err := pool.runJob(context.TODO(), job)
		assert.NoError(t, err)
	})
	t.Run("run job with error", func(t *testing.T) {
		expectedErr := errors.New("error")
		job := Job(func(ctx context.Context) error {
			_ = ctx
			return expectedErr
		})
		pool := Pool{}
		err := pool.runJob(context.TODO(), job)
		assert.ErrorIs(t, err, expectedErr)
	})
	t.Run("run job with panic", func(t *testing.T) {
		expectedErr := errors.New("error")
		job := Job(func(ctx context.Context) error {
			_ = ctx
			panic(expectedErr)
		})
		pool := Pool{}
		err := pool.runJob(context.TODO(), job)
		assert.ErrorIs(t, err, expectedErr)
	})
	t.Run("run job with canceled context", func(t *testing.T) {
		job := Job(func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				return nil
			}
		})
		pool := Pool{}

		ctx, cancel := context.WithCancel(context.TODO())
		cancel()
		err := pool.runJob(ctx, job)
		assert.ErrorIs(t, err, context.Canceled)
	})
}
