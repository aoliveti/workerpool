package workerpool

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGroup_Append(t *testing.T) {
	t.Run("append function", func(t *testing.T) {
		g := Group{}

		g.Append(func(ctx context.Context) error {
			_ = ctx
			return nil
		})

		assert.NotZero(t, len(g))
	})
	t.Run("append nil function", func(t *testing.T) {
		g := Group{}

		g.Append(nil)

		assert.NotZero(t, len(g))
	})
}

func TestGroup_toStream(t *testing.T) {
	t.Run("empty group", func(t *testing.T) {
		g := Group{}

		counter := 0
		done := make(chan interface{})
		jobs := g.toStream(context.TODO())
		go func() {
			for range jobs {
				counter++
			}
			done <- struct{}{}
		}()
		<-done

		assert.Zero(t, counter)
	})
	t.Run("not empty group", func(t *testing.T) {
		g := Group{}
		expectedCount := 100
		for i := 0; i < expectedCount; i++ {
			g.Append(func(ctx context.Context) error {
				_ = ctx
				return nil
			})
		}

		counter := 0
		done := make(chan interface{})
		jobs := g.toStream(context.TODO())
		go func() {
			for range jobs {
				counter++
			}
			done <- struct{}{}
		}()
		<-done

		assert.Equal(t, expectedCount, counter)
	})
}
