package workerpool

import (
	"context"
	"fmt"
	"testing"
)

var table = []struct {
	workers int
	jobs    int
}{
	{1, 100},
	{2, 100},
	{4, 100},
	{1, 1000},
	{2, 1000},
	{4, 1000},
	{1, 100000},
	{2, 100000},
	{4, 100000},
}

func BenchmarkWorkerPool(b *testing.B) {
	for _, row := range table {
		b.Run(
			fmt.Sprintf("BenchmarkWorkerPool_%dw_%dj", row.workers, row.jobs),
			func(b *testing.B) {
				benchmarkPool(b, row.workers, row.jobs)
			})
	}
}

func benchmarkPool(b *testing.B, workers int, numberOfJobs int) {
	pool := New(workers)

	job := func(ctx context.Context) error {
		_ = ctx
		return nil
	}

	jobs := make(chan Job)
	done := make(chan interface{})
	go func() {
		if err := pool.Run(context.Background(), jobs); err != nil {
			b.Errorf("pool: %v", err)
		}
		done <- struct{}{}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < numberOfJobs; j++ {
			jobs <- job
		}
	}
	close(jobs)

	<-done
}
