package domain

import "time"

type RetryTask struct {
	ID          string    `json:"id"`
	RecordID    string    `json:"record_id"`
	Vendor      string    `json:"vendor"`
	Attempt     int       `json:"attempt"`
	MaxAttempts int       `json:"max_attempts"`
	NextAt      time.Time `json:"next_at"`
	LastError   string    `json:"last_error"`
	Resolved    bool      `json:"resolved"`
}

func (t RetryTask) IsDue(now time.Time) bool {
	return !t.Resolved && !t.NextAt.After(now)
}

func (t *RetryTask) Schedule(now time.Time, message string) {
	t.Attempt++
	t.LastError = message
	t.NextAt = now.Add(time.Duration(t.Attempt) * time.Minute)
	if t.Attempt >= t.MaxAttempts {
		t.Resolved = true
	}
}

func (t *RetryTask) Resolve() {
	t.Resolved = true
	t.LastError = ""
}
