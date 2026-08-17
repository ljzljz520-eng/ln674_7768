package retry

import (
	"time"

	"campus-device-hub/internal/domain"
)

type Policy struct {
	MaxAttempts int
	BaseDelay   time.Duration
}

func DefaultPolicy() Policy {
	return Policy{MaxAttempts: 3, BaseDelay: time.Minute}
}

func (p Policy) Delay(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	if p.BaseDelay <= 0 {
		return 0
	}
	return time.Duration(attempt) * p.BaseDelay
}

func (p Policy) Apply(task *domain.RetryTask, now time.Time, message string) {
	if task == nil {
		return
	}
	task.Attempt++
	task.MaxAttempts = p.MaxAttempts
	task.LastError = message
	task.NextAt = now.Add(p.Delay(task.Attempt))
	if task.Attempt >= p.MaxAttempts {
		task.Resolved = true
	}
}

func (p Policy) Exhausted(task domain.RetryTask) bool {
	return task.Resolved || task.Attempt >= p.MaxAttempts
}
