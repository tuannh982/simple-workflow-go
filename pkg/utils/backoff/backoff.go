package backoff

import (
	"time"
)

type BackOff interface {
	Reset()
	GetBackOffDuration() time.Duration
	BackOff()
}

type exponentialBackOff struct {
	initialDuration time.Duration
	maxDuration     time.Duration
	multiplier      float64
	currentDuration time.Duration
}

func NewExponentialBackOff(
	initialInterval time.Duration,
	maxInterval time.Duration,
	multiplier float64,
) BackOff {
	e := &exponentialBackOff{
		initialDuration: initialInterval,
		maxDuration:     maxInterval,
		multiplier:      multiplier,
	}
	e.Reset()
	return e
}

func (e *exponentialBackOff) Reset() {
	e.currentDuration = e.initialDuration
}

func (e *exponentialBackOff) GetBackOffDuration() time.Duration {
	return e.currentDuration
}

func (e *exponentialBackOff) BackOff() {
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

func CalculateNextBackOffDuration(
	initialDuration time.Duration,
	maxDuration time.Duration,
	multiplier float64,
	currentDuration time.Duration,
) time.Duration {
	bo := &exponentialBackOff{
		initialDuration: initialDuration,
		maxDuration:     maxDuration,
		multiplier:      multiplier,
		currentDuration: currentDuration,
	}
	bo.BackOff()
	return bo.GetBackOffDuration()
}
