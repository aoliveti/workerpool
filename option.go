package workerpool

// Option represents an option function that can be used to configure a Pool.
type Option func(p *Pool)

// WithErrorPropagationDisabled returns an option function that disables error propagation in the pool.
func WithErrorPropagationDisabled() Option {
	return func(p *Pool) {
		p.disableErrorPropagation = true
	}
}
