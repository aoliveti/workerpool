package workerpool

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWithErrorPropagationDisabled(t *testing.T) {
	t.Run("disable propagation error option", func(t *testing.T) {
		p := New(1, WithErrorPropagationDisabled())
		assert.Equal(t, true, p.disableErrorPropagation)
	})
}
