# workerpool

This package offers a simple and minimalist solution for managing concurrent job execution via a worker pool. It implements a concurrency-limiting goroutine pool, effectively constraining the concurrency of job execution.
## Installation

```bash
go get github.com/aoliveti/workerpool
```

## Usage

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aoliveti/workerpool"
)

func main() {
	pool := workerpool.New(2)

	jobs := make(chan workerpool.Job)
	go func() {
		defer close(jobs)

		for i := 0; i < 10; i++ {
			num := i
			// Define a job function
			job := func(ctx context.Context) error {
				_ = ctx
				fmt.Printf("Job number: %d done!\n", num)
				return nil
			}

			jobs <- job
		}
	}()

	if err := pool.Run(context.Background(), jobs); err != nil {
		log.Fatal("pool:", err)
	}
}
```

In case of an error or panic occurring within the job, the default pool halts execution and propagates the error. To disable this functionality, you need to pass the WithErrorPropagationDisabled() option to the constructor:
```go
pool := workerpool.New(2, workerpool.WithErrorPropagationDisabled())
````

## Benchmark
```shell
goos: darwin
goarch: amd64
pkg: github.com/aoliveti/workerpool
cpu: Intel(R) Core(TM) i5-7360U CPU @ 2.30GHz
BenchmarkWorkerPool/BenchmarkWorkerPool_1w_100j-4                  45645             23979 ns/op               0 B/op          0 allocs/op
BenchmarkWorkerPool/BenchmarkWorkerPool_2w_100j-4                  42039             28131 ns/op               0 B/op          0 allocs/op
BenchmarkWorkerPool/BenchmarkWorkerPool_4w_100j-4                  39540             30274 ns/op               0 B/op          0 allocs/op
BenchmarkWorkerPool/BenchmarkWorkerPool_1w_1000j-4                  4921            246290 ns/op               0 B/op          0 allocs/op
BenchmarkWorkerPool/BenchmarkWorkerPool_2w_1000j-4                  4364            288214 ns/op               0 B/op          0 allocs/op
BenchmarkWorkerPool/BenchmarkWorkerPool_4w_1000j-4                  3973            295878 ns/op               0 B/op          0 allocs/op
BenchmarkWorkerPool/BenchmarkWorkerPool_1w_100000j-4                  49          24046569 ns/op               4 B/op          0 allocs/op
BenchmarkWorkerPool/BenchmarkWorkerPool_2w_100000j-4                  44          28404396 ns/op               6 B/op          0 allocs/op
BenchmarkWorkerPool/BenchmarkWorkerPool_4w_100000j-4                  40          29593806 ns/op              49 B/op          0 allocs/op
PASS
ok      github.com/aoliveti/workerpool  14.701s
```

## License

The library is released under the MIT license. See [LICENSE](LICENSE) file.