package workerpool

type Option func(p *Pool)

func WithErrorPropagationDisabled() Option {
	return func(p *Pool) {
		p.disableErrorPropagation = true
	}
}
