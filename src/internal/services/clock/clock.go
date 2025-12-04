package clock

import "time"

// Clock abstracts time for testability
type Clock interface {
	Now() time.Time
}

// SystemClock is the production implementation of Clock
type SystemClock struct{}

func (c *SystemClock) Now() time.Time {
	return time.Now()
}

// New creates a new Clock instance
func New() Clock {
	return &SystemClock{}
}
