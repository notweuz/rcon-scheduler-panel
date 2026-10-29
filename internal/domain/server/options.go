package server

type options struct {
	IncludeProtocol bool
}

type Option func(*options)

func WithPreloadProtocol() Option {
	return func(o *options) {
		o.IncludeProtocol = true
	}
}