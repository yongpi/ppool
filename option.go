package ppool

import "time"

type Option func(options *Options)

func WithMaxBlockingNum(n int32) Option {
	return func(options *Options) {
		options.MaxBlockingNum = n
	}
}

func WithWorkerCleanTime(t time.Duration) Option {
	return func(options *Options) {
		options.WorkerCleanDuration = t
	}
}

func WithPanicHandler(handler func(interface{})) Option {
	return func(options *Options) {
		options.PanicHandler = handler
	}
}

func WithLogger(logger Logger) Option {
	return func(options *Options) {
		options.Logger = logger
	}
}
