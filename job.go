package workerpool

import (
	"context"
)

type Job func(ctx context.Context) error

type Group []Job

func (g *Group) Append(j Job) {
	*g = append(*g, j)
}

func (g *Group) toStream(ctx context.Context) <-chan Job {
	fetchedJobs := make(chan Job)
	go func() {
		defer close(fetchedJobs)

		for _, job := range *g {
			select {
			case <-ctx.Done():
				return
			case fetchedJobs <- job:
			}
		}
	}()

	return fetchedJobs
}
