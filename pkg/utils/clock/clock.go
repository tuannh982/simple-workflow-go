package clock

import "time"

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func NewRealClock() Clock {
	return &realClock{}
}

func (c *realClock) Now() time.Time {
	return time.Now()
}

type mockClock struct {
	currentTime  time.Time
	factor       float64   // Acceleration factor: 1.0 is normal speed, 2.0 is double speed, etc.
	lastRealTime time.Time // Tracks the last real time we checked
}

func NewMockClock() Clock {
	now := time.Now()
	return &mockClock{
		currentTime:  now,
		factor:       1.0,
		lastRealTime: now,
	}
}

// Now returns the current mock time, advancing it based on real elapsed time if needed.
func (c *mockClock) Now() time.Time {
	realNow := time.Now()
	elapsed := realNow.Sub(c.lastRealTime)

	// Apply the factor to the elapsed time
	adjustedElapsed := time.Duration(float64(elapsed) * c.factor)

	// Update the current time and last real time
	c.currentTime = c.currentTime.Add(adjustedElapsed)
	c.lastRealTime = realNow

	return c.currentTime
}

// SetTime sets the mock clock to a specific time.
func (c *mockClock) SetTime(t time.Time) {
	c.currentTime = t
	c.lastRealTime = time.Now()
}

// Advance moves the mock clock forward by the specified duration.
func (c *mockClock) Advance(d time.Duration) {
	c.currentTime = c.currentTime.Add(d)
	c.lastRealTime = time.Now()
}

// SetFactor sets the acceleration factor for the mock clock.
// A factor of 1.0 means normal speed, 2.0 means double speed, etc.
func (c *mockClock) SetFactor(factor float64) {
	if factor <= 0 {
		factor = 1.0
	}
	// Update the current time before changing the factor
	c.Now()
	c.factor = factor
}

// AdvanceWithFactor advances time by the specified duration, adjusted by the acceleration factor.
func (c *mockClock) AdvanceWithFactor(d time.Duration) {
	adjustedDuration := time.Duration(float64(d) * c.factor)
	c.currentTime = c.currentTime.Add(adjustedDuration)
	c.lastRealTime = time.Now()
}

// Reset resets the mock clock to normal speed and the current time.
func (c *mockClock) Reset() {
	c.currentTime = time.Now()
	c.lastRealTime = c.currentTime
	c.factor = 1.0
}
