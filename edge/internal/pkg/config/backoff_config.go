package config

import (
	"time"

	"github.com/cenkalti/backoff/v4"
)

type BackoffConfig struct {
	// InitialInterval is the initial RetryInterval
	InitialInterval time.Duration

	// RandomizationFactor is used to randomize the retry intervals.
	// Each interval is calculated as follows:
	// randomized interval = CurrentRetryInterval *
	// (random value in range [1 - RandomizationFactor, 1 + RandomizationFactor])
	RandomizationFactor float64

	// Multiplier is the interval growth factor.
	// Every time a new interval is requested, the previous interval
	// will be multiplied by Multiplier
	Multiplier float64

	// MaxInterval is the maximum retry interval.
	// Note: MaxInterval caps the RetryInterval and not the randomized interval.
	MaxInterval time.Duration
}

func (c *BackoffConfig) NewBackoff() *backoff.ExponentialBackOff {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = c.InitialInterval
	b.RandomizationFactor = c.RandomizationFactor
	b.Multiplier = c.Multiplier
	b.MaxInterval = c.MaxInterval
	b.Reset()

	return b
}

func DefaultBackoffConfig() BackoffConfig {
	return BackoffConfig{
		InitialInterval:     backoff.DefaultInitialInterval,
		RandomizationFactor: backoff.DefaultRandomizationFactor,
		Multiplier:          backoff.DefaultMultiplier,
		MaxInterval:         backoff.DefaultMaxInterval,
	}
}
