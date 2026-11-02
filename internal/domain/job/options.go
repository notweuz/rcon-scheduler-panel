package job

type options struct {
	IncludeServer bool
}

type Option func(*options)

func WithPreloadServer() Option {
	return func(o *options) {
		o.IncludeServer = true
	}
}
