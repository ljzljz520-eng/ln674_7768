package service

import "time"

type Clock interface {
	Now() time.Time
}

type FixedClock struct {
	current time.Time
}

func NewFixedClock(current time.Time) *FixedClock {
	return &FixedClock{current: current}
}

func (c *FixedClock) Now() time.Time {
	return c.current
}

func (c *FixedClock) Advance(delta time.Duration) {
	c.current = c.current.Add(delta)
}

type SequenceClock struct {
	values []time.Time
	index  int
}

func NewSequenceClock(values ...time.Time) *SequenceClock {
	return &SequenceClock{values: append([]time.Time(nil), values...)}
}

func (c *SequenceClock) Now() time.Time {
	if len(c.values) == 0 {
		return time.Unix(0, 0).UTC()
	}
	value := c.values[c.index]
	if c.index < len(c.values)-1 {
		c.index++
	}
	return value
}
