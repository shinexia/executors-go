package executors

import "time"

type Option func(ro *runOption)

type runOption struct {
	retryCount    int
	expiration    time.Duration
	backoff       Backoff
	cleanup       bool
	callback      []TaskCallback
	retryCallback []RetryCallback
	skipPrev      bool
	runOnce       bool
}

func newRunOption(optsList ...[]Option) runOption {
	ro := runOption{}
	for _, opts := range optsList {
		for _, opt := range opts {
			opt(&ro)
		}
	}
	return ro
}

// WithRetryCount sets the maximum number of retries for a task.
// The retry parameter specifies the maximum number of retry attempts.
func WithRetryCount(retry int) Option {
	return func(ro *runOption) {
		ro.retryCount = retry
	}
}

func WithExpiration(expiration time.Duration) Option {
	return func(ro *runOption) {
		ro.expiration = expiration
	}
}

func WithBackoff(fn Backoff) Option {
	return func(ro *runOption) {
		ro.backoff = fn
	}
}

// WithCleanup sets the cleanup option to the provided boolean value.
// When cleanup is true, the task will be marked as FastFail and will wait for all tasks to complete.
func WithCleanup(cleanup bool) Option {
	return func(ro *runOption) {
		ro.cleanup = cleanup
	}
}

// WithCallback adds a TaskCallback to the runOption's callback list.
// If the provided callback is nil, it will not be added.
func WithCallback(cb TaskCallback) Option {
	return func(ro *runOption) {
		if cb != nil {
			ro.callback = append(ro.callback, cb)
		}
	}
}

func WithRetryCallback(cb RetryCallback) Option {
	return func(ro *runOption) {
		if cb != nil {
			ro.retryCallback = append(ro.retryCallback, cb)
		}
	}
}

// WithSkipPrev sets the skipPrev option to the provided boolean value.
// When skipPrev is true, the task will skip previously blocked tasks.
func WithSkipPrev(skip bool) Option {
	return func(ro *runOption) {
		ro.skipPrev = skip
	}
}

// WithRunOnce sets the runOnce option to the provided boolean value.
// When runOnce is true, the task will be attempted once before creating a new goroutine.
// This is recommended when all tasks are serial or very lightweight.
func WithRunOnce(b bool) Option {
	return func(ro *runOption) {
		ro.runOnce = b
	}
}
