package backoff

import (
	"time"
)

type BackOff interface {
	Reset()
	GetBackOffInterval() time.Duration
	BackOff()
}

type exponentialBackOff struct {
	initialInterval time.Duration
	maxInterval     time.Duration
	multiplier      float64
	currentInterval time.Duration
}

func NewExponentialBackOff(
	initialInterval time.Duration,
	maxInterval time.Duration,
	multiplier float64,
) BackOff {
	e := &exponentialBackOff{
		initialInterval: initialInterval,
		maxInterval:     maxInterval,
		multiplier:      multiplier,
	}
	e.Reset()
	return e
}

func (e *exponentialBackOff) Reset() {
	e.currentInterval = e.initialInterval
}

func (e *exponentialBackOff) GetBackOffInterval() time.Duration {
	return e.currentInterval
}

func (e *exponentialBackOff) BackOff() {
	var nextInterval time.Duration
	if e.currentInterval == e.maxInterval {
		nextInterval = e.maxInterval
	} else if float64(e.currentInterval) >= float64(e.maxInterval)/e.multiplier {
		nextInterval = e.maxInterval
	} else {
		nextInterval = time.Duration(float64(e.currentInterval) * e.multiplier)
	}
	e.currentInterval = nextInterval
}
