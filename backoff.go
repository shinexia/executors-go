package executors

import (
	"math"
	"math/rand/v2"
	"time"
)

// Backoff is a function type for calculating retry durations
// It takes the number of attempts as input and returns the duration to wait before the next attempt
type Backoff func(attempts int) time.Duration

// DefaultBackoff is an exponential backoff strategy with a factor of 1.5 and a maximum interval of 30 seconds
var DefaultBackoff = ExponentialBackoff(1.5, 30*time.Second)

// ExponentialBackoff increases retry duration exponentially
func ExponentialBackoff(factor float64, maxInterval time.Duration) Backoff {
	return func(attempts int) time.Duration {
		if attempts == 0 {
			return time.Duration(0)
		}
		// Use math.Pow to calculate the exponential backoff duration
		d := time.Duration(math.Pow(factor, float64(attempts))) * time.Second
		// Ensure the backoff duration does not exceed the maximum interval
		if maxInterval > 0 && d > maxInterval {
			d = maxInterval
		}
		// Introduce randomness to avoid thundering herd problem
		d = time.Duration(float64(d) * (0.8 + 0.4*rand.Float64()))
		return d
	}
}

// FixedIntervalBackoff returns a fixed retry duration, primarily used for testing
// The interval should be a non-negative duration
func FixedIntervalBackoff(interval time.Duration) Backoff {
	// Check if the interval is negative
	if interval < 0 {
		panic("interval must be a non-negative duration")
	}
	return func(attempts int) time.Duration {
		// Always return the fixed interval regardless of the number of attempts
		return interval
	}
}
