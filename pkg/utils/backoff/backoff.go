package backoff

import (
	"time"
)

type Backoff interface {
	Reset()
	GetBackoffDuration() time.Duration
	Backoff()
}

type exponentialBackoff struct {
	initialDuration time.Duration
	maxDuration     time.Duration
	multiplier      float64
	currentDuration time.Duration
}

func NewExponentialBackoff(
	initialInterval time.Duration,
	maxInterval time.Duration,
	multiplier float64,
) Backoff {
	e := &exponentialBackoff{
		initialDuration: initialInterval,
		maxDuration:     maxInterval,
		multiplier:      multiplier,
	}
	e.Reset()
	return e
}

func (e *exponentialBackoff) Reset() {
	e.currentDuration = e.initialDuration
}

func (e *exponentialBackoff) GetBackoffDuration() time.Duration {
	return e.currentDuration
}

func (e *exponentialBackoff) Backoff() {
	var nextInterval time.Duration
	if e.currentDuration >= e.maxDuration {
		nextInterval = e.maxDuration
	} else if float64(e.currentDuration) >= float64(e.maxDuration)/e.multiplier {
		nextInterval = e.maxDuration
	} else {
		nextInterval = time.Duration(float64(e.currentDuration) * e.multiplier)
	}
	e.currentDuration = nextInterval
}

func CalculateNextBackoffDuration(
	initialDuration time.Duration,
	maxDuration time.Duration,
	multiplier float64,
	currentDuration time.Duration,
) time.Duration {
	bo := &exponentialBackoff{
		initialDuration: initialDuration,
		maxDuration:     maxDuration,
		multiplier:      multiplier,
		currentDuration: currentDuration,
	}
	bo.Backoff()
	return bo.GetBackoffDuration()
}
