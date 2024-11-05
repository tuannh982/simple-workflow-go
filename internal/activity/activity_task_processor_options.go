package activity

import "time"

type ActivityTaskProcessorOptions struct {
	InitialBackoffInterval time.Duration
	MaxBackoffInterval     time.Duration
	BackoffMultiplier      float64
}

func NewActivityTaskProcessorOptions() *ActivityTaskProcessorOptions {
	return &ActivityTaskProcessorOptions{
		InitialBackoffInterval: 15 * time.Second,
		MaxBackoffInterval:     5 * time.Minute,
		BackoffMultiplier:      1.2,
	}
}

func WithInitialBackoffInterval(duration time.Duration) func(*ActivityTaskProcessorOptions) {
	return func(options *ActivityTaskProcessorOptions) {
		options.InitialBackoffInterval = duration
	}
}

func WithMaxBackoffInterval(duration time.Duration) func(*ActivityTaskProcessorOptions) {
	return func(options *ActivityTaskProcessorOptions) {
		options.MaxBackoffInterval = duration
	}
}

func WithBackoffMultiplier(multiplier float64) func(*ActivityTaskProcessorOptions) {
	return func(options *ActivityTaskProcessorOptions) {
		options.BackoffMultiplier = multiplier
	}
}
